# Description: This script checks if the header is present in all go files.
import glob
import os
import sys

HEADER = "© 2019-present nextmv.io inc"

GO_HEADER = f"// {HEADER}"
GO_IGNORE = []

PYTHON_HEADER = f"# {HEADER}"
PYTHON_IGNORE = ["venv/*", "tests/*"]


def main() -> None:
    """Checks if the header is present all files, for the given language."""

    check_var = os.getenv("HEADER_CHECK_LANGUAGE", "go")
    if check_var == "go":
        files = glob.glob("**/*.go", recursive=True)
        header = GO_HEADER
        ignore = GO_IGNORE
    elif check_var == "python":
        files = glob.glob("**/*.py", recursive=True)
        header = PYTHON_HEADER
        ignore = PYTHON_IGNORE
    else:
        raise ValueError(f"Unsupported language: {check_var}")

    check(files, header, ignore)


def check(files: list[str], header: str, ignore: list[str]) -> None:
    """Checks if the header is present in all files."""

    # Check if the header is the first line of each file
    missing = []
    checked = 0
    for file in files:
        # Check if the path is in the ignore list with a glob pattern.
        if any(glob.fnmatch.fnmatch(file, pattern) for pattern in ignore):
            continue

        with open(file) as f:
            first_line = f.readline().strip()
            if first_line != header:
                missing.append(file)
            checked += 1

    # Print the results
    if missing:
        print(f"Missing header in {len(missing)} of {checked} files:")
        for file in missing:
            print(f"  {file}")
        sys.exit(1)
    else:
        print(f"Header is present in all {checked} files")


if __name__ == "__main__":
    main()
