# Windows Installer Commands

## Build the Go executable (Windows & Linux)

Windows:

```PowerShell
.\scripts\build.ps1 <amd64|arm64>
```

Linux (must have installed PowerShell first):

```bash
pwsh # should launch a new shell
```

```PowerShell
./scripts/build-windows-on-linux.ps1 <amd64|arm64>
```

## Install Pre-requisites

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

## Build the Windows MSI Installer


```PowerShell
# Requires an admin session and you may to run cmdlet `Set-ExecutionPolicy <PolicyValue>` to be able execute a script.
.\windows-installer\build-installer.ps1
```

The error `light.exe : warning LGHT1032 : Unable to reset acls on destination files.` can happen if you're working on the files within a shared folder between host and guest Windows VM; copy the repo to a location on the guest Windows system.

## Inspect the Windows MSI Installer (on Windows)

The SDK has some tools for this. Download [here](https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/).
Note the installation path in the install wizard (e.g. `C:\Program Files (x86)\Windows Kits\10\`). Only the MSI Tools are needed.

## Inspect the Windows MSI Installer (on Unix)

## Install the Windows MSI

Double click it.

## Debug the installation

```PowerShell
 msiexec /i "<absolute-path-to-root-of-project>\windows-installer\PacketSentryInstaller_amd64_v1.0.0.msi" /L*v "<absolute-path-to-root-of-project>\install.log"
```

## Clean up

```PowerShell
rm .\install.log

rm .\windows-installer\Product.wixobj

rm .\windows-installer\PacketSentryInstaller_*.msi
```