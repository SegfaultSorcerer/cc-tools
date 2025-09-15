#!/bin/bash

# CCTools Build Script
# Compiles for multiple platforms and architectures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build information
APP_NAME="cctools"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(go version | awk '{print $3}')

echo -e "${BLUE}Building $APP_NAME $VERSION${NC}"
echo -e "${BLUE}Go Version: $GO_VERSION${NC}"
echo -e "${BLUE}Build Time: $BUILD_TIME${NC}"
echo ""

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf dist/
mkdir -p dist

# Build flags
LDFLAGS="-s -w -X main.version=$VERSION -X main.buildTime=$BUILD_TIME"

# Build function
build_for_platform() {
    local os=$1
    local arch=$2
    local extension=$3
    local platform_name="${os}_${arch}"

    echo -e "${BLUE}Building for $os/$arch...${NC}"

    # Create platform directory
    mkdir -p "dist/$platform_name"

    # Set environment variables
    export GOOS=$os
    export GOARCH=$arch
    export CGO_ENABLED=0

    # Build binary
    local output_name="$APP_NAME$extension"
    go build -ldflags "$LDFLAGS" -o "dist/$platform_name/$output_name" .

    if [ $? -eq 0 ]; then
        local size=$(du -h "dist/$platform_name/$output_name" | cut -f1)
        echo -e "${GREEN}✓ Built $platform_name ($size)${NC}"
    else
        echo -e "${RED}✗ Failed to build $platform_name${NC}"
        exit 1
    fi

    # Copy documentation
    cp -r docs/ "dist/$platform_name/"
    cp README.md "dist/$platform_name/" 2>/dev/null || echo "# $APP_NAME $VERSION" > "dist/$platform_name/README.md"

    # Create platform-specific archive
    cd dist
    if command -v zip >/dev/null 2>&1; then
        if [ "$os" = "windows" ]; then
            zip -r "${APP_NAME}_${platform_name}.zip" "$platform_name/" >/dev/null
        else
            tar -czf "${APP_NAME}_${platform_name}.tar.gz" "$platform_name/"
        fi
        echo -e "${GREEN}✓ Archived $platform_name${NC}"
    fi
    cd ..
}

echo -e "${YELLOW}Starting cross-compilation...${NC}"
echo ""

# Windows builds
echo -e "${BLUE}=== Windows Builds ===${NC}"
build_for_platform "windows" "386" ".exe"
build_for_platform "windows" "amd64" ".exe"

# Linux builds
echo -e "${BLUE}=== Linux Builds ===${NC}"
build_for_platform "linux" "386" ""
build_for_platform "linux" "amd64" ""
build_for_platform "linux" "arm64" ""

# macOS builds
echo -e "${BLUE}=== macOS Builds ===${NC}"
build_for_platform "darwin" "amd64" ""
build_for_platform "darwin" "arm64" ""

# Additional platforms (optional)
echo -e "${BLUE}=== Additional Platforms ===${NC}"
build_for_platform "freebsd" "amd64" ""

echo ""
echo -e "${GREEN}✓ Build completed successfully!${NC}"
echo ""

# Show build summary
echo -e "${YELLOW}Build Summary:${NC}"
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│ Platform     │ Architecture │ Binary Size │ Archive         │"
echo "├─────────────────────────────────────────────────────────────┤"

for dir in dist/*/; do
    if [ -d "$dir" ]; then
        platform=$(basename "$dir")
        binary_path="$dir$APP_NAME"
        [ "$platform" = "windows_"* ] && binary_path="${binary_path}.exe"

        if [ -f "$binary_path" ]; then
            size=$(du -h "$binary_path" | cut -f1)
            os_arch=$(echo "$platform" | sed 's/_/ \/ /')

            # Find archive
            archive=""
            if [ -f "dist/${APP_NAME}_${platform}.zip" ]; then
                archive="${APP_NAME}_${platform}.zip"
            elif [ -f "dist/${APP_NAME}_${platform}.tar.gz" ]; then
                archive="${APP_NAME}_${platform}.tar.gz"
            fi

            printf "│ %-12s │ %-12s │ %-11s │ %-15s │\n" \
                "$(echo $os_arch | cut -d'/' -f1)" \
                "$(echo $os_arch | cut -d'/' -f3)" \
                "$size" \
                "$archive"
        fi
    fi
done

echo "└─────────────────────────────────────────────────────────────┘"
echo ""

# Show total size
total_size=$(du -sh dist/ | cut -f1)
echo -e "${BLUE}Total build size: $total_size${NC}"

# List all files
echo ""
echo -e "${YELLOW}Generated files:${NC}"
find dist/ -type f -name "$APP_NAME*" -o -name "*.zip" -o -name "*.tar.gz" | sort

echo ""
echo -e "${GREEN}Build script completed! 🚀${NC}"
echo -e "${BLUE}Files are available in the 'dist/' directory${NC}"