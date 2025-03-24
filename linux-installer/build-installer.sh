#!/usr/bin/env bash

set -euo pipefail

# Expect to run this build script from root directory of the project
ROOT_DIR=$(pwd)

usage() {
  echo "Usage: $0 [version] [arch]"
  echo "  version: The version of the build."
  echo "  arch: The architecture of the build <amd64|arm64>."
  echo "Example: $0 1.0.0 amd64"
}

if [ $# -ne 2 ]; then
    echo "Error: incorrect number of arguments."
    usage
    exit 1
fi

VERSION="$1"
ARCH="$2"
PACKAGE_NAME="packet-sentry-agent"
INSTALL_DIR="/opt/packet-sentry"
BOOTSTRAP_FILE="${INSTALL_DIR}/agentBootstrap.json"

echo "Building Linux installer for version: ${VERSION}, architecture: ${ARCH}"

# Set architecture mappings
if [[ "$ARCH" == "amd64" ]]; then
    DEB_ARCH="amd64"
    RPM_ARCH="x86_64"
    BIN_FILE="$ROOT_DIR/build/packet_sentry_linux_amd64"
elif [[ "$ARCH" == "arm64" ]]; then
    DEB_ARCH="arm64"
    RPM_ARCH="aarch64"
    BIN_FILE="$ROOT_DIR/build/packet_sentry_linux_arm64"
else
    echo "Unsupported architecture: $ARCH"
    usage
    exit 1
fi

# Clean previous builds
rm -rf "$ROOT_DIR/linux-installer/build/"
mkdir -p "$ROOT_DIR/linux-installer/build/$INSTALL_DIR/bin"
mkdir -p "$ROOT_DIR/linux-installer/build/etc/systemd/system"

# Copy the Linux Go build for the ARCH
cp -f "$BIN_FILE" "$ROOT_DIR/linux-installer/build/$INSTALL_DIR/bin/packet-sentry-agent"
chmod +x "$ROOT_DIR/linux-installer/build/$INSTALL_DIR/bin/packet-sentry-agent"

# Copy the systemd service file if available
if [[ -f "$ROOT_DIR/linux-installer/package/packet-sentry-agent.service" ]]; then
    cp -f "$ROOT_DIR/linux-installer/package/packet-sentry-agent.service" "$ROOT_DIR/linux-installer/build/etc/systemd/system/packet-sentry-agent.service"
fi

# Package as .deb
DEB_DIR="$ROOT_DIR/linux-installer/deb-${ARCH}"
mkdir -p "$DEB_DIR/DEBIAN"

chmod +x "$DEB_DIR/DEBIAN/postinst"

chmod +x "$DEB_DIR/DEBIAN/prerm"

mkdir -p "$DEB_DIR/$INSTALL_DIR/bin"
mkdir -p "$DEB_DIR/etc/systemd/system"
cp -r "$ROOT_DIR/linux-installer/build/"* "$DEB_DIR/"

# Build the .deb package
dpkg-deb --build "$DEB_DIR"
mv "$DEB_DIR.deb" "$ROOT_DIR/linux-installer/${PACKAGE_NAME}_${VERSION}_${DEB_ARCH}.deb"

echo "DEB package built: $ROOT_DIR/linux-installer/${PACKAGE_NAME}_${VERSION}_${DEB_ARCH}.deb"

# Package as .rpm
RPM_DIR="$ROOT_DIR/linux-installer/rpm-${ARCH}"
mkdir -p "$ROOT_DIR/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}"
mkdir -p "$RPM_DIR/$INSTALL_DIR/bin"
mkdir -p "$RPM_DIR/etc/systemd/system"
cp -r "$ROOT_DIR/linux-installer/build/"* "$RPM_DIR/"

tar -czf "$ROOT_DIR/rpmbuild/SOURCES/${PACKAGE_NAME}-${VERSION}.tar.gz" -C "$RPM_DIR" .

rpmbuild -ba "$ROOT_DIR/rpmbuild/SPECS/${PACKAGE_NAME}.spec"

mv "$ROOT_DIR/rpmbuild/RPMS/${RPM_ARCH}/${PACKAGE_NAME}-${VERSION}-1.${RPM_ARCH}.rpm" "$ROOT_DIR/linux-installer/${PACKAGE_NAME}_${VERSION}_${RPM_ARCH}.rpm"

echo "RPM package built: $ROOT_DIR/linux-installer/${PACKAGE_NAME}_${VERSION}_${RPM_ARCH}.rpm"
