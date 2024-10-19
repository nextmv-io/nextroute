# © 2019-present nextmv.io inc

import os
import subprocess


def nextroute_version() -> str:
    """
    Get the version of the embedded Nextroute binary.
    """
    executable = os.path.join(os.path.dirname(__file__), "bin", "nextroute.exe")
    return subprocess.check_output([executable, "--version"]).decode().strip()
