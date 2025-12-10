#!/bin/bash

# Production Build Script for Photo Backup Server
# Builds optimized binaries for server and CLI tools

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build configuration
BUILD_DIR="dist"
VERSION=${VERSION:-"1.0.0"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}"

# Supported platforms
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Photo Backup Server production binaries

OPTIONS:
    -v, --version VERSION     Set version (default: $VERSION)
    -p, --platform PLATFORM   Build for specific platform (e.g., darwin/amd64)
    -a, --all                 Build for all supported platforms
    -c, --clean               Clean build directory before building
    -h, --help                Show this help message

EXAMPLES:
    $0                        Build for current platform
    $0 -a                     Build for all platforms
    $0 -p linux/amd64         Build for specific platform
    $0 -c -a                  Clean and build for all platforms

EOF
}

# Parse command line arguments
CLEAN=false
BUILD_ALL=false
TARGET_PLATFORM=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}"
            shift 2
            ;;
        -p|--platform)
            TARGET_PLATFORM="$2"
            shift 2
            ;;
        -a|--all)
            BUILD_ALL=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Clean build directory
clean() {
    log_step "Cleaning build directory..."
    rm -rf "${BUILD_DIR}"
    mkdir -p "${BUILD_DIR}"
}

# Build binary for a specific platform
build_binary() {
    local platform=$1
    local os=$(echo $platform | cut -d'/' -f1)
    local arch=$(echo $platform | cut -d'/' -f2)

    log_info "Building for ${os}/${arch}..."

    # Set environment variables for cross-compilation
    export GOOS=$os
    export GOARCH=$arch
    export CGO_ENABLED=0

    # Create output directory for this platform
    local platform_dir="${BUILD_DIR}/${os}-${arch}"
    mkdir -p "${platform_dir}"

    # Build server
    log_info "  Building server..."
    local server_binary="photo-backup-server"
    if [ "$os" = "windows" ]; then
        server_binary="photo-backup-server.exe"
    fi

    go build -ldflags "${LDFLAGS}" \
        -o "${platform_dir}/${server_binary}" \
        -a -installsuffix cgo \
        ./cmd/server

    # Build CLI
    log_info "  Building CLI..."
    local cli_binary="photo-backup-cli"
    if [ "$os" = "windows" ]; then
        cli_binary="photo-backup-cli.exe"
    fi

    go build -ldflags "${LDFLAGS}" \
        -o "${platform_dir}/${cli_binary}" \
        -a -installsuffix cgo \
        ./cmd/cli

    # Strip binary if not Windows
    if [ "$os" != "windows" ]; then
        strip "${platform_dir}/${server_binary}" 2>/dev/null || true
        strip "${platform_dir}/${cli_binary}" 2>/dev/null || true
    fi

    # Calculate checksums
    log_info "  Generating checksums..."
    cd "${platform_dir}"
    if command -v shasum &> /dev/null; then
        shasum -a 256 * > "checksums.txt"
    elif command -v sha256sum &> /dev/null; then
        sha256sum * > "checksums.txt"
    fi
    cd - > /dev/null

    log_info "  ✓ Built successfully for ${os}/${arch}"
}

# Create release archive
create_archive() {
    local platform=$1
    local os=$(echo $platform | cut -d'/' -f1)
    local arch=$(echo $platform | cut -d'/' -f2)

    log_step "Creating release archive for ${os}/${arch}..."

    local platform_dir="${BUILD_DIR}/${os}-${arch}"
    local archive_name="photo-backup-server-${VERSION}-${os}-${arch}"

    cd "${platform_dir}"

    # Create archive
    if [ "$os" = "windows" ]; then
        zip -r "../${archive_name}.zip" . > /dev/null
        log_info "  ✓ Created ${archive_name}.zip"
    else
        tar -czf "../${archive_name}.tar.gz" .
        log_info "  ✓ Created ${archive_name}.tar.gz"
    fi

    cd - > /dev/null
}

# Main build function
main() {
    echo "=========================================="
    echo "  Photo Backup Server - Production Build"
    echo "=========================================="
    echo ""
    log_info "Version: ${VERSION}"
    log_info "Commit: ${COMMIT}"
    log_info "Build Time: ${BUILD_TIME}"
    echo ""

    # Clean if requested
    if [ "$CLEAN" = true ]; then
        clean
    else
        # Create build directory if it doesn't exist
        mkdir -p "${BUILD_DIR}"
    fi

    # Determine what to build
    if [ "$BUILD_ALL" = true ]; then
        log_step "Building for all platforms..."
        for platform in "${PLATFORMS[@]}"; do
            build_binary "$platform"
            create_archive "$platform"
            echo ""
        done
    elif [ -n "$TARGET_PLATFORM" ]; then
        log_step "Building for ${TARGET_PLATFORM}..."
        build_binary "$TARGET_PLATFORM"
        create_archive "$TARGET_PLATFORM"
    else
        # Build for current platform
        log_step "Building for current platform..."
        # Get current platform
        local current_platform=$(go env GOOS)/$(go env GOARCH)
        build_binary "$current_platform"
        create_archive "$current_platform"
    fi

    echo ""
    echo "=========================================="
    echo "  Build Complete!"
    echo "=========================================="
    echo ""
    log_info "Output directory: ${BUILD_DIR}/"
    echo ""

    # List all artifacts
    log_info "Artifacts:"
    find "${BUILD_DIR}" -type f -name "*.tar.gz" -o -name "*.zip" -o -name "checksums.txt" | while read file; do
        local size=$(du -h "$file" | cut -f1)
        echo "  - $(basename $file) ($size)"
    done

    echo ""
    log_info "Build artifacts are ready in ${BUILD_DIR}/"
}

# Run main
main
