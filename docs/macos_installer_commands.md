# macOS Installer Commands

## Build the Go executable

```bash
./scripts/build_agent darwin <amd64|arm64>
```

## Build macOS Installer

```bash
./macos-installer/build-installer.sh <version> <amd64|arm64>
```

## Validate macOS Installer

```bash
./macos-installer/validate-installer.sh <version> <amd64|arm64>
```

## Inspect the macOS installer pkg

See the logic of the `validate-installer.sh` script for a more complete picture.

```bash
cd ./macos-installer/package
pkgutil --expand packet-sentry-agent.pkg expanded
cd expanded
cat Distribution
ls Resources
ls agent.pkg
cd agent.pkg
lsbom Bom
cat PackageInfo
ls Scripts
cpio -i < Payload
ls
cd opt/packet-sentry
ls
cat com.danielhoward314.packet-sentry-agent.plist
ls bin/
cd ../../../..
```

## Install the macOS installer pkg

```bash
# this method bypasses the code signing and notarization checks
sudo installer -pkg ./macos-installer/package/packet-sentry-agent.pkg -target /
```

## Debug the macOS installation

```bash
less +F /var/log/install.log
less +F /var/log/packet-sentry-preinstall.log
less +F /var/log/packet-sentry-postinstall.log

sudo ls /opt/packet-sentry

sudo launchctl list com.danielhoward314.packet-sentry-agent
cat /var/run/packetsentryagent.pid
ps -p <pid> -o pid,uid,gid,%cpu,%mem,etime,args,args
sudo lsof -p <pid>
```

## Clean up the macOS installation

```bash
rm macos-installer/package/agent.pkg macos-installer/package/packet-sentry-agent.pkg

# if `pkgutil --expand packet-sentry-agent.pkg expanded` was done
rm -rf ./macos-installer/package/expanded/ 

sudo launchctl list com.danielhoward314.packet-sentry-agent
# list command should have pid of daemon
sudo kill -9 <pid>
sudo launchctl unload /Library/LaunchDaemons/com.danielhoward314.packet-sentry-agent.plist

sudo rm -rf /opt/packet-sentry/
```