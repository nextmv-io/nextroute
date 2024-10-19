# © 2019-present nextmv.io inc

import os
import platform
import subprocess

from setuptools import Distribution, setup

try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

    class MyWheel(_bdist_wheel):
        def finalize_options(self):
            _bdist_wheel.finalize_options(self)
            self.root_is_pure = False

        def get_tag(self):
            python, abi, plat = _bdist_wheel.get_tag(self)
            # python, abi = "py3", "none"
            return python, abi, plat

    class MyDistribution(Distribution):
        def __init__(self, *attrs):
            Distribution.__init__(self, *attrs)
            self.cmdclass["bdist_wheel"] = MyWheel

        def is_pure(self):
            return False

        def has_ext_modules(self):
            return True

except ImportError:

    class MyDistribution(Distribution):
        def is_pure(self):
            return False

        def has_ext_modules(self):
            return True


# Compile Nextroute binary. We cross-compile (if necessary) for the current
# platform. We also set CGO_ENABLED=0 to ensure that the binary is statically
# linked.
goos = platform.system().lower()
goarch = platform.machine().lower()

if goos not in ["linux", "windows", "darwin"]:
    raise Exception(f"unsupported operating system: {goos}")

# Translate the architecture to the Go convention.
if goarch == "x86_64":
    goarch = "amd64"
elif goarch == "aarch64":
    goarch = "arm64"

if goarch not in ["amd64", "arm64"]:
    raise Exception(f"unsupported architecture: {goarch}")

# Compile the binary.
print(f"Compiling Nextroute binary for {goos} {goarch}...")
cwd = os.getcwd()
standalone_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "cmd")
os.chdir(standalone_dir)
call = ["go", "build", "-o", "../src/nextroute/bin/nextroute.exe", "."]

try:
    subprocess.check_call(
        call,
        env={
            **os.environ,
            "GOOS": goos,
            "GOARCH": goarch,
            "CGO_ENABLED": "0",
        },
    )
finally:
    os.chdir(cwd)


# Get version from version file.
__version__ = "v0.0.0"
exec(open("./src/nextroute/__about__.py").read())

# Setup package.
setup(
    distclass=MyDistribution,
    version=__version__,
)
