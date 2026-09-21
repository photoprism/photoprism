#!/usr/bin/env python3
"""Check CUDA installation and recovery with synthetic packages and a private prefix."""

import json
import os
from pathlib import Path
import shutil
import signal
import stat
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
INSTALLER = ROOT / "scripts/dist/install-cuda.sh"
LIBRARIES = (
    "libcudart", "libcublas", "libcublasLt", "libcurand", "libcudnn",
    "libcudnn_adv", "libcudnn_cnn", "libcudnn_engines_precompiled",
    "libcudnn_engines_runtime_compiled", "libcudnn_engines_tensor_ir",
    "libcudnn_ext", "libcudnn_graph", "libcudnn_heuristic", "libcudnn_ops",
)

STUB = r'''
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys

base = Path(os.environ["CUDA_CHECK_ROOT"])
command = Path(sys.argv[0]).name
args = sys.argv[1:]
mode = os.environ["CUDA_CHECK_FAILURE"]
real = json.loads(os.environ["CUDA_CHECK_COMMANDS"])

# check_path refuses filesystem operations outside the isolated fixture.
def check_path(value):
    path = Path(os.path.abspath(value))
    if not path.is_relative_to(base):
        raise RuntimeError("command escaped the test prefix: " + command)
    return path

# count records operation progress independently of library enumeration order.
def count(kind):
    path = base / (kind + ".count")
    value = int(path.read_text()) + 1 if path.exists() else 1
    path.write_text(str(value))
    return value

# stop delivers a catchable signal to the installer without changing the test process.
def stop():
    os.kill(os.getppid(), int(os.environ["CUDA_CHECK_SIGNAL"]))
    sys.exit(0)

if command == "curl":
    destination = check_path(args[args.index("-o") + 1])
    with (base / "downloads.jsonl").open("a") as log:
        log.write(json.dumps({"url": args[-1], "file": destination.name}) + "\n")
    if count("download") == 2 and mode == "download":
        sys.exit(1)
    destination.write_text("synthetic package")
elif command == "sha256sum":
    sys.stdin.read()
elif command == "dpkg-deb":
    destination = check_path(args[-1])
    if not (destination / "usr").exists():
        shutil.copytree(base / "package", destination, dirs_exist_ok=True, symlinks=True)
elif command == "getconf":
    print("glibc 2.39")
elif command == "ldconfig":
    (base / "ldconfig.called").touch()
    sys.exit(1 if mode == "ldconfig" else 0)
else:
    if command in ("cp", "mv"):
        source, target = check_path(args[-2]), check_path(args[-1])
    elif command in ("mkdir", "rm"):
        for arg in args:
            if not arg.startswith("-"):
                check_path(arg)
    elif command == "chmod":
        target = check_path(args[-1])

    if command == "cp":
        kind = "backup" if "--link" in args else "copy"
        if count(kind) == 2 and mode == kind:
            sys.exit(1)
    elif command == "chmod" and target.parent.name == "new":
        if count("mode") == 2 and mode == "mode":
            sys.exit(1)
    elif command == "mv" and source.parent.name == "new":
        position = count("publish")
        if position == 2 and mode == "publish-before":
            sys.exit(1)
        if position == 2 and mode == "signal-before":
            stop()
    elif command == "mv" and source.parent.name == "previous":
        if count("restore") == 1 and mode == "rollback":
            sys.exit(1)

    result = subprocess.run([real[command], *args]).returncode
    if result:
        sys.exit(result)

    if command == "mv" and source.parent.name == "new" and position == 2:
        if mode in ("publish-after", "rollback"):
            sys.exit(1)
        if mode == "signal-after":
            stop()
'''


