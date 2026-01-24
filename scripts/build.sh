#!/bin/bash

# EdgeFlow Build Script
# Automated build for different profiles and platforms

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
VERSION=${VERSION:-0.1.0}
BUILD_DIR="bin"
BINARY_NAME="edgeflow"

# Print banner
print_banner() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           EdgeFlow Build Script v${VERSION}                      ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Print help
print_help() {
    print_banner
    echo -e "${GREEN}Usage:${NC}"
    echo "  ./scripts/build.sh [profile] [platform]"
    echo ""
    echo -e "${GREEN}Profiles:${NC}"
    echo "  minimal  - Pi Zero, BeagleBone (10MB, 50MB RAM)"
    echo "  standard - Pi 3/4 (20MB, 200MB RAM)"
    echo "  full     - Pi 4/5, Jetson (35MB, 400MB RAM)"
    echo "  all      - Build all profiles"
    echo ""
    echo -e "${GREEN}Platforms:${NC}"
    echo "  linux-arm64  - Raspberry Pi 4/5 (64-bit)"
    echo "  linux-arm    - Raspberry Pi 3/Zero (32-bit)"
    echo "  linux-amd64  - Linux x86_64"
    echo "  all          - Build for all platforms"
    echo ""
    echo -e "${GREEN}Examples:${NC}"
    echo "  ./scripts/build.sh minimal linux-arm64"
    echo "  ./scripts/build.sh standard all"
    echo "  ./scripts/build.sh all all"
    echo ""
}

# Build for specific profile and platform
build() {
    local profile=$1
    local platform=$2

    local build_tags=""
    local ldflags=""
    local optimization=""

    # Configure build based on profile
    case $profile in
        minimal)
            build_tags="minimal"
            ldflags="-w -s -X main.Version=${VERSION} -X main.Profile=minimal"
            optimization="-trimpath"
            ;;
        standard)
            build_tags="standard,network,gpio,database"
            ldflags="-w -s -X main.Version=${VERSION} -X main.Profile=standard"
            optimization="-trimpath"
            ;;
        full)
            build_tags="full,network,gpio,database,messaging,ai,industrial,advanced"
            ldflags="-w -X main.Version=${VERSION} -X main.Profile=full"
            optimization=""
            ;;
        *)
            echo -e "${RED}Error: Invalid profile '$profile'${NC}"
            print_help
            exit 1
            ;;
    esac

    # Configure platform
    local goos=""
    local goarch=""
    local goarm=""
    local output_suffix=""

    case $platform in
        linux-arm64)
            goos="linux"
            goarch="arm64"
            output_suffix="linux-arm64"
            ;;
        linux-arm)
            goos="linux"
            goarch="arm"
            goarm="7"
            output_suffix="linux-arm"
            ;;
        linux-amd64)
            goos="linux"
            goarch="amd64"
            output_suffix="linux-amd64"
            ;;
        darwin-arm64)
            goos="darwin"
            goarch="arm64"
            output_suffix="darwin-arm64"
            ;;
        darwin-amd64)
            goos="darwin"
            goarch="amd64"
            output_suffix="darwin-amd64"
            ;;
        *)
            echo -e "${RED}Error: Invalid platform '$platform'${NC}"
            print_help
            exit 1
            ;;
    esac

    # Output file name
    local output_file="${BUILD_DIR}/${BINARY_NAME}-${profile}-${output_suffix}"

    echo -e "${GREEN}Building ${profile} profile for ${platform}...${NC}"
    echo -e "${BLUE}  Tags: ${build_tags}${NC}"
    echo -e "${BLUE}  Output: ${output_file}${NC}"

    # Create build directory
    mkdir -p "${BUILD_DIR}"

    # Build command
    GOOS=${goos} GOARCH=${goarch} ${goarm:+GOARM=${goarm}} \
        go build -tags "${build_tags}" \
        -ldflags "${ldflags}" \
        ${optimization} \
        -o "${output_file}" \
        ./cmd/edgeflow

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Build successful${NC}"
        ls -lh "${output_file}"
        echo ""
    else
        echo -e "${RED}✗ Build failed${NC}"
        exit 1
    fi
}

# Build all profiles for a platform
build_all_profiles() {
    local platform=$1
    echo -e "${BLUE}Building all profiles for ${platform}...${NC}"
    echo ""
    build "minimal" "$platform"
    build "standard" "$platform"
    build "full" "$platform"
}

# Build a profile for all platforms
build_all_platforms() {
    local profile=$1
    echo -e "${BLUE}Building ${profile} profile for all platforms...${NC}"
    echo ""
    build "$profile" "linux-arm64"
    build "$profile" "linux-arm"
    build "$profile" "linux-amd64"
}

# Build everything
build_all() {
    echo -e "${BLUE}Building all profiles for all platforms...${NC}"
    echo ""

    # Linux ARM64 (Pi 4/5)
    build "minimal" "linux-arm64"
    build "standard" "linux-arm64"
    build "full" "linux-arm64"

    # Linux ARM32 (Pi 3/Zero)
    build "minimal" "linux-arm"
    build "standard" "linux-arm"
    build "full" "linux-arm"

    # Linux AMD64 (development)
    build "minimal" "linux-amd64"
    build "standard" "linux-amd64"
    build "full" "linux-amd64"

    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    Build Summary                               ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    ls -lh "${BUILD_DIR}/"
}

# Main script
main() {
    print_banner

    if [ $# -eq 0 ] || [ "$1" == "help" ] || [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
        print_help
        exit 0
    fi

    local profile=$1
    local platform=${2:-linux-arm64}

    if [ "$profile" == "all" ] && [ "$platform" == "all" ]; then
        build_all
    elif [ "$profile" == "all" ]; then
        build_all_profiles "$platform"
    elif [ "$platform" == "all" ]; then
        build_all_platforms "$profile"
    else
        build "$profile" "$platform"
    fi

    echo -e "${GREEN}✓ All builds completed successfully${NC}"
}

main "$@"
