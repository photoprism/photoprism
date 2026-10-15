"""Create repeatable Export acceptance media and rows in an isolated SQLite library."""
import argparse
import hashlib
import json
import os
import secrets
from pathlib import Path
import sqlite3
import subprocess
import tempfile

from PIL import Image


# run executes fixture tools without exposing their configuration or credentials.
def run(args, env=None):
    result = subprocess.run(args, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if result.returncode:
        raise RuntimeError(f"Fixture tool {Path(args[0]).name} failed; verify its dependencies and isolated configuration")
    return result.stdout


# insert copies a template row with explicit overrides into the isolated database.
def insert(db, table, template, values):
    row = dict(template)
    row.pop("id", None)
    row.update(values)
    keys = list(row)
    return db.execute(f'INSERT INTO {table} ({",".join(keys)}) VALUES ({",".join("?" for _ in keys)})', list(row.values())).lastrowid


# main generates sources, original stacks, previews, and a complete backend manifest.
def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", required=True, help="Disposable library directory below .local/")
    parser.add_argument("--binary", default="./photoprism")
    parser.add_argument("--edition", choices=["ce", "plus", "pro"], default="ce")
    args = parser.parse_args()
    repo = Path.cwd().resolve()
    root = Path(args.root).resolve()
    if not root.is_relative_to(repo / ".local") or root == repo / ".local":
        raise ValueError("Fixture root must be a directory below this repository's .local/")
    marker = root / ".export-fixtures"
    if root.exists() and not marker.is_file():
        raise ValueError("Refusing to modify an existing directory without the Export fixture marker")
    root.mkdir(parents=True, exist_ok=True)
    marker.touch()
    for directory in ["originals", "sidecar", "storage", "config"]:
        location = root / directory
        if location.is_symlink():
            raise ValueError("Fixture directories must not be symlinks")
        location.mkdir(exist_ok=True)
    binary = str(Path(args.binary).resolve())
    env = os.environ.copy()
    overrides = {
        "PHOTOPRISM_CONFIG_PATH": str(root / "config"), "PHOTOPRISM_STORAGE_PATH": str(root / "storage"),
        "PHOTOPRISM_ORIGINALS_PATH": str(root / "originals"), "PHOTOPRISM_SIDECAR_PATH": str(root / "sidecar"),
        "PHOTOPRISM_ASSETS_PATH": str(repo / "assets"), "PHOTOPRISM_DATABASE_DRIVER": "sqlite",
        "PHOTOPRISM_CACHE_PATH": str(root / "storage/cache"), "PHOTOPRISM_TEMP_PATH": str(root / "storage/temp"),
        "PHOTOPRISM_IMPORT_PATH": str(root / "import"), "PHOTOPRISM_BACKUP_PATH": str(root / "storage/backups"),
        "PHOTOPRISM_DATABASE_DSN": str(root / "storage/index.db"), "PHOTOPRISM_AUTH_MODE": "public",
        "PHOTOPRISM_DISABLE_TENSORFLOW": "true", "PHOTOPRISM_DISABLE_FACES": "true",
        "PHOTOPRISM_DISABLE_CLASSIFICATION": "true", "PHOTOPRISM_DISABLE_PLACES": "true",
        "PHOTOPRISM_DISABLE_VECTORS": "true", "PHOTOPRISM_INDEX_WORKERS": "1",
        "PHOTOPRISM_THUMB_UNCACHED": "true", "PHOTOPRISM_HTTP_HOST": "127.0.0.1",
    }
    env.update(overrides)
    (root / "environment.json").write_text(json.dumps(overrides, indent=2))
    (root / "config/settings.yml").write_text("Features:\n  Export: true\n  Download: false\n  Places: false\n")
    seed = root / "originals/fixture-seed/export-seed.jpg"
    seed.parent.mkdir(exist_ok=True)
    if not seed.exists():
        Image.new("RGB", (1600, 1000), "orange").save(seed)
    run([binary, "index", "fixture-seed"], env)
    db = sqlite3.connect(root / "storage/index.db")
    db.row_factory = sqlite3.Row
    template = db.execute("SELECT * FROM files WHERE file_name='fixture-seed/export-seed.jpg' AND file_root='/'").fetchone()
    if template is None:
        raise RuntimeError("Mandatory indexed seed is missing")
    photo_template = db.execute("SELECT * FROM photos WHERE id=?", (template["photo_id"],)).fetchone()
    # Generate fixture identities using PhotoPrism's UID generator.
    with tempfile.TemporaryDirectory(dir=root) as scratch:
        helper = Path(scratch) / "ids.go"
        helper.write_text('package main\nimport("encoding/json";"os";"github.com/photoprism/photoprism/pkg/rnd")\nfunc main(){ids:=make([]map[string]string,256);for i:=range ids{ids[i]=map[string]string{"photo":rnd.GenerateUID(\'p\'),"file":rnd.GenerateUID(\'f\')}};json.NewEncoder(os.Stdout).Encode(ids)}\n')
        ids = iter(json.loads(run(["go", "run", str(helper)])))
    cases = []
    counter = 0

    # source creates a real media source and, when needed, an independently sized still preview.
    def source(label, format, media="image", codec="", preview=False):
        nonlocal counter
        counter += 1
        name = f"export-{label}-{counter}.{format}"
        filename = root / "originals" / name
        width, height = (160, 120) if codec or media == "animated" else (1600, 1000)
        color = (counter * 37 % 255, counter * 71 % 255, counter * 19 % 255)
        image = Image.new("RGB", (width, height), color)
        if codec:
            command = ["ffmpeg", "-v", "error", "-f", "lavfi", "-i", f"color=c=0x{color[0]:02x}{color[1]:02x}{color[2]:02x}:size=160x120:rate=10", "-t", "0.4", "-threads", "1", "-c:v", codec]
            if codec == "libx265":
                command += ["-x265-params", "pools=1:frame-threads=1:log-level=error", "-tag:v", "hvc1"]
            run(command + ["-y", str(filename)])
        elif media == "animated":
            image.save(filename, save_all=True, append_images=[Image.new("RGB", image.size, "white")], duration=200, loop=0)
        elif format == "dng":
            filename.write_bytes((repo / "assets/samples/canon_eos_6d.dng").read_bytes() + name.encode())
        elif format == "heic":
            filename.write_bytes((repo / "assets/samples/iphone_7.heic").read_bytes())
        elif format == "svg":
            filename.write_text(f'<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="1000"><rect width="1600" height="1000" fill="rgb{color}"/></svg>')
        else:
            image.save(filename, format="PDF" if format == "pdf" else None)
        rows = [(name, format, media, width, height, "/")]
        if preview:
            for ext, dimensions in [("jpg", (800, 600)), ("png", (600, 300))]:
                Image.new("RGB", dimensions, color).save(root / "sidecar" / f"{name}.{ext}")
                rows.append((f"{name}.{ext}", ext, "image", *dimensions, "sidecar"))
        return rows

    # motion embeds the exact native HEVC movie bytes in a JPEG original.
    def motion(label):
        still = source(label, "jpg")
        movie = source(label + "-native", "mov", "video", "libx265")
        (root / "originals" / still[0][0]).write_bytes((root / "originals" / still[0][0]).read_bytes() + (root / "originals" / movie[0][0]).read_bytes())
        name, format, _, w, h, directory = still[0]
        return [(name, format, "live", w, h, directory)], movie[0][0]

    # add replaces only fixture-owned rows so repeated setup never accumulates files.
    def add(label, rows, native=None):
        title = "Export Acceptance " + label
        old = db.execute("SELECT id,photo_uid FROM photos WHERE photo_title=?", (title,)).fetchone()
        if old:
            pid, uid = old["id"], old["photo_uid"]
            db.execute("DELETE FROM files WHERE photo_id=?", (pid,))
        else:
            uid = next(ids)["photo"]
            pid = insert(db, "photos", photo_template, {"photo_uid": uid, "photo_title": title, "photo_type": "image", "photo_private": 0, "photo_quality": 3, "deleted_at": None})
        files = []
        for index, (name, format, media, w, h, directory) in enumerate(rows):
            filename = root / ("originals" if directory == "/" else "sidecar") / name
            hash_value = hashlib.sha1(filename.read_bytes()).hexdigest()
            file_uid = next(ids)["file"]
            primary = index == 0
            insert(db, "files", template, {"photo_id": pid, "photo_uid": uid, "file_uid": file_uid, "file_name": name, "original_name": name, "file_root": directory, "file_type": format, "media_type": media, "file_hash": hash_value, "file_size": filename.stat().st_size, "file_width": w, "file_height": h, "file_video": int(media == "video"), "file_frames": 2 if media == "animated" else 0, "file_duration": 400000000 if media in ["video", "animated", "live"] else 0, "file_primary": int(primary), "file_sidecar": 0, "file_missing": 0, "file_error": "", "file_codec": "", "file_orientation": 1, "deleted_at": None, "time_index": None, "media_id": f'{10000000000-pid}-{int(not primary)}-{file_uid}'})
            files.append({"name": name, "hash": hash_value, "root": directory, "format": format, "width": w, "height": h, "media": media})
        cases.append({"label": label, "uid": uid, "files": files, "native": native or {}})

    add("HEVC", source("hevc", "mov", "video", "libx265"))
    add("AVC", source("avc", "mov", "video", "libx264"))
    for label, format in [("AppleJPEG", "jpg"), ("AppleHEIC", "heic")]:
        still = source(label, format, preview=format == "heic")
        movie = source(label, "mov", "video", "libx265")
        old = movie[0][0]
        name = str(Path(still[0][0]).with_suffix(".mov"))
        (root / "originals" / name).write_bytes((root / "originals" / old).read_bytes())
        movie[0] = (name, *movie[0][1:])
        add(label, still + movie)
    rows, native = motion("motion")
    add("Motion", rows, {rows[0][0]: native})
    for ext in ["gif", "png", "webp"]:
        add("Animation" + ext, source("animation", ext, "animated", preview=True))
    add("Mixed", source("mixed", "mov", "video", "libx265") + source("mixed", "jpg") + source("mixed", "gif", "animated", preview=True))
    for label, format, media, codec, preview in [("TwoImage", "jpg", "image", "", False), ("TwoRAW", "dng", "raw", "", True), ("TwoSVG", "svg", "vector", "", True), ("TwoPDF", "pdf", "document", "", True), ("TwoGIF", "gif", "animated", "", True), ("TwoAPNG", "png", "animated", "", True), ("TwoWebP", "webp", "animated", "", True), ("TwoVideo", "mov", "video", "libx265", True)]:
        add(label, source(label, format, media, codec, preview) + source(label, format, media, codec, preview))
    first, first_native = motion("two-motion")
    second, second_native = motion("two-motion")
    add("TwoMotion", first + second, {first[0][0]: first_native, second[0][0]: second_native})
    apple = []
    for _ in range(2):
        still = source("two-apple", "jpg")
        movie = source("two-apple", "mov", "video", "libx265")
        name = str(Path(still[0][0]).with_suffix(".mov"))
        (root / "originals" / name).write_bytes((root / "originals" / movie[0][0]).read_bytes())
        movie[0] = (name, *movie[0][1:])
        apple += still + movie
    add("TwoApple", apple)
    raw = source("raw-jpeg", "dng", "raw", preview=True)
    jpg = source("raw-jpeg", "jpg")
    # Name the JPEG sibling after the RAW to exercise mixed-stack ownership.
    name = str(Path(raw[0][0]).with_suffix(".jpg"))
    (root / "originals" / name).write_bytes((root / "originals" / jpg[0][0]).read_bytes())
    jpg[0] = (name, *jpg[0][1:])
    add("RAWPlusJPEG", raw + jpg)
    add("RoleJPEG", source("role", "jpg"))
    role_photo = cases[-1]["uid"]
    db.commit()
    password_path = root / "password.env"
    if not password_path.exists():
        password_path.write_text("PHOTOPRISM_EXPORT_TEST_PASSWORD=" + secrets.token_urlsafe(24) + "\n")
        password_path.chmod(0o600)
    password = password_path.read_text().strip().split("=", 1)[1]
    allowed = {"ce": ["admin"], "plus": ["admin", "user"], "pro": ["admin", "manager", "user"]}[args.edition]
    restricted = {"ce": ["guest"], "plus": ["viewer", "guest"], "pro": ["viewer", "contributor", "guest"]}[args.edition]
    role_cases = []
    for role in allowed + restricted:
        username = "export" + role
        existing = db.execute("SELECT user_uid FROM auth_users WHERE user_name=?", (username,)).fetchone()
        run([binary, "users", "mod" if existing else "add", "--role", role, "--password", password, username], env)
        role_cases.append({"role": role, "export": role in allowed, "download": True, "url": "/s/{token}/{album}" if role in restricted else "/library/browse", "photo": role_photo, "username": username, "passwordEnv": "PHOTOPRISM_EXPORT_TEST_PASSWORD"})
    role_cases.append({"role": "share-link", "export": False, "download": True, "url": "/s/{token}/{album}", "photo": role_photo})
    (root / "roles.json").write_text(json.dumps({"edition": args.edition, "sessions": role_cases}, indent=2))
    db.close()
    (root / "cases.json").write_text(json.dumps(cases, indent=2))
    print(f"Prepared {len(cases)} required backend cases in {root}")


if __name__ == "__main__":
    main()
