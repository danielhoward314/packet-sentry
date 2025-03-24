# Linux Commands

## Build the Go executable

```bash
./scripts/build linux <amd64|arm64>
```

## Build Linux the .deb Installer

```bash
./linux-installer/build-installer.sh <version> <amd64|arm64>
```

Depends on `dpkg-deb`.
