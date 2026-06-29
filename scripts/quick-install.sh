#!/bin/bash
# quick-install.sh - Quick install script for fan-curve-server
# Usage: curl -sSL https://raw.githubusercontent.com/nasoooor29/fan-thing-v0.1/main/quick-install.sh | sudo bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="fan-curve-server"
APP_NAME="fan-curve-server"
GITHUB_REPO="nasoooor29/fan-thing-v0.1"
INSTALL_DIR="/opt/${SERVICE_NAME}"
BINARY_PATH="${INSTALL_DIR}/${APP_NAME}"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
USER="${SUDO_USER:-$USER}"
APP_VERSION="${1:-latest}"

echo -e "${GREEN}=== Fan Curve Server Quick Install ===${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

# Detect OS
if [ ! -f /etc/os-release ]; then
    echo -e "${RED}Error: Cannot detect OS. /etc/os-release not found${NC}"
    exit 1
fi

source /etc/os-release
echo -e "${BLUE}Detected OS: ${NAME} ${VERSION_ID}${NC}"

# Check if systemd is available
if ! command -v systemctl &> /dev/null; then
    echo -e "${RED}Error: systemd is required but not found${NC}"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64|arm64)
        GOARCH="arm64"
        ;;
    armv7l)
        GOARCH="armv7"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

OS_TYPE="linux"
echo -e "${BLUE}Platform: ${OS_TYPE}/${GOARCH}${NC}"
echo ""

# Create installation directory
echo -e "${YELLOW}[1/6] Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"

# Download binary from GitHub releases
echo -e "${YELLOW}[2/6] Downloading ${SERVICE_NAME}...${NC}"

if [ "$APP_VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${APP_NAME}-${OS_TYPE}-${GOARCH}.tar.gz"
else
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${APP_VERSION}/${APP_NAME}-${APP_VERSION}-${OS_TYPE}-${GOARCH}.tar.gz"
fi

echo -e "${BLUE}URL: ${DOWNLOAD_URL}${NC}"

TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

if ! curl -L -f -o release.tar.gz "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Failed to download from GitHub releases${NC}"
    echo "Make sure the release exists for your platform."
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Extract the downloaded archive
echo -e "${YELLOW}[3/6] Extracting archive...${NC}"
tar -xzf release.tar.gz

# Find the binary in the extracted directory
EXTRACTED_DIR=$(find . -maxdepth 1 -type d -name "${APP_NAME}-*" | head -n 1)
if [ -z "$EXTRACTED_DIR" ]; then
    echo -e "${RED}Error: Could not find extracted directory${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Copy binary to installation directory
if [ ! -f "${EXTRACTED_DIR}/${APP_NAME}" ]; then
    echo -e "${RED}Error: Binary not found in extracted archive${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

echo -e "${YELLOW}[4/6] Installing binary...${NC}"
cp "${EXTRACTED_DIR}/${APP_NAME}" "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Copy assets if they exist
if [ -d "${EXTRACTED_DIR}/assets" ]; then
    cp -r "${EXTRACTED_DIR}/assets" "$INSTALL_DIR/"
fi

# Clean up
cd /
rm -rf "$TEMP_DIR"

# Create systemd service file
echo -e "${YELLOW}[5/6] Creating systemd service...${NC}"
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Fan Curve Server
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$BINARY_PATH
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=false
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

chmod 644 "$SERVICE_FILE"

# Reload systemd and start service
echo -e "${YELLOW}[6/6] Starting service...${NC}"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

# Wait and check status
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo ""
    echo -e "${GREEN}=== Installation Complete! ===${NC}"
    echo ""
    echo -e "${GREEN}Service installed and running successfully!${NC}"
    echo ""
    echo "Web interface: ${BLUE}http://localhost:8080${NC}"
    echo ""
    echo "Useful commands:"
    echo "  Status:  ${BLUE}sudo systemctl status $SERVICE_NAME${NC}"
    echo "  Stop:    ${BLUE}sudo systemctl stop $SERVICE_NAME${NC}"
    echo "  Start:   ${BLUE}sudo systemctl start $SERVICE_NAME${NC}"
    echo "  Restart: ${BLUE}sudo systemctl restart $SERVICE_NAME${NC}"
    echo "  Logs:    ${BLUE}sudo journalctl -u $SERVICE_NAME -f${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}Error: Service failed to start${NC}"
    echo "Check logs with: ${BLUE}sudo journalctl -u $SERVICE_NAME -n 50${NC}"
    exit 1
fi
