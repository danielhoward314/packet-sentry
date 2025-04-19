param (
    [string]$GOARCH
)

# Define an array to hold the name of each successfully built executable
$BUILD_ARRAY = New-Object 'System.Collections.Generic.List[string]'

$SCRIPTS_DIR = (Get-Item -LiteralPath $MyInvocation.MyCommand.Path).DirectoryName
$ROOT_DIR = (Get-Item -Path (Join-Path $SCRIPTS_DIR "..")).FullName

# Clean previous builds
Remove-Item -Recurse -Force "$ROOT_DIR\build" -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "$ROOT_DIR\build" | Out-Null

function build_for_target {
    param (
        [string]$GOARCH
    )
    $GOOS = "windows"
    $EXECUTABLE_NAME = "packet_sentry_${GOOS}_${GOARCH}.exe"
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH

    Write-Host "Downloading go dependencies with 'go mod download'..."
    go mod download

    $CommitHash = (git rev-parse --short HEAD).Trim()
    $BuildTime = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
    $LDFLAGS = "-w -s -buildmode=exe -X `"main.Version=1.0.0`" -X `"main.CommitHash=$CommitHash`" -X `"main.BuildTime=$BuildTime`""

    Write-Host "Building executable: $EXECUTABLE_NAME for $GOOS $GOARCH..."
    $buildResult = $env:CGO_ENABLED="1"; & "go" "build" "-trimpath" "-ldflags" $LDFLAGS "-o" "$ROOT_DIR\build\$EXECUTABLE_NAME" "$ROOT_DIR\cmd\agent"

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed for GOOS=$GOOS GOARCH=$GOARCH"
        Write-Host "Error details:`n$buildResult"
        exit 1
    }

    $BUILD_ARRAY.Add("$ROOT_DIR\build\$EXECUTABLE_NAME")
    Write-Host "Build succeeded: $EXECUTABLE_NAME"

    $INSTALL_ACTIONS_NAME = "install_actions_${GOOS}_${GOARCH}.exe"
    Write-Host "Building executable: $INSTALL_ACTIONS_NAME for $GOOS $GOARCH..."
    $buildResult = & "go" "build" "-o" "$ROOT_DIR\build\$INSTALL_ACTIONS_NAME" "$ROOT_DIR\cmd\installeractions"

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed for GOOS=$GOOS GOARCH=$GOARCH"
        Write-Host "Error details:`n$buildResult"
        exit 1
    }

    $BUILD_ARRAY.Add("$ROOT_DIR\build\$INSTALL_ACTIONS_NAME")
    Write-Host "Build succeeded: $INSTALL_ACTIONS_NAME"
}

function validate_executable_format {
    param (
        [string]$EXECUTABLE
    )

    $BASE_NAME = [System.IO.Path]::GetFileNameWithoutExtension($EXECUTABLE)
    $OS_ARCH = $BASE_NAME -split "_"

    $GOOS = $OS_ARCH[2]
    $GOARCH = $OS_ARCH[3]

    if (-not $GOOS -or -not $GOARCH) {
        Write-Host "Error: Invalid executable format: $EXECUTABLE"
        exit 1
    }

    Write-Host "Validating executable: $BASE_NAME (GOOS=$GOOS, GOARCH=$GOARCH)"

    if (-not (Test-Path $EXECUTABLE)) {
        Write-Host "Error: $EXECUTABLE not found."
        exit 1
    }

    $fileInfo = Get-Item $EXECUTABLE
    if ($fileInfo.Extension -ne ".exe") {
        Write-Host "Error: $EXECUTABLE is not a valid Windows executable."
        exit 1
    }

    $binaryContent = [System.IO.File]::ReadAllBytes($EXECUTABLE)

    if ($binaryContent[0] -ne 0x4D -or $binaryContent[1] -ne 0x5A) {
        Write-Host "Error: $EXECUTABLE is not a valid PE file."
        exit 1
    }

    $peHeaderOffset = [BitConverter]::ToInt32($binaryContent, 0x3C)

    if ($binaryContent[$peHeaderOffset] -ne 0x50 -or
        $binaryContent[$peHeaderOffset + 1] -ne 0x45 -or
        $binaryContent[$peHeaderOffset + 2] -ne 0x00 -or
        $binaryContent[$peHeaderOffset + 3] -ne 0x00) {
        Write-Host "Error: Invalid PE signature."
        exit 1
    }

    $machineType = [BitConverter]::ToUInt16($binaryContent, $peHeaderOffset + 4)

    $arch = switch ($machineType) {
        0x8664 { "x86-64 (AMD64)" }
        0xAA64 { "ARM64" }
        0x014C { "x86 (32-bit)" }
        default { "Unknown" }
    }

    if ($GOARCH -eq "amd64" -and $machineType -ne 0x8664) {
        Write-Host "Error: Expected AMD64 binary, found $arch."
        exit 1
    }
    if ($GOARCH -eq "arm64" -and $machineType -ne 0xAA64) {
        Write-Host "Error: Expected ARM64 binary, found $arch."
        exit 1
    }

    Write-Host "Validation successful: $EXECUTABLE is a valid $arch executable."
}

if (-not $GOARCH) {
    Write-Host "Must specify architecture -GOARCH <amd64|arm64>"
    exit 1
} elseif ($GOARCH -in @("amd64", "arm64")) {
    build_for_target $GOARCH
} else {
    Write-Host "Invalid architecture: $GOARCH"
    Write-Host "Usage: .\scripts\build_agent.ps1 [-GOARCH <amd64|arm64>]"
    exit 1
}

if ($BUILD_ARRAY.Count -eq 0) {
    Write-Host "No builds completed."
    exit 1
}

foreach ($EXECUTABLE in $BUILD_ARRAY) {
    validate_executable_format $EXECUTABLE
}

Write-Host "Build process completed successfully!"
