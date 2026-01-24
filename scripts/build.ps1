# EdgeFlow Build Script (PowerShell)
# Automated build for different profiles and platforms

param(
    [Parameter(Position=0)]
    [ValidateSet("minimal", "standard", "full", "all", "help")]
    [string]$Profile = "help",

    [Parameter(Position=1)]
    [ValidateSet("linux-arm64", "linux-arm", "linux-amd64", "darwin-arm64", "darwin-amd64", "windows-amd64", "all")]
    [string]$Platform = "linux-arm64"
)

# Configuration
$VERSION = if ($env:VERSION) { $env:VERSION } else { "0.1.0" }
$BUILD_DIR = "bin"
$BINARY_NAME = "edgeflow"

# Colors (limited support in PowerShell)
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Print banner
function Print-Banner {
    Write-ColorOutput "╔════════════════════════════════════════════════════════════════╗" "Cyan"
    Write-ColorOutput "║           EdgeFlow Build Script v$VERSION                      ║" "Cyan"
    Write-ColorOutput "╚════════════════════════════════════════════════════════════════╝" "Cyan"
    Write-Host ""
}

# Print help
function Print-Help {
    Print-Banner
    Write-ColorOutput "Usage:" "Green"
    Write-Host "  .\scripts\build.ps1 [profile] [platform]"
    Write-Host ""
    Write-ColorOutput "Profiles:" "Green"
    Write-Host "  minimal  - Pi Zero, BeagleBone (10MB, 50MB RAM)"
    Write-Host "  standard - Pi 3/4 (20MB, 200MB RAM)"
    Write-Host "  full     - Pi 4/5, Jetson (35MB, 400MB RAM)"
    Write-Host "  all      - Build all profiles"
    Write-Host ""
    Write-ColorOutput "Platforms:" "Green"
    Write-Host "  linux-arm64   - Raspberry Pi 4/5 (64-bit)"
    Write-Host "  linux-arm     - Raspberry Pi 3/Zero (32-bit)"
    Write-Host "  linux-amd64   - Linux x86_64"
    Write-Host "  windows-amd64 - Windows x86_64"
    Write-Host "  all           - Build for all platforms"
    Write-Host ""
    Write-ColorOutput "Examples:" "Green"
    Write-Host "  .\scripts\build.ps1 minimal linux-arm64"
    Write-Host "  .\scripts\build.ps1 standard all"
    Write-Host "  .\scripts\build.ps1 all all"
    Write-Host ""
}

# Build for specific profile and platform
function Build-Binary {
    param(
        [string]$Prof,
        [string]$Plat
    )

    $buildTags = ""
    $ldflags = ""
    $optimization = ""

    # Configure build based on profile
    switch ($Prof) {
        "minimal" {
            $buildTags = "minimal"
            $ldflags = "-w -s -X main.Version=$VERSION -X main.Profile=minimal"
            $optimization = "-trimpath"
        }
        "standard" {
            $buildTags = "standard,network,gpio,database"
            $ldflags = "-w -s -X main.Version=$VERSION -X main.Profile=standard"
            $optimization = "-trimpath"
        }
        "full" {
            $buildTags = "full,network,gpio,database,messaging,ai,industrial,advanced"
            $ldflags = "-w -X main.Version=$VERSION -X main.Profile=full"
            $optimization = ""
        }
        default {
            Write-ColorOutput "Error: Invalid profile '$Prof'" "Red"
            Print-Help
            exit 1
        }
    }

    # Configure platform
    $goos = ""
    $goarch = ""
    $goarm = ""
    $outputSuffix = ""
    $extension = ""

    switch ($Plat) {
        "linux-arm64" {
            $goos = "linux"
            $goarch = "arm64"
            $outputSuffix = "linux-arm64"
        }
        "linux-arm" {
            $goos = "linux"
            $goarch = "arm"
            $goarm = "7"
            $outputSuffix = "linux-arm"
        }
        "linux-amd64" {
            $goos = "linux"
            $goarch = "amd64"
            $outputSuffix = "linux-amd64"
        }
        "darwin-arm64" {
            $goos = "darwin"
            $goarch = "arm64"
            $outputSuffix = "darwin-arm64"
        }
        "darwin-amd64" {
            $goos = "darwin"
            $goarch = "amd64"
            $outputSuffix = "darwin-amd64"
        }
        "windows-amd64" {
            $goos = "windows"
            $goarch = "amd64"
            $outputSuffix = "windows-amd64"
            $extension = ".exe"
        }
        default {
            Write-ColorOutput "Error: Invalid platform '$Plat'" "Red"
            Print-Help
            exit 1
        }
    }

    # Output file name
    $outputFile = "$BUILD_DIR\$BINARY_NAME-$Prof-$outputSuffix$extension"

    Write-ColorOutput "Building $Prof profile for $Plat..." "Green"
    Write-ColorOutput "  Tags: $buildTags" "Cyan"
    Write-ColorOutput "  Output: $outputFile" "Cyan"

    # Create build directory
    New-Item -ItemType Directory -Force -Path $BUILD_DIR | Out-Null

    # Build command
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    if ($goarm) {
        $env:GOARM = $goarm
    }

    $buildCmd = "go build -tags `"$buildTags`" -ldflags `"$ldflags`" $optimization -o `"$outputFile`" .\cmd\edgeflow"

    try {
        Invoke-Expression $buildCmd
        Write-ColorOutput "✓ Build successful" "Green"
        Get-Item $outputFile | Format-Table Name, Length, LastWriteTime
        Write-Host ""
    }
    catch {
        Write-ColorOutput "✗ Build failed: $_" "Red"
        exit 1
    }
    finally {
        Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
        Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
        Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
    }
}

