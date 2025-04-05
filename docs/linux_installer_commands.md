# Linux Installer Commands

## Build the Go executable

Depends on `libpcap-dev` for libpcap header files.

```bash
# Ubuntu
sudo apt-get install libpcap-dev

# Fedora or similar
sudo dnf install libpcap-devel
```

```bash
./scripts/build linux <amd64|arm64>
```

## Installer Pre-requisites

While packages `libpcap-dev`/`libpcap-devel` fulfill the build-time dependencies, there is still a runtime dependency on libpcap:

```bash
# Ubuntu
sudo apt-get install libpcap0.8

# Fedora or similar
sudo dnf install libpcap
```

## .deb Installer

### Build the .deb Installer

Depends on `dpkg-deb`.

```bash
go run ./linux-installer/main.go <version> <amd64|arm64> deb
```

This Go program creates a `./linux-installer/build/deb` directory with the hierarchy of folders required. The hierarchy looks like the following:

```bash
tree linux-installer/build/
linux-installer/build/
└── deb
    ├── DEBIAN
    │   ├── control
    │   ├── postinst
    │   └── prerm
    ├── etc
    │   └── systemd
    │       └── system
    │           └── packet-sentry-agent.service
    └── opt
        └── packet-sentry
            └── bin
                └── packet-sentry-agent

8 directories, 5 files
```

The Go programs copies the systemd service file into `etc/systemd/system`, the binary into `/opt/packet-sentry/bin`, and it fills out the dynamic data for the templates in `./linux-installer/deb-templates` and outputs the resulting files in the `./linux-installer/build/debfinalout` directory.

### Use the .deb installer

```bash
sudo dpkg -i ./linux-installer/build/debfinalout/packet-sentry-agent_1.0.0_amd64.deb
```

### Debug the .deb Installer

```bash
# show contents (1.0.0 version and built for amd64)
dpkg-deb -c ./linux-installer/build/debfinalout/packet-sentry-agent_1.0.0_amd64.deb

# if well-formed, this should show the contents of the control file
dpkg-deb --info ./linux-installer/build/debfinalout/packet-sentry-agent_1.0.0_amd64.deb
dpkg-deb --ctrl-tarfile ./linux-installer/build/debfinalout/packet-sentry-agent_1.0.0_amd64.deb | tar -tvf -

sudo systemctl status packet-sentry-agent
sudo journalctl -u packet-sentry-agent -f

sudo ls -l /opt/packet-sentry/
sudo ls -l /opt/packet-sentry/bin
```

### Clean up

```bash
rm -rf ./linux-installer/build
rm -rf /opt/packet-sentry/

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

## .rpm Installer

### Background

Since I found it somewhat more challenging to build the installer as an RPM package, I've included this section with what I had to learn to do it correctly.

Here is a good [Red Hat video](https://www.youtube.com/live/WVSEzg8E_wg?si=8GJwNmf3UkG15vmV) that provides a primer on building RPM packages.

```bash
dnf -y install rpmdevtools rpmlint
```

Before moving on to building our RPM, we can inspect the contents of these RPMs we just downloaded:

```bash
dnf download --source rpmdevtools
dnf download --source rpmlint
# Next command will list the spec file, any uncompressed/unarchived contents, plus a tarball
rpm -qpl <package-name>-<version>-<release>.<distro>.src.rpm
rpm2cpio <package-name>-<version>-<release>.<distro>.src.rpm | cpio -idmv
# Previous command should output the spec file and tarball
cat <package-name>.spec
tar -xJvf <package-name-tarball>.tar.xz

```

```bash
echo '%packager First Last <firstlast@email.com>' >> ~/.rpmmacros
rpmdev-setuptree
```

Use `tree` to see the directory structure the last command set up:

```bash
cd rpmbuild
tree
.
├── BUILD
├── RPMS
├── SOURCES
├── SPECS
└── SRPMS

6 directories, 0 files
```

The Go program that builds the .rpm installer does the following:

- copies the systemd service file and the Go executable into the `SOURCES` directory
- fills out the dynamic parts of the `./linux-installer/rpm-templates/packet-sentry-agent.spec.tmpl` and copies it into the `SPECS` directory
- executes `rpmbuild` which should put the final .rpm nested under the `RPMS` directory by architecture
- copies the final .rpm to `./linux-installer/rpmfinalout`

### Build the .rpm Installer

Depends on `rpmbuild`.

```bash
go run ./linux-installer/main.go <version> <amd64|arm64> rpm
```

### Use the .rpm Installer

```bash
sudo dnf install ./linux-installer/build/rpmfinalout/<rpm>
```

__Note__: unlike the .deb installer, which prompts for the install key during installation, the .rpm installer requires you to run `/opt/packet-sentry/setup.sh` after installing to lay down the install key.

### Clean up

```bash
sudo dnf remove packet-sentry-agent

# if the above doesn't work, these commands can manually undo the install
sudo systemctl stop packet-sentry-agent.service
sudo systemctl disable packet-sentry-agent.service
sudo rm -f /etc/systemd/system/packet-sentry-agent.service
sudo systemctl daemon-reload
sudo rm -rf /opt/packet-sentry
```