#!/usr/bin/env python3
"""Exports and verifies checksum-recorded NSFW classifiers for PhotoPrism."""

from __future__ import annotations

import argparse
import hashlib
import importlib.metadata
import json
import math
import os
import platform
import shlex
import shutil
import sys
import urllib.request
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

os.environ.setdefault("PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION", "python")

import numpy as np
import onnx
import onnxruntime
import timm
import torch
import transformers
from huggingface_hub import hf_hub_download
from PIL import Image
from safetensors.torch import load_file
from torchvision import transforms
from torchvision.transforms import InterpolationMode
from transformers import AutoModelForImageClassification


@dataclass(frozen=True)
class ModelSpec:
    """Describes one immutable publisher artifact and its export contract."""

    name: str
    repo_id: str
    revision: str
    source_file: str
    output_file: str
    license: str
    architecture: str
    width: int
    mean: tuple[float, float, float]
    std_dev: tuple[float, float, float]
    interpolation: str
    output_width: int
    unsafe_index: int | None = None
    neutral_index: int | None = None
    publisher_onnx: bool = False
    publisher_caffe: bool = False
    outputs_logits: bool = True
    resize_mode: str = "stretch"
    color_order: str = "RGB"
    input_name: str = "input"
    output_name: str = "logits"


MODEL_SPECS = {
    "adamcodd_vit_base_nsfw_fp32": ModelSpec(
        name="adamcodd_vit_base_nsfw_fp32",
        repo_id="AdamCodd/vit-base-nsfw-detector",
        revision="8587de998f441aac03fdd57a85d2e4cb808c7d64",
        source_file="onnx/model.onnx",
        output_file="adamcodd_vit_base_nsfw_fp32.onnx",
        license="Apache-2.0",
        architecture="publisher-onnx",
        width=384,
        mean=(0.5, 0.5, 0.5),
        std_dev=(0.5, 0.5, 0.5),
        interpolation="bilinear",
        output_width=2,
        unsafe_index=1,
        publisher_onnx=True,
    ),
    "adamcodd_vit_base_nsfw_int8": ModelSpec(
        name="adamcodd_vit_base_nsfw_int8",
        repo_id="AdamCodd/vit-base-nsfw-detector",
        revision="8587de998f441aac03fdd57a85d2e4cb808c7d64",
        source_file="onnx/model_int8.onnx",
        output_file="adamcodd_vit_base_nsfw_int8.onnx",
        license="Apache-2.0",
        architecture="publisher-onnx-int8",
        width=384,
        mean=(0.5, 0.5, 0.5),
        std_dev=(0.5, 0.5, 0.5),
        interpolation="bilinear",
        output_width=2,
        unsafe_index=1,
        publisher_onnx=True,
    ),
    "falconsai_nsfw_image_detection_224": ModelSpec(
        name="falconsai_nsfw_image_detection_224",
        repo_id="Falconsai/nsfw_image_detection",
        revision="04367978d3474804ab1a00a9bd6548b741764069",
        source_file="model.safetensors",
        output_file="falconsai_nsfw_image_detection_224.onnx",
        license="Apache-2.0",
        architecture="transformers-vit",
        width=224,
        mean=(0.5, 0.5, 0.5),
        std_dev=(0.5, 0.5, 0.5),
        interpolation="bilinear",
        output_width=2,
        unsafe_index=1,
    ),
    "freepik_nsfw_image_detector": ModelSpec(
        name="freepik_nsfw_image_detector",
        repo_id="Freepik/nsfw_image_detector",
        revision="15b85477e4fd2000db76ae9aae0f89a72f95e2e3",
        source_file="model.safetensors",
        output_file="freepik_nsfw_image_detector.onnx",
        license="MIT",
        architecture="eva02_base_patch14_448",
        width=448,
        mean=(0.48145466, 0.4578275, 0.40821073),
        std_dev=(0.26862954, 0.26130258, 0.27577711),
        interpolation="bicubic",
        output_width=4,
        neutral_index=0,
    ),
    "yahoo_open_nsfw": ModelSpec(
        name="yahoo_open_nsfw",
        repo_id="yahoo/open_nsfw",
        revision="a4e13931465f4380742545932657eeea0a10aa48",
        source_file="nsfw_model/resnet_50_1by2_nsfw.caffemodel",
        output_file="yahoo_open_nsfw.onnx",
        license="BSD-2-Clause",
        architecture="caffe-resnet-50-1by2",
        width=224,
        mean=(104 / 255, 117 / 255, 123 / 255),
        std_dev=(1 / 255, 1 / 255, 1 / 255),
        interpolation="bilinear",
        output_width=2,
        unsafe_index=1,
        publisher_caffe=True,
        outputs_logits=False,
        resize_mode="center-crop",
        color_order="BGR",
        input_name="data_input",
        output_name="prob",
    ),
}


