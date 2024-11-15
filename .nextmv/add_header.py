# Description: This script adds a header to all go files that are missing it.
import glob

HEADER = "// © 2019-present nextmv.io inc"

# List all go files in all subdirectories
go_files = glob.glob("**/*.go", recursive=True)

# Check if the header is the first line of each file
missing = []
checked = 0
for file in go_files:
    with open(file) as f:
        first_line = f.readline().strip()
        if first_line != HEADER:
            missing.append(file)
        checked += 1

# Add the header to all missing files
for file in missing:
    print(f"Adding header to {file}")
    with open(file) as f:
        content = f.read()
    with open(file, "w") as f:
        f.write(HEADER + "\n\n" + content)

print(f"Checked {checked} files, added header to {len(missing)} files")
