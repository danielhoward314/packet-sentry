param(
    [string]$arch = "amd64",
    [string]$version = "1.0.0"
)

# Map GOARCH architecture to WiX platform names
$platformMap = @{
    "amd64" = "x64"
    "arm64" = "arm64"
}

if (-not $platformMap.ContainsKey($arch)) {
    Write-Error "Invalid architecture. Use -arch amd64 or -arch arm64."
    exit 1
}

$platform = $platformMap[$arch]

# Define paths relative to the script's location
$installerDir = $PSScriptRoot
$projectRoot = (Resolve-Path "$installerDir\..").Path
$binDir = Join-Path -Path $installerDir -ChildPath "bin\$platform"
$buildDir = Join-Path -Path $projectRoot -ChildPath 'build'

Write-Host "installer dir: $installerDir"
Write-Host "projectRoot dir: $projectRoot"
Write-Host "bin dir: $binDir"
Write-Host "build dir: $buildDir"

# Ensure bin directory exists
if (!(Test-Path -Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

# Copy the main Go executable. Pre-requisite: `.\scripts\build_agent.ps1` has already run to build the Go executable.
$packetSentryExe = Get-Item "$buildDir\packet_sentry_windows_$arch.exe" -ErrorAction SilentlyContinue
if ($packetSentryExe) {
    Copy-Item $packetSentryExe.FullName "$binDir\packet_sentry.exe"
} else {
    Write-Error "No Packet Sentry executable found for $arch."
    exit 1
}

# Copy the install actions CLI Go. Pre-requisite: `.\scripts\build_agent.ps1` has already run to build the Go executable.
$install_actions = Get-Item "$buildDir\install_actions_windows_$arch.exe" -ErrorAction SilentlyContinue
if ($install_actions) {
    Copy-Item $install_actions.FullName "$binDir\install_actions.exe"
} else {
    Write-Error "No Install Actions executable found for $arch."
    exit 1
}

function Find-WiXToolset {
    $wixPaths = @(
        "C:\Program Files (x86)\WiX Toolset v3.14\bin",
        "C:\Program Files (x86)\WiX Toolset v3.11\bin",
        "C:\Program Files (x86)\WiX Toolset v3.10\bin",
        "C:\Program Files (x86)\WiX Toolset v3.9\bin"
    )

    foreach ($path in $wixPaths) {
        $candleExists = Test-Path "$path\candle.exe"
        $lightExists = Test-Path "$path\light.exe"
        if ($candleExists -and $lightExists) {
            return $path
        }
    }

    # If not found in common paths, check registry
    $registryPath = "HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall"
    $installedApps = Get-ChildItem $registryPath | ForEach-Object {
        Get-ItemProperty $_.PsPath
    }

    $wixInstall = $installedApps | Where-Object { $_.DisplayName -like "WiX Toolset*" }
    
    if ($wixInstall) {
        $installPath = $wixInstall.InstallLocation
        if ($installPath) {
            $candleExists = Test-Path "$installPath\bin\candle.exe"
            $lightExists = Test-Path "$installPath\bin\light.exe"
            if ($candleExists -and $lightExists) {
                return "$installPath\bin"
            }
        }
    }

    return $null
}


Write-Host "Checking for WiX via environment variable..."
if ($Env:WIX) {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    $NewPath = "$($Env:WIX)\bin;$currentPath"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "Machine")
    $wixBinPath = "$($Env:WIX)\bin"
} else {
    Write-Host "Environment variable WIX not set. Falling back to local detection."
    $wixBinPath = Find-WiXToolset
}

if ($wixBinPath) {
    $candleExe = "$wixBinPath\candle.exe"
    $lightExe = "$wixBinPath\light.exe"
} else {
    Write-Host "Error: WiX Toolset not found!"
    exit 1
}

Write-Host "WiX Toolset found at: $wixBinPath"

# Compile the WiX installer for the correct architecture and version
&$candleExe -dPlatform="${platform}" -dVersion="${version}" "$installerDir\Product.wxs" -o "$installerDir\Product.wixobj"

# Ensure the Product.wixobj file was created
$productWixObj = "$installerDir\Product.wixobj"
if (-Not (Test-Path $productWixObj)) {
    Write-Error "Failed to create Product.wixobj file: $productWixObj"
    exit 1
}

# Now run light.exe to generate the MSI
&$lightExe $productWixObj -o "$installerDir\PacketSentryInstaller_${arch}_v${version}.msi"
if (-Not (Test-Path "$installerDir\PacketSentryInstaller_${arch}_v${version}.msi")) {
    Write-Error "Failed to create MSI file"
    exit 1
}

Write-Host "Installer built successfully: $installerDir\PacketSentryInstaller_${arch}_v${version}.msi"
