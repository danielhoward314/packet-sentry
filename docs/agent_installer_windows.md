# Windows Installer Commands

## Build the Go executable

Pre-requisites:

- git

```PowerShell
.\scripts\build_agent.ps1 -GOARCH <amd64|arm64>
```

## Pre-requisites for Building the MSI Installer

Install `winget` package manager, either from the Microsoft Store or using PowerShell. From Microsoft documentation:

> The WinGet command line tool is only supported on Windows 10 1709 (build 16299) or later at this time. WinGet will not be available until you have logged into Windows as a user for the first time, triggering Microsoft Store to register the Windows Package Manager as part of an asynchronous process. If you have recently logged in as a user for the first time and find that WinGet is not yet available, you can open PowerShell and enter the following command to request this WinGet registration: `Add-AppxPackage -RegisterByFamilyName -MainPackage Microsoft.DesktopAppInstaller_8wekyb3d8bbwe`.

Install Visual Studio and the .NET SDK 9.0; must be done in an Administrator PowerShell session.

```PowerShell
winget install --id=Microsoft.VisualStudio.2022.Community  -e
winget install Microsoft.DotNet.SDK.9
winget install --id=WiXToolset.WiXToolset -e
```

Restart the PowerShell session.

Follow the WiX quick start documentation [here](https://docs.firegiant.com/quick-start/) (takes 5 minutes to try it) to get a feel for creating an MSI.

Useful commands from the WiX documentation to build, install, and uninstall:

```PowerShell
dotnet build
msiexec /i bin\Debug\QuickStart.msi /l*v install_log.txt
msiexec /x bin\Debug\QuickStart.msi /l*v uninstall_log.txt
```

## WiX/MSI Development

The [wixedit](https://wixedit.github.io/) GUI tool is useful while editing the WiX XML. It shows the UI and install execution sequences, you can see what the dialogs actually look like, and it has the metadata about custom actions.

This project's `Product.wxs` is pretty simple and borrows heavily from the WiX samples. The WiX XML is a declarative expression of the desired state of:

- installed directories (`<Directory>.../<Directory>`)
- files (`<Component><File>...</File></Component>`)
- a sequence of UI elements as users progress through the dialogs (`<InstallUISequence>...</InstallUISequence>`)
- an install sequence where you can fire custom actions (`<InstallExecuteSequence>...</InstallExecuteSequence>`)
- the appearance of UI elements (`<UI>...</UI>`) which can also define additional sequencing and publish properties for use by custom actions

The `Id` attribute of XML is used to connect things in the files:

- the `Component` and `ComponentRef` share the `PacketSentryServiceComponent` value for their `Id`
- the `Binary` has the `InstallActions` value for its `Id` that ties it to the `BinaryKey` of the `CustomAction`

The `ServiceInstall` and `ServiceControl` are WiX built-ins for defining a Windows Service that can be managed with `sc.exe`.

The custom actions have the property `Execute="immediate"` or `Execute="deferred"` where the former doesn't have elevations and the latter do elevate.

The `<Property Id="INSTALLKEY" Secure="yes" />` is published by the `Control` within the `PacketSentryInstallKeyDlg` dialog. The `CustomActionWriteInstallKey` custom action uses the property.

The installer build script (`.\windows-installer\build-installer.ps1`) invokes `candle.exe` and `light.exe` from the WiX Toolset. The former compiles the `.wxs` files into `.wixobj` that `light.exe` uses to build the MSI. You can pass arguments into `candle.exe` with syntax like `-dVersion="${version}"` which become available in the `.wxs` files as variables for dynamic values with syntax like `$(var.Version)`.

The GUIDs are generated:

```PowerShell
[guid]::NewGuid()
```

## Build the Windows MSI Installer

```PowerShell
# Requires an admin session. Also, you may need to run cmdlet `Set-ExecutionPolicy <PolicyValue>` to be able execute a script.
.\windows-installer\build-installer.ps1 -arch amd64 -version 2.3.4
```

The error `light.exe : warning LGHT1032 : Unable to reset acls on destination files.` can happen if you're working on the files within a shared folder between host and guest Windows VM; copy the repo to a location on the guest Windows system.

## Inspect the Windows MSI Installer (on Windows)

The SDK has some tools for this. Download [here](https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/).
Note the installation path in the install wizard (e.g. `C:\Program Files (x86)\Windows Kits\10\`). Orca is a GUI-based tool for inspecting MSIs; you have to install it from a directory within the SDK, similar to a path like `"C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\x86\Orca-x86_en-us.msi"`.

```PowerShell
 msiexec /a "<absolute-path-to-root-of-project>\windows-installer\PacketSentryInstaller_amd64_v1.0.0.msi"  /qb TARGETDIR=C:\extracted

Get-ChildItem -Path C:\extracted -Filter "*.exe" -Recurse
Get-AuthenticodeSignature <msi>
```

## Inspect the Windows MSI Installer (on Unix)

Install `msitools` and `cabextract`.

```bash
# Ubuntu
sudo apt update
sudo apt install msitools
sudo apt install cabextract

# macOS
brew install msitools
brew install cabextract
```

Inspect the .msi:

```bash
msiinfo tables <msi>
msiinfo streams <msi>
mkdir msiextracted
msiextract --directory msiextracted/
# should have directory structure that would be expected upon install: /'Program Files'/PacketSentry/packet_sentry.exe
ls -R msiextracted
#   -t, --tables -s, --streams -S, --signature -d DIR  Dump to given directory DIR
msidump -t -s -S -d msidumped <msi>
ls -R msidumped
mkdir cabextracted
cabextract -d cabextracted msidumped/_Streams/packet_sentry.cab
# similar output just doing it on the msi directly
cabextract -l <msi>
```

## Install the Windows MSI

Pre-requisite:

- npcap is installed. Downloads available through [here](https://npcap.com/#download).

Double click it, or use msiexec:

```PowerShell
msiexec /i "C:\Users\dhoward\Desktop\scratch\windows-installer\PacketSentryInstaller_amd64_v1.0.0.msi"
```

## Debug the installation

Use msiexec with verbose logging written to a log file:

```PowerShell
msiexec /i "<absolute-path-to-root-of-project>\windows-installer\PacketSentryInstaller_amd64_v1.0.0.msi" /L*v "<absolute-path-to-root-of-project>\install.log"
```

## Clean up

```PowerShell
rm .\install.log

rm .\windows-installer\Product.wixobj

rm .\windows-installer\PacketSentryInstaller_*.msi
```
