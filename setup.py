# © 2019-present nextmv.io inc

import os
import platform
import shutil
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
            # The wheel bundles a prebuilt Go binary but contains no C
            # extension, so it is platform-specific while being independent of
            # the interpreter version. Tagging it py3-none-<platform> instead of
            # cp3XX-cp3XX-<platform> means a single wheel per platform serves
            # every Python we support, including versions released after this
            # wheel was built.
            python, abi = "py3", "none"
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
standalone_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "cmd")

# The Go sources are deliberately not part of the source distribution, so a
# build from the sdist cannot work. Fail with an explanation rather than with a
# bare FileNotFoundError from the chdir below.
if not os.path.isdir(standalone_dir):
    raise Exception(
        "cannot build nextroute from source: the Go sources required to compile "
        "the Nextroute binary are not shipped in the source distribution. "
        "Install one of the prebuilt wheels instead. If no wheel matches your "
        "platform or Python version, build from a clone of "
        "https://github.com/nextmv-io/nextroute."
    )

# Building also needs a Go toolchain, which cannot be expressed in
# build-system.requires, so check for it up front.
if shutil.which("go") is None:
    raise Exception(
        "cannot build nextroute from source: the Go toolchain is required to "
        "compile the Nextroute binary but `go` was not found on PATH. Install Go "
        "(https://go.dev/dl/), or install one of the prebuilt wheels instead."
    )

print(f"Compiling Nextroute binary for {goos} {goarch}...")
cwd = os.getcwd()
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