# Build all profiles for a platform
function Build-AllProfiles {
    param([string]$Plat)

    Write-ColorOutput "Building all profiles for $Plat..." "Cyan"
    Write-Host ""
    Build-Binary "minimal" $Plat
    Build-Binary "standard" $Plat
    Build-Binary "full" $Plat
}

# Build a profile for all platforms
function Build-AllPlatforms {
    param([string]$Prof)

    Write-ColorOutput "Building $Prof profile for all platforms..." "Cyan"
    Write-Host ""
    Build-Binary $Prof "linux-arm64"
    Build-Binary $Prof "linux-arm"
    Build-Binary $Prof "linux-amd64"
    Build-Binary $Prof "windows-amd64"
}

# Build everything
function Build-All {
    Write-ColorOutput "Building all profiles for all platforms..." "Cyan"
    Write-Host ""

    # Linux ARM64 (Pi 4/5)
    Build-Binary "minimal" "linux-arm64"
    Build-Binary "standard" "linux-arm64"
    Build-Binary "full" "linux-arm64"

    # Linux ARM32 (Pi 3/Zero)
    Build-Binary "minimal" "linux-arm"
    Build-Binary "standard" "linux-arm"
    Build-Binary "full" "linux-arm"

    # Linux AMD64 (development)
    Build-Binary "minimal" "linux-amd64"
    Build-Binary "standard" "linux-amd64"
    Build-Binary "full" "linux-amd64"

    # Windows AMD64 (development)
    Build-Binary "minimal" "windows-amd64"
    Build-Binary "standard" "windows-amd64"
    Build-Binary "full" "windows-amd64"

    Write-ColorOutput "╔════════════════════════════════════════════════════════════════╗" "Green"
    Write-ColorOutput "║                    Build Summary                               ║" "Green"
    Write-ColorOutput "╚════════════════════════════════════════════════════════════════╝" "Green"
    Write-Host ""
    Get-ChildItem $BUILD_DIR | Format-Table Name, Length, LastWriteTime
}

# Main script
Print-Banner

if ($Profile -eq "help") {
    Print-Help
    exit 0
}

if ($Profile -eq "all" -and $Platform -eq "all") {
    Build-All
}
elseif ($Profile -eq "all") {
    Build-AllProfiles $Platform
}
elseif ($Platform -eq "all") {
    Build-AllPlatforms $Profile
}
else {
    Build-Binary $Profile $Platform
}

Write-ColorOutput "✓ All builds completed successfully" "Green"
