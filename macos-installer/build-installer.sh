#!/usr/bin/env bash

set -euo pipefail

# Expect to run this build script from root directory of the project
ROOT_DIR=$(pwd)

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
echo "building installer for architecture ${ARCH} and version ${VERSION}"

# Clean previous builds.
rm -rf "$ROOT_DIR/macos-installer/build/"
mkdir -p "$ROOT_DIR/macos-installer/build/opt/packet-sentry/bin"

# Copy the darwin Go build for the ARCH to where we need for `pkgbuild` and `productbuild`.
if [[ "$ARCH" == "amd64" ]]; then
    cp -f "$ROOT_DIR/build/packet_sentry_darwin_amd64" "$ROOT_DIR/macos-installer/build/opt/packet-sentry/bin/packet-sentry-agent"
elif [[ "$ARCH" == "arm64" ]]; then
    cp -f "$ROOT_DIR/build/packet_sentry_darwin_arm64" "$ROOT_DIR/macos-installer/build/opt/packet-sentry/bin/packet-sentry-agent"
else
    echo "Unsupported architecture: $ARCH"
    usage
    exit 1
fi

cp -f "$ROOT_DIR/macos-installer/package/com.danielhoward314.packet-sentry-agent.plist" "$ROOT_DIR/macos-installer/build/opt/packet-sentry/com.danielhoward314.packet-sentry-agent.plist"

chmod +x "$ROOT_DIR/macos-installer/build/opt/packet-sentry/bin/packet-sentry-agent"

pushd "$ROOT_DIR/macos-installer/package"

# Copy the template to distribution.xml
cp distribution.xml.template distribution.xml
# Use sed to replace hard-coded version in the template with the actual version
sed -i '' -E "s/(<pkg-ref id=\"com\.danielhoward314\.packet-sentry-agent\" version=\")([^\"]+)(\">)/\1$VERSION\3/" distribution.xml
sed -i '' -E "s/(<product version=\")([^\"]+)(\"\/>)/\1$VERSION\3/" distribution.xml
echo "Generated distribution.xml with version $VERSION"

PKG_NAME="packet-sentry-agent_${VERSION}_${ARCH}.pkg"

pkgbuild \
  --root ../build \
  --identifier com.danielhoward314.packet-sentry-agent \
  --scripts scripts \
  --version "${VERSION}" \
  --ownership recommended \
  agent-${ARCH}.pkg

productbuild \
  --distribution "./distribution.xml" \
  --resources "./resources" \
  --version "${VERSION}" \
  "./package/${PKG_NAME}"

echo "Built ${PKG_NAME}"
popd

echo "Successfully built macOS installer pkg."