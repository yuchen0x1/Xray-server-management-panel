import os
import re
import socket
import subprocess
import sys
import tarfile
import tempfile
import time
import urllib.error
import urllib.request
import zipfile
from pathlib import Path


def extract_archive(archive, directory):
    if archive.suffix == ".zip":
        with zipfile.ZipFile(archive) as package:
            package.extractall(directory)
    else:
        with tarfile.open(archive) as package:
            package.extractall(directory, filter="data")


def check_http(base_url, process):
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    for _ in range(30):
        if process.poll() is not None:
            raise RuntimeError(f"Panel exited with status {process.returncode}")
        try:
            with opener.open(base_url, timeout=2) as response:
                page = response.read().decode()
            match = re.search(r'<script[^>]+src="([^"]+\.js)"', page)
            if not match:
                raise RuntimeError("Login page is missing its embedded JavaScript")
            asset_url = urllib.parse.urljoin(base_url, match[1])
            if not asset_url.startswith(base_url):
                raise RuntimeError(f"Unexpected asset URL: {asset_url}")
            with opener.open(asset_url, timeout=2) as response:
                if "javascript" not in response.headers.get("Content-Type", ""):
                    raise RuntimeError("Embedded asset did not serve JavaScript")
            return
        except urllib.error.URLError:
            time.sleep(1)
    raise RuntimeError("Timed out waiting for the packaged panel")


def expected_panel_version(base_version, environment):
    ref = environment.get("GITHUB_REF", "")
    if not ref or ref.startswith("refs/tags/"):
        return base_version
    commit = environment.get("GITHUB_SHA", "")
    if not re.fullmatch(r"[0-9a-fA-F]{40}", commit):
        raise RuntimeError("A non-tag CI archive requires a valid GITHUB_SHA")
    return "dev+" + commit[:8]


def smoke(archive):
    with tempfile.TemporaryDirectory(prefix="x-ui-smoke-") as temporary:
        directory = Path(temporary)
        extract_archive(archive, directory)
        panel = directory / "x-ui"
        windows = archive.suffix == ".zip"
        executable = panel / ("x-ui.exe" if windows else "x-ui")
        xray = panel / "bin" / ("xray-windows-amd64.exe" if windows else "xray-linux-amd64")
        if not windows:
            for name in ("update.sh", "x-ui.sh", "x-ui.rc", "x-ui.service.debian",
                         "x-ui.service.arch", "x-ui.service.rhel"):
                if not (panel / name).is_file():
                    raise RuntimeError(f"Missing Linux service file {name}")
        if not (panel / "LICENSE").is_file():
            raise RuntimeError("Missing project license")
        subprocess.run([str(xray), "version"], cwd=panel, check=True, timeout=15)
        version = subprocess.check_output([str(executable), "-v"], cwd=panel, text=True).strip()
        base_version = (Path(__file__).resolve().parents[2] / "internal/config/version").read_text().strip()
        expected = expected_panel_version(base_version, os.environ)
        if version != expected:
            raise RuntimeError(f"Packaged version {version!r} differs from {expected!r}")
        with socket.socket() as listener:
            listener.bind(("127.0.0.1", 0))
            port = listener.getsockname()[1]
        env = dict(os.environ, XUI_PORT=str(port), XUI_DB_TYPE="sqlite",
                   XUI_DB_FOLDER=str(directory / "db"), XUI_LOG_FOLDER=str(directory),
                   XUI_BIN_FOLDER=str(panel / "bin"), XUI_DEBUG="false")
        subprocess.run([str(executable), "setting", "-listenIP", "127.0.0.1",
                        "-webBasePath", "/archive-smoke/"], cwd=panel, env=env,
                       check=True, timeout=15)
        log = directory / "panel.log"
        with log.open("wb") as output:
            process = subprocess.Popen([str(executable)], cwd=panel, env=env,
                                       stdout=output, stderr=subprocess.STDOUT)
        try:
            check_http(f"http://127.0.0.1:{port}/archive-smoke/", process)
        except RuntimeError:
            print(log.read_text(errors="replace"), file=sys.stderr)
            raise
        finally:
            if process.poll() is None:
                if windows:
                    subprocess.run(["taskkill", "/PID", str(process.pid), "/T", "/F"], check=False)
                else:
                    process.terminate()
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait()
        if not (directory / "db" / "x-ui.db").is_file():
            raise RuntimeError("The packaged panel did not initialize SQLite")
    print(f"Archive verified: {archive.name} ({version})")


if __name__ == "__main__":
    smoke(Path(sys.argv[1]).resolve())
