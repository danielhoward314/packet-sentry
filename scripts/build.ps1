# Define an array to hold the name of each executable that builds successfully
$BUILD_ARRAY = @()

# Set the root directory to the parent directory of the script
$ROOT_DIR = (Get-Item -LiteralPath $MyInvocation.MyCommand.Path).DirectoryName

# Clean previous builds
Remove-Item -Recurse -Force "$ROOT_DIR\build" -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "$ROOT_DIR\build"

function build_for_target {
    param(
        [string]$GOOS,
        [string]$GOARCH
    )
    $EXECUTABLE_NAME = "packet_sentry_${GOOS}_${GOARCH}"
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $buildResult = & "go" "build" "-o" "$ROOT_DIR\build\$EXECUTABLE_NAME" "$ROOT_DIR\..."
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed for GOOS=${GOOS} GOARCH=${GOARCH}"
        exit 1
    }
    $BUILD_ARRAY += "$ROOT_DIR\build\$EXECUTABLE_NAME"
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
        Write-Host "Error: unable to get executable format and architecture to check from file name: $EXECUTABLE"
        exit 1
    }

    Write-Host "Validating executable: $BASE_NAME with GOOS=$GOOS, GOARCH=$GOARCH"

    if ($GOOS -eq "windows") {
        Write-Host "Validation skipped for Windows binaries."
        return
    }

    # Validate for non-Windows
    $fileOutput = & "file" $EXECUTABLE
    if ($fileOutput -notmatch "PE32" -and $fileOutput -notmatch "PE32+"){
        Write-Host "Error: $EXECUTABLE is not a valid Windows binary"
        exit 1
    }
    
    if ($GOARCH -eq "amd64" -and $fileOutput -notmatch "x86-64") {
        Write-Host "Error: $EXECUTABLE is not a valid x86-64 Windows binary"
        exit 1
    }
    if ($GOARCH -eq "arm64" -and $fileOutput -notmatch "ARM64") {
        Write-Host "Error: $EXECUTABLE is not a valid ARM64 Windows binary"
        exit 1
    }
}

function usage {
    Write-Host "Usage: $($MyInvocation.MyCommand.Name) [GOOS] [GOARCH]"
    Write-Host "  GOOS: The target operating system <windows>."
    Write-Host "  GOARCH: The target architecture <amd64|arm64>."
    Write-Host "If no arguments are provided, builds all targets."
    Write-Host "If both arguments are provided, builds only for the specified target."
    Write-Host "Example: $($MyInvocation.MyCommand.Name) windows amd64"
}

# Validate input arguments
if ($args.Count -eq 0) {
    Write-Host "Building for all targets..."
    build_for_target "windows" "amd64"
    build_for_target "windows" "arm64"
} elseif ($args.Count -eq 2) {
    if ($args[0] -ne "windows") {
        Write-Host "Invalid GOOS: $($args[0])"
        Write-Host "Only 'windows' is supported in this PowerShell script."
        exit 1
    }

    if ($args[1] -ne "amd64" -and $args[1] -ne "arm64") {
        Write-Host "Invalid GOARCH: $($args[1])"
        usage
        exit 1
    }

    Write-Host "Building for GOOS=$($args[0]) GOARCH=$($args[1])..."
    build_for_target $args[0] $args[1]
} else {
    Write-Host "Invalid number of arguments."
    usage
    exit 1
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
