# Linux Installer Commands

## Build the Go executable

```bash
./scripts/build linux <amd64|arm64>
```

## Build Linux the .deb Installer

```bash
go run ./linux-installer/main.go <version> <amd64|arm64>
```

Depends on `dpkg-deb`.

## Use the .deb installer

```bash
sudo dpkg -i ./linux-installer/packet-sentry-agent_1.0.0_amd64.deb
```

## Debug the .deb Installer

```bash
# show contents (1.0.0 version and built for amd64)
dpkg-deb -c ./linux-installer/packet-sentry-agent_1.0.0_amd64.deb

# if well-formed, this should show the contents of the control file
dpkg-deb --info ./linux-installer/packet-sentry-agent_1.0.0_amd64.deb

sudo systemctl status packet-sentry-agent
sudo journalctl -u packet-sentry-agent -f

sudo ls -l /opt/packet-sentry/
sudo ls -l /opt/packet-sentry/bin
```

## Clean up

```bash
rm -rf ./linux-installer/build
rm ./linux-installer/packet-sentry-agent_1.0.0_amd64.deb

# an error-free install usually means this next command is all that's needed for uninstall
# and the prerm script should invoke the systemctl commands for you
sudo dpkg -r packet-sentry-agent

# some bigger hammers for uninstall
sudo dpkg --purge packet-sentry-agent
sudo apt-get remove --purge packet-sentry-agent
sudo dpkg --remove --force-remove-reinstreq packet-sentry-agent
sudo apt-get autoremove
sudo apt-get clean

# if all else fails
ls -l /var/lib/dpkg/info | grep packet-sentry*
sudo mv /var/lib/dpkg/info/packet-sentry* /tmp/
sudo dpkg --remove --force-remove-reinstreq packet-sentry-agent
sudo systemctl stop packet-sentry-agent.service
sudo systemctl disable packet-sentry-agent.service
sudo rm -f /etc/systemd/system/packet-sentry-agent.service

# Reload systemd to ensure all services are cleared
systemctl daemon-reload
```