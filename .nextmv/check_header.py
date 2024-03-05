# Description: This script checks if the header is present in all go files.
import glob
import sys

HEADER = "// © 2019-present nextmv.io inc"

# List all go files in all subdirectories
go_files = glob.glob("**/*.go", recursive=True)

# Check if the header is the first line of each file
missing = []
checked = 0
for file in go_files:
    with open(file, "r") as f:
        first_line = f.readline().strip()
        if first_line != HEADER:
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
