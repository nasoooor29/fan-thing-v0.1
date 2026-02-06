#!/bin/bash
# build-release.sh - Build cross-platform binaries and create GitHub release

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="fan-curve-server"
DIST_DIR="dist"

# Parse version argument
VERSION="${1}"
if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Version number required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Validate version format (should start with 'v')
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${YELLOW}Warning: Version should follow format 'vX.Y.Z' (e.g., v1.0.0)${NC}"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${GREEN}Building ${APP_NAME} ${VERSION} for multiple platforms${NC}"

# Clean and create dist directory
echo -e "${YELLOW}Cleaning dist directory...${NC}"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build information
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm/7"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo -e "${BLUE}Building for platforms:${NC}"
for platform in "${PLATFORMS[@]}"; do
    echo "  - $platform"
done
echo ""

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    GOARM="${platform_split[2]:-}"
    
    # Set output name
    OUTPUT_NAME="${APP_NAME}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    # Set archive name
    ARCHIVE_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    if [ -n "$GOARM" ]; then
        ARCHIVE_NAME="${ARCHIVE_NAME}v${GOARM}"
    fi
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}${GOARM:+v$GOARM}...${NC}"
    
    # Create platform directory
    BUILD_DIR="${DIST_DIR}/${ARCHIVE_NAME}"
    mkdir -p "$BUILD_DIR"
    
    # Build
    if [ -n "$GOARM" ]; then
        env GOOS="$GOOS" GOARCH="$GOARCH" GOARM="$GOARM" \
            go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
            -o "${BUILD_DIR}/${OUTPUT_NAME}" .
    else
        env GOOS="$GOOS" GOARCH="$GOARCH" \
            go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
            -o "${BUILD_DIR}/${OUTPUT_NAME}" .
    fi
    
    # Copy assets and documentation
    cp -r assets "${BUILD_DIR}/"
    
    # Create README for the release
    cat > "${BUILD_DIR}/README.txt" <<EOF
${APP_NAME} ${VERSION}

Platform: ${GOOS}/${GOARCH}${GOARM:+v$GOARM}
Built: ${BUILD_TIME}
Commit: ${GIT_COMMIT}

Quick Start:
1. Run the ${OUTPUT_NAME} binary
2. Open http://localhost:8080 in your browser
3. Configuration saves to config.json

For Linux/macOS service installation, use the install-service.sh script.

EOF
    
    # Copy install scripts for Linux
    if [ "$GOOS" = "linux" ]; then
        cp install-service.sh uninstall-service.sh "${BUILD_DIR}/"
    fi
    
    # Create archive
    cd "$DIST_DIR"
    if [ "$GOOS" = "windows" ]; then
        # Use zip for Windows
        zip -r -q "${ARCHIVE_NAME}.zip" "${ARCHIVE_NAME}"
        echo -e "${GREEN}Created: ${ARCHIVE_NAME}.zip${NC}"
    else
        # Use tar.gz for Unix-like systems
        tar -czf "${ARCHIVE_NAME}.tar.gz" "${ARCHIVE_NAME}"
        echo -e "${GREEN}Created: ${ARCHIVE_NAME}.tar.gz${NC}"
    fi
    cd ..
    
    # Remove build directory (keep only archives)
    rm -rf "${BUILD_DIR}"
done

echo ""
echo -e "${GREEN}Build complete! Archives created in ${DIST_DIR}/${NC}"
echo ""

# Generate checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd "$DIST_DIR"
sha256sum *.{tar.gz,zip} 2>/dev/null > SHA256SUMS || shasum -a 256 *.{tar.gz,zip} > SHA256SUMS
cd ..
echo -e "${GREEN}Checksums saved to ${DIST_DIR}/SHA256SUMS${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI (gh) not found.${NC}"
    echo "To create a release, install gh: https://cli.github.com/"
    echo ""
    echo "Then run:"
    echo "  gh release create ${VERSION} ${DIST_DIR}/* --title \"Release ${VERSION}\" --generate-notes"
    exit 0
fi

# Check if git remote exists
if ! git remote get-url origin &> /dev/null; then
    echo -e "${YELLOW}No git remote found. Skipping release creation.${NC}"
    echo "To create a release manually:"
    echo "  gh release create ${VERSION} ${DIST_DIR}/* --title \"Release ${VERSION}\" --generate-notes"
    exit 0
fi

# Ask if user wants to create a GitHub release
echo -e "${BLUE}Do you want to create a GitHub release now?${NC}"
read -p "This will push the tag and upload all archives (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Creating GitHub release...${NC}"
    
    # Create and push tag if it doesn't exist
    if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
        echo -e "${YELLOW}Creating git tag ${VERSION}...${NC}"
        git tag -a "$VERSION" -m "Release ${VERSION}"
        git push origin "$VERSION"
    fi
    
    # Create release with all archives
    gh release create "$VERSION" ${DIST_DIR}/* \
        --title "Release ${VERSION}" \
        --generate-notes
    
    echo -e "${GREEN}GitHub release created successfully!${NC}"
else
    echo -e "${YELLOW}Skipping GitHub release.${NC}"
    echo "To create a release later, run:"
    echo "  git tag -a ${VERSION} -m 'Release ${VERSION}'"
    echo "  git push origin ${VERSION}"
    echo "  gh release create ${VERSION} ${DIST_DIR}/* --title 'Release ${VERSION}' --generate-notes"
fi

echo ""
echo -e "${GREEN}All done!${NC}"
