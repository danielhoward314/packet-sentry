# Define an array to hold the name of each executable that builds successfully
$BUILD_ARRAY = New-Object 'System.Collections.Generic.List[string]'

$SCRIPTS_DIR = (Get-Item -LiteralPath $MyInvocation.MyCommand.Path).DirectoryName
$ROOT_DIR = (Get-Item -Path (Join-Path $SCRIPTS_DIR "..")).FullName

# Clean previous builds
Remove-Item -Recurse -Force "$ROOT_DIR\build" -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "$ROOT_DIR\build"

function build_for_target {
    param(
        [string]$GOOS,
        [string]$GOARCH
    )
    $EXECUTABLE_NAME = "packet_sentry_${GOOS}_${GOARCH}.exe"
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    Write-Host "Building executable named $EXECUTABLE_NAME for $env:GOOS $env:GOARCH."
    $buildResult = & "go" "build" "-o" "$ROOT_DIR\build\$EXECUTABLE_NAME" "$ROOT_DIR\..."
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed for GOOS=${GOOS} GOARCH=${GOARCH}"
        Write-Host "Error details:`n$buildResult"
        exit 1
    }
    $BUILD_ARRAY.Add("$ROOT_DIR\build\$EXECUTABLE_NAME")
    Write-Host "Build succeeded for GOOS=${GOOS} GOARCH=${GOARCH}. Executable: $EXECUTABLE_NAME"
}

function validate_executable_format {
    param (
        [string]$EXECUTABLE
    )

    $BASE_NAME = [System.IO.Path]::GetFileName($EXECUTABLE)
    $OS_ARCH = $BASE_NAME -split "_"

    $GOOS = $OS_ARCH[2]
    $GOARCH = $OS_ARCH[3]

    if (-not $GOOS -or -not $GOARCH) {
        Write-Host "Error: unable to get expected executable format and architecture from file name: $EXECUTABLE"
        exit 1
    }

    Write-Host "Validating executable: $BASE_NAME with GOOS=$GOOS, GOARCH=$GOARCH"

    if (-not (Test-Path $EXECUTABLE)) {
        Write-Host "Error: $EXECUTABLE does not exist."
        exit 1
    }

    $fileInfo = Get-Item $EXECUTABLE
    if ($fileInfo.Extension -ne ".exe") {
        Write-Host "Error: $EXECUTABLE is not an executable file."
        exit 1
    }

    $binaryContent = [System.IO.File]::ReadAllBytes($EXECUTABLE)

    # PE files always start with "MZ" (0x4D 0x5A)
    if ($binaryContent[0] -ne 0x4D -or $binaryContent[1] -ne 0x5A) {
        Write-Host "Error: $EXECUTABLE is not a valid Windows PE file."
        exit 1
    }

    # The PE header offset is located at byte 0x3C (60) in the DOS header
    $peHeaderOffset = [BitConverter]::ToInt32($binaryContent, 0x3C)

    # The PE signature should be "PE\0\0" (0x50 0x45 0x00 0x00)
    if ($binaryContent[$peHeaderOffset] -ne 0x50 -or
        $binaryContent[$peHeaderOffset + 1] -ne 0x45 -or
        $binaryContent[$peHeaderOffset + 2] -ne 0x00 -or
        $binaryContent[$peHeaderOffset + 3] -ne 0x00) {
        Write-Host "Error: Invalid PE signature."
        exit 1
    }

    # The machine type is located at offset 0x4 (4 bytes after PE signature)
    $machineType = [BitConverter]::ToUInt16($binaryContent, $peHeaderOffset + 4)

    $arch = switch ($machineType) {
        0x8664 { "x86-64 (AMD64)" }
        0xAA64 { "ARM64" }
        0x014C { "x86 (32-bit)" }
        default { "Unknown" }
    }

    if ($GOARCH -eq "amd64" -and $machineType -ne 0x8664) {
        Write-Host "Error: $EXECUTABLE is not a valid x86-64 (AMD64) binary. Found $arch machine type."
        exit 1
    }
    if ($GOARCH -eq "arm64" -and $machineType -ne 0xAA64) {
        Write-Host "Error: $EXECUTABLE is not a valid ARM64 binary. Found $arch machine type."
        exit 1
    }

    Write-Host "Validation successful: $EXECUTABLE is a valid $arch Windows executable."
}

function usage {
    Write-Host "Usage: .\scripts\build.ps1 [GOOS] [GOARCH]"
    Write-Host "  GOOS: The target operating system <windows>."
    Write-Host "  GOARCH: The target architecture <amd64|arm64>."
    Write-Host "Example: .\scripts\build.ps1 windows amd64"
}

# Validate input arguments
if ($args.Count -ne 2) {
    Write-Host "Invalid number of arguments."
    usage
    exit 1
} else {
    if ($args[0] -ne "windows") {
        Write-Host "Invalid GOOS: $($args[0])"
        Write-Host "Use scripts/build for $($args[0])."
        exit 1
    }

    if ($args[1] -ne "amd64" -and $args[1] -ne "arm64") {
        Write-Host "Invalid GOARCH: $($args[1])"
        usage
        exit 1
    }

    Write-Host "Building for GOOS=$($args[0]) GOARCH=$($args[1])..."
    build_for_target $args[0] $args[1]
}

if ($BUILD_ARRAY.Count -eq 0) {
    Write-Host "No builds were completed."
    exit 1
}

# Validate built executables
foreach ($EXECUTABLE in $BUILD_ARRAY) {
    validate_executable_format $EXECUTABLE
}

Write-Host "Build process completed."