@unittest.skipUnless(sys.platform.startswith("linux"), "the CUDA installer targets Linux")
class CudaInstallTest(unittest.TestCase):
    """Exercise the real installer without network requests or system-library writes."""

    def setUp(self):
        """Create a synthetic package, command stubs, and a private destination."""
        folder = tempfile.TemporaryDirectory(prefix="photoprism-cuda-check-")
        self.addCleanup(folder.cleanup)
        self.base = Path(folder.name).resolve()
        self.prefix = self.base / "prefix"
        self.output = self.prefix / "lib"
        self.output.mkdir(parents=True)
        self.downloads = self.base / "downloads"
        self.downloads.mkdir()
        self.package = self.base / "package/usr/lib/x86_64-linux-gnu"
        self.package.mkdir(parents=True)
        for name in LIBRARIES:
            library = self.package / (name + ".so.1")
            library.write_text("new " + name)
            library.chmod(0o4755)
            (self.package / (name + ".so")).symlink_to(library.name)
        self.note = self.output / "operator-note"
        self.note.write_text("leave this alone")
        self.commands = self.base / "bin"
        self.commands.mkdir()
        real = {}
        for name in ("cp", "mv", "chmod", "mkdir", "rm"):
            real[name] = shutil.which(name)
            self.assertIsNotNone(real[name], name + " is required")
        for name in (*real, "curl", "sha256sum", "dpkg-deb", "getconf", "ldconfig"):
            stub = self.commands / name
            stub.write_text("#!" + sys.executable + "\n" + STUB)
            stub.chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.commands) + os.pathsep + os.environ["PATH"],
                        PHOTOPRISM_ARCH="amd64", TMPDIR=str(self.downloads),
                        CUDA_CHECK_ROOT=str(self.base), CUDA_CHECK_COMMANDS=json.dumps(real))

    def populate(self):
        """Seed existing libraries, including a dangling symlink and a retained old version."""
        for name in LIBRARIES:
            library = self.output / (name + ".so.1")
            library.write_text("original " + name)
            library.chmod(0o640)
            os.utime(library, ns=(1_700_000_000_000_000_000,) * 2)
            (self.output / (name + ".so")).symlink_to(library.name)
        link = self.output / "libcudart.so"
        link.unlink()
        link.symlink_to("absent-original")
        (self.output / "libcudart.so.0").write_text("retained old version")

    def snapshot(self):
        """Capture destination contents and metadata without following symlinks."""
        result = {}
        for path in self.output.iterdir():
            metadata = path.lstat()
            if path.is_symlink():
                value = ("link", os.readlink(path))
            elif path.is_file():
                value = ("file", path.read_bytes())
            else:
                value = ("directory",)
            result[path.name] = (value, stat.S_IMODE(metadata.st_mode), metadata.st_ino,
                                 metadata.st_mtime_ns, metadata.st_uid, metadata.st_gid)
        return result

    def run_installer(self, failure="", stop_signal=signal.SIGTERM):
        """Run one installer attempt with bounded execution and synthetic failure injection."""
        env = dict(self.env, CUDA_CHECK_FAILURE=failure, CUDA_CHECK_SIGNAL=str(int(stop_signal)))
        result = subprocess.run(["bash", str(INSTALLER), str(self.prefix)], env=env,
                                capture_output=True, text=True, timeout=30)
        self.assertEqual([], list(self.downloads.iterdir()), result.stderr)
        if failure and failure != "ldconfig":
            phase = "publish" if failure.startswith(("publish-", "signal-")) else failure
            phase = "restore" if phase == "rollback" else phase
            counter = self.base / (phase + ".count")
            self.assertTrue(counter.exists(), result.stdout + result.stderr)
            self.assertGreaterEqual(int(counter.read_text()), 1 if phase == "restore" else 2)
        return result

    def assert_restored(self, before, result):
        """Require failed installation to restore every destination entry exactly."""
        self.assertNotEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertEqual(before, self.snapshot(), result.stderr)
        self.assertFalse((self.base / "ldconfig.called").exists())

    def test_success(self):
        """A successful reinstall publishes complete libraries without mutating old inodes."""
        self.populate()
        with (self.output / "libcudart.so.1").open() as original:
            result = self.run_installer()
            self.assertEqual(0, result.returncode, result.stdout + result.stderr)
            self.assertEqual("original libcudart", original.read())
        self.assert_installed()
        self.assertEqual("retained old version", (self.output / "libcudart.so.0").read_text())

    def assert_installed(self):
        """Require complete libraries, normalized modes, and preserved unrelated files."""
        for name in LIBRARIES:
            path = self.output / (name + ".so.1")
            self.assertEqual("new " + name, path.read_text())
            self.assertEqual(0o644, stat.S_IMODE(path.stat().st_mode))
            self.assertEqual(path.name, os.readlink(self.output / (name + ".so")))
        self.assertEqual([], list(self.output.glob(".cuda-*")))
        self.assertEqual("leave this alone", self.note.read_text())
        self.assertTrue((self.base / "ldconfig.called").exists())

    def test_first_install_success(self):
        """A first installation fetches only from NVIDIA and publishes the complete set."""
        result = self.run_installer()
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assert_installed()
        downloads = [json.loads(line) for line in (self.base / "downloads.jsonl").read_text().splitlines()]
        self.assertEqual(4, len(downloads))
        for item in downloads:
            self.assertEqual("https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2604/x86_64/"
                             + item["file"], item["url"])

    def test_download_failure(self):
        """An unsuccessful download leaves the installed libraries unchanged."""
        self.populate()
        before = self.snapshot()
        result = self.run_installer("download")
        self.assert_restored(before, result)
        self.assertIn("Failed to download", result.stderr)
        self.assertEqual(2, len((self.base / "downloads.jsonl").read_text().splitlines()))

    def test_private_prefix_ldconfig_failure(self):
        """A private-prefix cache refresh failure preserves the committed installation."""
        self.populate()
        result = self.run_installer("ldconfig")
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assert_installed()
        self.assertEqual("retained old version", (self.output / "libcudart.so.0").read_text())

    def test_preparation_failures(self):
        """Copy, snapshot, and mode-setting failures leave a prior installation untouched."""
        self.populate()
        before = self.snapshot()
        for failure in ("copy", "backup", "mode"):
            with self.subTest(failure=failure):
                for counter in self.base.glob("*.count"):
                    counter.unlink()
                self.assert_restored(before, self.run_installer(failure))

    def test_publication_failures(self):
        """A failed rename before or after replacement restores the old installation."""
        self.populate()
        before = self.snapshot()
        for failure in ("publish-before", "publish-after"):
            with self.subTest(failure=failure):
                for counter in self.base.glob("*.count"):
                    counter.unlink()
                self.assert_restored(before, self.run_installer(failure))

    def test_first_install_failure(self):
        """Failed publication removes new entries without removing unrelated files."""
        before = self.snapshot()
        self.assert_restored(before, self.run_installer("publish-after"))

    def test_signals(self):
        """Catchable stop signals restore files on either side of an atomic rename."""
        self.populate()
        before = self.snapshot()
        for stop_signal in (signal.SIGHUP, signal.SIGINT, signal.SIGTERM):
            for failure in ("signal-before", "signal-after"):
                with self.subTest(signal=stop_signal, failure=failure):
                    for counter in self.base.glob("*.count"):
                        counter.unlink()
                    result = self.run_installer(failure, stop_signal)
                    self.assertEqual(128 + stop_signal, result.returncode, result.stderr)
                    self.assert_restored(before, result)

    def test_first_install_signal(self):
        """A stopped first installation removes its newly published subset."""
        before = self.snapshot()
        result = self.run_installer("signal-after")
        self.assertEqual(143, result.returncode, result.stderr)
        self.assert_restored(before, result)

    def test_rollback_failure(self):
        """A failed restoration reports and retains the remaining recovery snapshot."""
        self.populate()
        result = self.run_installer("rollback")
        self.assertNotEqual(0, result.returncode)
        recovery = list(self.output.glob(".cuda-*"))
        self.assertEqual(1, len(recovery), result.stderr)
        self.assertIn(str(recovery[0]), result.stderr)
        self.assertIn("recovery is incomplete", result.stderr)
        self.assertTrue(list((recovery[0] / "previous").iterdir()))

    def test_directory_destination(self):
        """An existing directory at a library path is refused without publication."""
        (self.output / "libcublas.so.1").mkdir()
        before = self.snapshot()
        self.assert_restored(before, self.run_installer())

    def test_duplicate_library(self):
        """Colliding package basenames are refused before replacing an existing library."""
        self.populate()
        duplicate = self.base / "package/usr/other"
        duplicate.mkdir()
        (duplicate / "libcudart.so.1").write_text("duplicate")
        before = self.snapshot()
        result = self.run_installer()
        self.assertIn("duplicate library", result.stderr)
        self.assert_restored(before, result)


if __name__ == "__main__":
    unittest.main(verbosity=2)
