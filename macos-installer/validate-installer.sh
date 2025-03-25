#!/usr/bin/env bash

set -euo pipefail

# Expect to run this build script from root directory of the project
ROOT_DIR=$(pwd)


# Function to validate scripts in pkg
validate_scripts() {
    # Extract script files from PackageInfo using xmllint, trim leading './' and whitespace
    expected_scripts=$(xmllint --xpath '//scripts/*/attribute::file' PackageInfo | \
    sed 's/file=//g' |          # Remove 'file='
    sed 's/"//g' |              # Remove double quotes
    sed 's/^[[:space:]]*//g' |  # Remove leading spaces
    cut -c 3-)                  # Remove leading './'
    echo "Found these scripts in PackageInfo <scripts>...</scripts> xml:"
    echo "${expected_scripts}"

    # List files in the ./Scripts directory and trim whitespace
    actual_scripts=$(ls ./Scripts | sed 's/^[[:space:]]*//')
    echo "Comparing against actual scripts in ./Scripts:"
    echo "${actual_scripts}"

    # Compare the two sets of files
    for actual_script in $actual_scripts; do
        if ! echo "$expected_scripts" | grep -q "$actual_script"; then
            echo "Error: PackageInfo <scripts> does not contain script found in ./Scripts/${actual_script}."
            exit 1
        fi
    done

    echo "Scripts in PackageInfo <scripts> xml matches scripts in ./Scripts."

    # Iterate over each script in the current Scripts directory
    for script in ./Scripts/*; do
        # Extract the script name (basename)
        script_name=$(basename "$script")

        echo "Found script ${script_name} in Scripts directory."
        
        
        echo "Checking that corresponding script exists in source directory"
        source_script="../../scripts/$script_name"
        if [[ ! -f "$source_script" ]]; then
            echo "Error: Corresponding script $source_script not found."
            exit 1
        fi

        # Compute SHA256 of the source and actual script
        echo "Found corresponding script in source directory, comparing sha256 checksums"
        expected_sha256=$(shasum -a 256 "$source_script" | cut -d ' ' -f 1)
        actual_sha256=$(shasum -a 256 "$script" | cut -d ' ' -f 1)

        # Compare the expected and actual SHA256 hashes
        if [[ "$expected_sha256" != "$actual_sha256" ]]; then
            echo "Error: $script_name in Payload does not match expected sha256"
            exit 1
        fi

        echo "Checksums match for script ${script}"
    done
}

usage() {
  echo "Usage: $0 [version] [arch]"
  echo "  version: The version of the build."
  echo "  arch: The architecture of the build <am64|arm64>."
  echo "Example: $0 1.0.0 amd64"
}

if [ $# -ne 2 ]; then
    echo "Error: incorrect number of arguments."
    usage
    exit 1
fi

VERSION="$1"
ARCH="$2"
echo "validating installer for ${ARCH} and version: ${VERSION}"

pushd "$ROOT_DIR/macos-installer/package"

# TODO: uncomment these checks once we have code signing and notarization
# pkgutil --check-signature packet-sentry-agent.pkg
# if [ $? -ne 0 ]; then
#     echo "Installer pkg is not code signed"
#     exit 1
# fi
# spctl --assess --type install packet-sentry-agent.pkg
# if [ $? -ne 0 ]; then
#     echo "Installer pkg is not notarized"
#     exit 1
# fi

# Clean up previous expansions.
rm -rf expanded

pkgutil --expand packet-sentry-agent.pkg expanded
cd ./expanded

# Check if at least one <pkg-ref> contains the correct version
if ! grep -q "<pkg-ref id=\"com\.danielhoward314\.packet-sentry-agent\" version=\"$VERSION\">" ./Distribution; then
  echo "Error: No <pkg-ref> entry with version $VERSION found."
  exit 1
fi

# Check if <product> contains the correct version
if ! grep -q "<product version=\"$VERSION\"/>" ./Distribution; then
  echo "Error: <product> entry does not match version $VERSION."
  exit 1
fi

echo "Distribution contains correct version $VERSION in <pkg-ref> and <product>."

# When expanded, the outer pkg (packet-sentry-agent.pkg) has an agent.pkg that is a directory.
# A strange quirk of Apple's installer format.
cd ./agent.pkg

# Ensures the Scripts directory has only the expected scripts and that their checksums match source.
validate_scripts

echo "Comparing Payload against Bill of Materials"
rm -rf extracted_payload
mkdir extracted_payload && cd extracted_payload
cpio -i < ../Payload
cd ..

mkbom extracted_payload/ generatedbom
generated_bom_contents="$(lsbom -s generatedbom)"
actual_bom_contents="$(lsbom -s Bom)"
if [[ "$generated_bom_contents" != "$actual_bom_contents" ]]; then
    echo "Error: generated Bill of Materials does not match actual BOM"
    exit 1
fi

# Get back to macos-installer/package directory
cd ../..
rm -rf expanded
popd

echo "Successfully validated macOS installer pkg."
