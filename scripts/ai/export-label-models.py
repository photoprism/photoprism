#!/usr/bin/env python3
"""Exports checksum-recorded ImageNet classifiers for PhotoPrism."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import platform
import shlex
import sys
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

import numpy as np
import onnx
import onnxruntime
import timm
import torch
from huggingface_hub import hf_hub_download
from PIL import Image
from timm.data import resolve_model_data_config
from timm.data.transforms import str_to_interp_mode
from timm.utils import reparameterize_model
from torchvision import transforms


@dataclass(frozen=True)
class ModelSpec:
    """Describes one publisher checkpoint and its exported artifact name."""

    name: str
    checkpoint: str
    revision: str
    file_name: str
    reparameterize: bool = False

    @property
    def repo_id(self) -> str:
        """Returns the Hugging Face repository containing the checkpoint."""

        return f"timm/{self.checkpoint}"


MODEL_SPECS = {
    "efficientformerv2_s1": ModelSpec(
        name="efficientformerv2_s1",
        checkpoint="efficientformerv2_s1.snap_dist_in1k",
        revision="0c1fc60a0e89b6309d9de451cdacc11ac0a8b987",
        file_name="efficientformerv2_s1.onnx",
    ),
    "repvit_m1_0": ModelSpec(
        name="repvit_m1_0",
        checkpoint="repvit_m1_0.dist_300e_in1k",
        revision="94445f5481b027599200e61ed5e108dbaedc0139",
        file_name="repvit_m1_0.onnx",
        reparameterize=True,
    ),
    "efficientnet_b0": ModelSpec(
        name="efficientnet_b0",
        checkpoint="efficientnet_b0.ra_in1k",
        revision="1b5383e5f79cc0f7fc067e372f8f26a5fa73f26a",
        file_name="efficientnet_b0.onnx",
    ),
    "efficientformerv2_s2": ModelSpec(
        name="efficientformerv2_s2",
        checkpoint="efficientformerv2_s2.snap_dist_in1k",
        revision="1c56a76355000c79568559d34ba3fa24416b5107",
        file_name="efficientformerv2_s2.onnx",
    ),
}


class LogitsOnly(torch.nn.Module):
    """Rejects training or distilled-head outputs instead of exporting them."""

    def __init__(self, model: torch.nn.Module) -> None:
        super().__init__()
        self.model = model

    def forward(self, value: torch.Tensor) -> torch.Tensor:
        """Returns one two-dimensional tensor of final combined logits."""

        result = self.model(value)
        if not isinstance(result, torch.Tensor):
            raise RuntimeError("model did not return final combined logits")
        if result.ndim != 2 or result.shape[1] != 1000:
            raise RuntimeError(f"model returned {tuple(result.shape)}, expected [1, 1000]")
        return result


def parse_args() -> argparse.Namespace:
    """Parses the reproducible export command line."""

    repo_root = Path(__file__).resolve().parents[2]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--model",
        action="append",
        choices=[*MODEL_SPECS, "all"],
        required=True,
        help="candidate to export; repeat the option or use all",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=repo_root / "assets" / "models",
        help="model directory prefix",
    )
    parser.add_argument(
        "--fixture",
        type=Path,
        default=repo_root / "assets" / "samples" / "dog_orange.jpg",
        help="normalized fixture used for PyTorch and ONNX comparison",
    )
    parser.add_argument(
        "--revision",
        help="Hugging Face commit override; valid only when exporting one model",
    )
    parser.add_argument("--opset", type=int, default=17, help="ONNX opset, minimum 17")
    parser.add_argument(
        "--logit-tolerance",
        type=float,
        default=1e-3,
        help="maximum absolute PyTorch versus ONNX logit difference",
    )
    parser.add_argument(
        "--license",
        default="unverified",
        help="reviewed weight license to embed; defaults to unverified",
    )
    parser.add_argument(
        "--exported-at",
        help="fixed RFC 3339 UTC export time for byte-reproducible metadata",
    )
    return parser.parse_args()


def sha256(file_name: Path) -> str:
    """Returns the SHA-256 digest of a file."""

    digest = hashlib.sha256()
    with file_name.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def package_versions() -> dict[str, str]:
    """Returns the versions that materially affect an export."""

    return {
        "python": platform.python_version(),
        "platform": platform.platform(),
        "torch": torch.__version__,
        "timm": timm.__version__,
        "onnx": onnx.__version__,
        "onnxruntime": onnxruntime.__version__,
        "numpy": np.__version__,
    }


def normalize_exported_at(value: str | None) -> str:
    """Returns a validated UTC export timestamp with second precision."""

    if value:
        try:
            parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
        except ValueError as error:
            raise ValueError("exported-at must be an RFC 3339 timestamp") from error
        if parsed.tzinfo is None:
            raise ValueError("exported-at must include a UTC offset")
    else:
        parsed = datetime.now(timezone.utc)

    return parsed.astimezone(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")


def select_models(names: list[str]) -> list[ModelSpec]:
    """Expands all and removes repeated model selections."""

    selected = list(MODEL_SPECS) if "all" in names else names
    return [MODEL_SPECS[name] for name in dict.fromkeys(selected)]


def checkpoint_file(spec: ModelSpec, revision: str | None) -> tuple[Path, str]:
    """Downloads an immutable publisher checkpoint and returns its revision."""

    resolved_revision = revision or spec.revision

    file_name = Path(
        hf_hub_download(
            repo_id=spec.repo_id,
            filename="model.safetensors",
            revision=resolved_revision,
        )
    )
    return file_name, resolved_revision


def load_model(spec: ModelSpec, weights: Path) -> LogitsOnly:
    """Loads the exact weights in eval mode and applies deploy reparameterization."""

    model = timm.create_model(spec.checkpoint, pretrained=False, checkpoint_path=str(weights))
    model.eval()
    if spec.reparameterize:
        model = reparameterize_model(model)
        model.eval()
    return LogitsOnly(model).eval()


def fixture_tensor(model: LogitsOnly, file_name: Path) -> tuple[torch.Tensor, dict[str, Any]]:
    """Runs the recorded PhotoPrism production preprocessing on one fixture."""

    if not file_name.is_file():
        raise FileNotFoundError(file_name)

    config = resolve_model_data_config(model.model)
    if tuple(config["input_size"]) != (3, 224, 224):
        raise RuntimeError(f"unexpected input size {config['input_size']}")

    short_edge = round(224 / float(config["crop_pct"]))
    transform = transforms.Compose(
        [
            transforms.Resize(short_edge, interpolation=str_to_interp_mode(config["interpolation"])),
            transforms.CenterCrop(224),
            transforms.ToTensor(),
            transforms.Normalize(mean=config["mean"], std=config["std"]),
        ]
    )
    with Image.open(file_name) as source:
        tensor = transform(source.convert("RGB")).unsqueeze(0)
    return tensor, config


def metadata_values(
    spec: ModelSpec,
    config: dict[str, Any],
    revision: str,
    source_hash: str,
    opset: int,
    license_name: str,
    exported_at: str,
) -> dict[str, str]:
    """Returns the preprocessing and provenance metadata embedded in the graph."""

    crop_pct = float(config["crop_pct"])
    short_edge = round(224 / crop_pct)
    source_url = f"https://huggingface.co/{spec.repo_id}/resolve/{revision}/model.safetensors"
    mean = [value * 255 for value in config["mean"]]
    std_dev = [value * 255 for value in config["std"]]
    return {
        "photoprism.source": source_url,
        "photoprism.sourceRevision": revision,
        "photoprism.sourceSHA256": source_hash,
        "photoprism.license": license_name,
        "photoprism.checkpoint": spec.checkpoint,
        "photoprism.exported": exported_at,
        "photoprism.exportedBy": (
            f"PhotoPrism export-label-models.py; torch {torch.__version__}; "
            f"timm {timm.__version__}; onnx {onnx.__version__}"
        ),
        "photoprism.opset": str(opset),
        "photoprism.inputWidth": "224",
        "photoprism.inputHeight": "224",
        "photoprism.layout": "NCHW",
        "photoprism.colorOrder": "RGB",
        "photoprism.mean": json.dumps(mean),
        "photoprism.stdDev": json.dumps(std_dev),
        "photoprism.resizeMode": "center-crop",
        "photoprism.shortEdge": str(short_edge),
        "photoprism.cropPct": str(crop_pct),
        "photoprism.interpolation": str(config["interpolation"]),
        "photoprism.inputName": "input",
        "photoprism.outputName": "logits",
        "photoprism.outputWidth": "1000",
        "photoprism.outputCount": "1",
        "photoprism.logits": "true",
        "photoprism.quantization": "fp32",
    }


def write_metadata(file_name: Path, values: dict[str, str]) -> None:
    """Checks, infers, and annotates the exported graph."""

    graph = onnx.load(file_name)
    graph = onnx.shape_inference.infer_shapes(graph)
    onnx.checker.check_model(graph, full_check=True)
    del graph.metadata_props[:]
    for key, value in sorted(values.items()):
        item = graph.metadata_props.add()
        item.key = key
        item.value = value
    onnx.save(graph, file_name)


def stable_softmax(logits: np.ndarray) -> np.ndarray:
    """Returns finite probabilities using the production maximum subtraction."""

    shifted = logits.astype(np.float64) - np.max(logits)
    probabilities = np.exp(shifted)
    probabilities /= probabilities.sum()
    if not np.isfinite(probabilities).all() or not math.isclose(
        float(probabilities.sum()), 1.0, abs_tol=1e-4
    ):
        raise RuntimeError("export produced invalid probabilities")
    return probabilities


def verify_export(
    file_name: Path,
    fixture: torch.Tensor,
    expected: np.ndarray,
    tolerance: float,
) -> dict[str, Any]:
    """Compares one fixed fixture through PyTorch and ONNX Runtime."""

    session = onnxruntime.InferenceSession(str(file_name), providers=["CPUExecutionProvider"])
    if len(session.get_inputs()) != 1 or len(session.get_outputs()) != 1:
        raise RuntimeError("ONNX graph must resolve to one input and one output")
    if session.get_inputs()[0].shape != [1, 3, 224, 224]:
        raise RuntimeError(f"unexpected ONNX input shape {session.get_inputs()[0].shape}")

    actual = session.run(["logits"], {"input": fixture.numpy()})[0]
    if actual.shape != (1, 1000):
        raise RuntimeError(f"unexpected ONNX output shape {actual.shape}")

    difference = float(np.max(np.abs(actual - expected)))
    expected_top5 = np.argsort(expected[0])[-5:][::-1].tolist()
    actual_top5 = np.argsort(actual[0])[-5:][::-1].tolist()
    if actual_top5 != expected_top5:
        raise RuntimeError(f"top-5 mismatch: PyTorch {expected_top5}, ONNX {actual_top5}")
    if difference > tolerance:
        raise RuntimeError(f"maximum logit difference {difference} exceeds {tolerance}")

    probabilities = stable_softmax(actual[0])
    return {
        "pytorch_top5": expected_top5,
        "onnx_top5": actual_top5,
        "max_absolute_logit_difference": difference,
        "probability_sum": float(probabilities.sum()),
    }


def export_model(spec: ModelSpec, args: argparse.Namespace) -> dict[str, Any]:
    """Exports, annotates, verifies, and records one candidate."""

    weights, revision = checkpoint_file(spec, args.revision)
    model = load_model(spec, weights)
    fixture, config = fixture_tensor(model, args.fixture)
    with torch.inference_mode():
        expected = model(fixture).detach().cpu().numpy()

    model_dir = args.output_dir / spec.name
    model_dir.mkdir(parents=True, exist_ok=True)
    output = model_dir / spec.file_name
    with torch.inference_mode():
        torch.onnx.export(
            model,
            fixture,
            output,
            input_names=["input"],
            output_names=["logits"],
            opset_version=args.opset,
            dynamic_axes=None,
            do_constant_folding=True,
            dynamo=False,
        )

    source_hash = sha256(weights)
    metadata = metadata_values(
        spec,
        config,
        revision,
        source_hash,
        args.opset,
        args.license,
        args.exported_at,
    )
    write_metadata(output, metadata)
    verification = verify_export(output, fixture, expected, args.logit_tolerance)
    result = {
        "model": asdict(spec),
        "source_url": metadata["photoprism.source"],
        "source_revision": revision,
        "source_sha256": source_hash,
        "license": args.license,
        "onnx_path": str(output),
        "onnx_sha256": sha256(output),
        "onnx_size_bytes": output.stat().st_size,
        "preprocessing": {
            "input_size": config["input_size"],
            "mean_0_1": config["mean"],
            "std_dev_0_1": config["std"],
            "mean_0_255": [value * 255 for value in config["mean"]],
            "std_dev_0_255": [value * 255 for value in config["std"]],
            "crop_pct": config["crop_pct"],
            "short_edge": round(224 / float(config["crop_pct"])),
            "interpolation": config["interpolation"],
        },
        "verification": verification,
    }
    manifest = model_dir / "export.json"
    manifest.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return result


def main() -> int:
    """Runs every selected export and writes a round-level manifest."""

    args = parse_args()
    if args.opset < 17:
        raise ValueError("opset must be at least 17")
    if args.logit_tolerance <= 0:
        raise ValueError("logit tolerance must be positive")
    args.exported_at = normalize_exported_at(args.exported_at)

    selected = select_models(args.model)
    if args.revision and len(selected) != 1:
        raise ValueError("revision override requires exactly one selected model")

    command = shlex.join(sys.argv)
    results = [export_model(spec, args) for spec in selected]
    manifest = {
        "command": command,
        "exported_at": args.exported_at,
        "environment": package_versions(),
        "fixture": str(args.fixture.resolve()),
        "fixture_sha256": sha256(args.fixture),
        "models": results,
    }
    args.output_dir.mkdir(parents=True, exist_ok=True)
    output = args.output_dir / "label-model-exports.json"
    output.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(manifest, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