class LogitsOnly(torch.nn.Module):
    """Returns the final logits tensor from a publisher model."""

    def __init__(self, model: torch.nn.Module, width: int) -> None:
        super().__init__()
        self.model = model
        self.width = width

    def forward(self, value: torch.Tensor) -> torch.Tensor:
        """Returns a two-dimensional tensor with the registered output width."""

        result = self.model(value)
        if hasattr(result, "logits"):
            result = result.logits
        if not isinstance(result, torch.Tensor):
            raise RuntimeError("model did not return final logits")
        if result.ndim != 2 or result.shape[1] != self.width:
            raise RuntimeError(
                f"model returned {tuple(result.shape)}, expected [1, {self.width}]"
            )
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
        help="fixture used for framework and ONNX comparison",
    )
    parser.add_argument("--opset", type=int, default=17, help="ONNX opset")
    parser.add_argument(
        "--logit-tolerance",
        type=float,
        default=1e-3,
        help="maximum framework versus ONNX logit difference",
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


def normalize_exported_at(value: str | None) -> str:
    """Returns a validated UTC export timestamp with second precision."""

    if value:
        parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
        if parsed.tzinfo is None:
            raise ValueError("exported-at must include a UTC offset")
    else:
        parsed = datetime.now(timezone.utc)
    return parsed.astimezone(timezone.utc).isoformat(timespec="seconds").replace(
        "+00:00", "Z"
    )


def select_models(names: list[str]) -> list[ModelSpec]:
    """Expands all and removes repeated model selections."""

    selected = list(MODEL_SPECS) if "all" in names else names
    return [MODEL_SPECS[name] for name in dict.fromkeys(selected)]


def fixture_tensor(spec: ModelSpec, file_name: Path) -> torch.Tensor:
    """Applies the exact publisher preprocessing to one fixture."""

    if not file_name.is_file():
        raise FileNotFoundError(file_name)
    interpolation = {
        "bilinear": InterpolationMode.BILINEAR,
        "bicubic": InterpolationMode.BICUBIC,
    }[spec.interpolation]
    with Image.open(file_name) as source:
        image = source.convert("RGB")
    if spec.resize_mode == "center-crop":
        image = transforms.Resize(256, interpolation=interpolation)(image)
        image = transforms.CenterCrop(spec.width)(image)
    else:
        image = transforms.Resize(
            (spec.width, spec.width), interpolation=interpolation
        )(image)
    pixels = np.asarray(image, dtype=np.float32)
    if spec.color_order == "BGR":
        pixels = pixels[:, :, ::-1]
    mean = np.asarray(spec.mean, dtype=np.float32) * 255
    std_dev = np.asarray(spec.std_dev, dtype=np.float32) * 255
    pixels = (pixels - mean) / std_dev
    return torch.from_numpy(pixels.copy().transpose(2, 0, 1)).unsqueeze(0)


def load_model(spec: ModelSpec, weights: Path) -> LogitsOnly:
    """Loads a publisher checkpoint in evaluation mode."""

    if spec.architecture == "transformers-vit":
        model = AutoModelForImageClassification.from_pretrained(
            spec.repo_id,
            revision=spec.revision,
            use_safetensors=True,
        )
    elif spec.architecture == "eva02_base_patch14_448":
        model = timm.create_model(spec.architecture, pretrained=False, num_classes=4)
        model.load_state_dict(load_file(weights), strict=True)
    else:
        raise ValueError(f"unsupported export architecture {spec.architecture}")
    return LogitsOnly(model.eval().float(), spec.output_width).eval()


def metadata_values(
    spec: ModelSpec, source_hash: str, opset: int, exported_at: str
) -> dict[str, str]:
    """Returns preprocessing and provenance metadata for an ONNX graph."""

    if spec.publisher_caffe:
        source = (
            f"https://raw.githubusercontent.com/{spec.repo_id}/"
            f"{spec.revision}/{spec.source_file}"
        )
    else:
        source = (
            f"https://huggingface.co/{spec.repo_id}/resolve/"
            f"{spec.revision}/{spec.source_file}"
        )
    values = {
        "photoprism.source": source,
        "photoprism.sourceRevision": spec.revision,
        "photoprism.sourceSHA256": source_hash,
        "photoprism.license": spec.license,
        "photoprism.checkpoint": spec.architecture,
        "photoprism.exported": exported_at,
        "photoprism.exportedBy": (
            f"PhotoPrism export-nsfw-models.py; torch {torch.__version__}; "
            f"transformers {transformers.__version__}; timm {timm.__version__}; "
            f"onnx {onnx.__version__}"
        ),
        "photoprism.opset": str(opset),
        "photoprism.inputWidth": str(spec.width),
        "photoprism.inputHeight": str(spec.width),
        "photoprism.layout": "NCHW",
        "photoprism.colorOrder": spec.color_order,
        "photoprism.mean": json.dumps([value * 255 for value in spec.mean]),
        "photoprism.stdDev": json.dumps([value * 255 for value in spec.std_dev]),
        "photoprism.resizeMode": spec.resize_mode,
        "photoprism.interpolation": (
            "linear" if spec.interpolation == "bilinear" else spec.interpolation
        ),
        "photoprism.inputName": spec.input_name,
        "photoprism.outputName": spec.output_name,
        "photoprism.outputWidth": str(spec.output_width),
        "photoprism.outputCount": "1",
        "photoprism.logits": str(spec.outputs_logits).lower(),
        "photoprism.quantization": "int8" if "int8" in spec.name else "fp32",
    }
    if spec.unsafe_index is not None:
        values["photoprism.unsafeClassIndex"] = str(spec.unsafe_index)
        values["photoprism.reduction"] = "softmax-unsafe"
    if spec.neutral_index is not None:
        values["photoprism.neutralClassIndex"] = str(spec.neutral_index)
        values["photoprism.reduction"] = "neutral-complement"
    return values


def write_metadata(file_name: Path, values: dict[str, str]) -> None:
    """Checks, infers, and annotates an exported graph."""

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
    """Returns finite probabilities using production maximum subtraction."""

    shifted = logits.astype(np.float64) - np.max(logits)
    probabilities = np.exp(shifted)
    probabilities /= probabilities.sum()
    if not np.isfinite(probabilities).all() or not math.isclose(
        float(probabilities.sum()), 1.0, abs_tol=1e-4
    ):
        raise RuntimeError("export produced invalid probabilities")
    return probabilities


def verify_export(
    spec: ModelSpec,
    file_name: Path,
    fixture: torch.Tensor,
    expected: np.ndarray | None,
    tolerance: float,
) -> dict[str, Any]:
    """Validates graph shape and compares publisher-framework output when available."""

    session = onnxruntime.InferenceSession(
        str(file_name), providers=["CPUExecutionProvider"]
    )
    if len(session.get_inputs()) != 1 or len(session.get_outputs()) != 1:
        raise RuntimeError("ONNX graph must resolve to one input and one output")
    input_info = session.get_inputs()[0]
    output_info = session.get_outputs()[0]
    if input_info.shape != [1, 3, spec.width, spec.width]:
        raise RuntimeError(f"unexpected ONNX input shape {input_info.shape}")
    actual = session.run([output_info.name], {input_info.name: fixture.numpy()})[0]
    if actual.shape != (1, spec.output_width):
        raise RuntimeError(f"unexpected ONNX output shape {actual.shape}")
    result: dict[str, Any] = {
        "input_name": input_info.name,
        "output_name": output_info.name,
        "probability_sum": float(stable_softmax(actual[0]).sum()),
    }
    if expected is not None:
        difference = float(np.max(np.abs(actual - expected)))
        if int(np.argmax(actual[0])) != int(np.argmax(expected[0])):
            raise RuntimeError("framework and ONNX top classes differ")
        if difference > tolerance:
            raise RuntimeError(
                f"maximum logit difference {difference} exceeds {tolerance}"
            )
        result["max_absolute_logit_difference"] = difference
    return result


def download_yahoo_source(spec: ModelSpec, model_dir: Path) -> tuple[Path, Path]:
    """Downloads the immutable Yahoo Caffe weights and graph definition."""

    base = f"https://raw.githubusercontent.com/{spec.repo_id}/{spec.revision}/"
    source = model_dir / Path(spec.source_file).name
    prototxt = model_dir / "deploy.prototxt"
    for relative, target in [
        (spec.source_file, source),
        ("nsfw_model/deploy.prototxt", prototxt),
    ]:
        if not target.is_file():
            urllib.request.urlretrieve(base + relative, target)
    return source, prototxt


def convert_yahoo_caffe(
    prototxt: Path, source: Path, output: Path, opset: int
) -> None:
    """Converts and freezes the publisher Caffe graph."""

    from caffe2onnx.src.caffe2onnx import Caffe2Onnx
    from caffe2onnx.src.load_save_model import loadcaffemodel
    from caffe2onnx.src.utils import freeze

    graph, params = loadcaffemodel(str(prototxt), str(source))
    converted = Caffe2Onnx(graph, params, str(output)).createOnnxModel()
    freeze(converted)
    for imported in converted.opset_import:
        if imported.domain == "":
            imported.version = min(opset, 13)
    onnx.checker.check_model(converted, full_check=True)
    onnx.save(converted, output)


def yahoo_caffe_output(
    prototxt: Path, source: Path, fixture: torch.Tensor
) -> np.ndarray:
    """Runs the publisher Caffe graph for numerical conversion verification."""

    import cv2

    model = cv2.dnn.readNet(str(source), str(prototxt), "caffe")
    model.setInput(fixture.numpy())
    return model.forward()


def export_model(spec: ModelSpec, args: argparse.Namespace) -> dict[str, Any]:
    """Exports or copies, annotates, verifies, and records one candidate."""

    model_dir = args.output_dir / spec.name
    model_dir.mkdir(parents=True, exist_ok=True)
    output = model_dir / spec.output_file
    expected = None
    prototxt = None
    if spec.publisher_caffe:
        source, prototxt = download_yahoo_source(spec, model_dir)
    else:
        source = Path(
            hf_hub_download(
                repo_id=spec.repo_id,
                filename=spec.source_file,
                revision=spec.revision,
            )
        )
    fixture = fixture_tensor(spec, args.fixture)
    if spec.publisher_onnx:
        shutil.copyfile(source, output)
    elif spec.publisher_caffe:
        assert prototxt is not None
        convert_yahoo_caffe(prototxt, source, output, args.opset)
        expected = yahoo_caffe_output(prototxt, source, fixture)
    else:
        model = load_model(spec, source)
        with torch.inference_mode():
            expected = model(fixture).detach().cpu().numpy()
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
    metadata = metadata_values(spec, sha256(source), args.opset, args.exported_at)
    if not spec.publisher_onnx:
        write_metadata(output, metadata)
    verification = verify_export(
        spec, output, fixture, expected, args.logit_tolerance
    )
    result = {
        "model": asdict(spec),
        "source_url": metadata["photoprism.source"],
        "source_sha256": sha256(source),
        "onnx_path": str(output),
        "onnx_sha256": sha256(output),
        "onnx_size_bytes": output.stat().st_size,
        "verification": verification,
    }
    if prototxt is not None:
        result["prototxt_sha256"] = sha256(prototxt)
    (model_dir / "export.json").write_text(
        json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    return result


def package_versions() -> dict[str, str]:
    """Returns package versions that materially affect an export."""

    result = {
        "python": platform.python_version(),
        "platform": platform.platform(),
        "torch": torch.__version__,
        "transformers": transformers.__version__,
        "timm": timm.__version__,
        "onnx": onnx.__version__,
        "onnxruntime": onnxruntime.__version__,
        "numpy": np.__version__,
    }
    for package in ("caffe2onnx", "opencv-python-headless"):
        try:
            result[package] = importlib.metadata.version(package)
        except importlib.metadata.PackageNotFoundError:
            pass
    return result


def main() -> int:
    """Runs selected exports and writes a round-level manifest."""

    args = parse_args()
    if args.opset < 17:
        raise ValueError("opset must be at least 17")
    if args.logit_tolerance <= 0:
        raise ValueError("logit tolerance must be positive")
    args.exported_at = normalize_exported_at(args.exported_at)
    results = [export_model(spec, args) for spec in select_models(args.model)]
    manifest = {
        "command": shlex.join(sys.argv),
        "exported_at": args.exported_at,
        "environment": package_versions(),
        "fixture": str(args.fixture.resolve()),
        "fixture_sha256": sha256(args.fixture),
        "models": results,
    }
    args.output_dir.mkdir(parents=True, exist_ok=True)
    output = args.output_dir / "nsfw-model-exports.json"
    output.write_text(
        json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    print(json.dumps(manifest, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
