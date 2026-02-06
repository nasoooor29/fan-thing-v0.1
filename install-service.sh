#!/bin/bash
# install-service.sh - Install fan-curve-server as a systemd service

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
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="/opt/${SERVICE_NAME}"
BINARY_PATH="${INSTALL_DIR}/${APP_NAME}"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
USER="${SUDO_USER:-$USER}"
VERSION="${1:-latest}"

echo -e "${GREEN}Installing ${SERVICE_NAME} as a systemd service${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
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

OS="linux"
echo -e "${BLUE}Detected platform: ${OS}/${GOARCH}${NC}"

# Create installation directory
echo -e "${YELLOW}Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"

# Download binary from GitHub releases
echo -e "${YELLOW}Downloading ${SERVICE_NAME} ${VERSION} from GitHub...${NC}"

if [ "$VERSION" = "latest" ]; then
    # Get latest release version
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${APP_NAME}-${OS}-${GOARCH}.tar.gz"
    echo -e "${BLUE}Downloading latest release...${NC}"
else
    # Use specific version
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${APP_NAME}-${VERSION}-${OS}-${GOARCH}.tar.gz"
    echo -e "${BLUE}Downloading version ${VERSION}...${NC}"
fi

# Download and extract
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

if ! curl -L -f -o release.tar.gz "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Failed to download from GitHub releases${NC}"
    echo "URL: $DOWNLOAD_URL"
    echo ""
    echo -e "${YELLOW}Falling back to local build...${NC}"
    cd "$CURRENT_DIR"
    
    # Check if go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed and binary download failed${NC}"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    go build -o "${APP_NAME}" .
    if [ ! -f "${APP_NAME}" ]; then
        echo -e "${RED}Error: Build failed${NC}"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    cp "${APP_NAME}" "$BINARY_PATH"
else
    # Extract the downloaded archive
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
    
    echo -e "${YELLOW}Installing binary...${NC}"
    cp "${EXTRACTED_DIR}/${APP_NAME}" "$BINARY_PATH"
    
    # Copy assets if they exist
    if [ -d "${EXTRACTED_DIR}/assets" ]; then
        cp -r "${EXTRACTED_DIR}/assets" "$INSTALL_DIR/"
    fi
fi

# Clean up temp directory
rm -rf "$TEMP_DIR"

# Set executable permissions
chmod +x "$BINARY_PATH"

# Create systemd service file
echo -e "${YELLOW}Creating systemd service file...${NC}"
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

# Set proper permissions
chmod 644 "$SERVICE_FILE"

# Reload systemd daemon
echo -e "${YELLOW}Reloading systemd daemon...${NC}"
systemctl daemon-reload

# Enable the service
echo -e "${YELLOW}Enabling service...${NC}"
systemctl enable "$SERVICE_NAME"

# Start the service
echo -e "${YELLOW}Starting service...${NC}"
systemctl start "$SERVICE_NAME"

# Check service status
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo -e "${GREEN}Service installed and started successfully!${NC}"
    echo ""
    echo "Useful commands:"
    echo "  Status:  sudo systemctl status $SERVICE_NAME"
    echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
    echo "  Start:   sudo systemctl start $SERVICE_NAME"
    echo "  Restart: sudo systemctl restart $SERVICE_NAME"
    echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "Access the web interface at: http://localhost:8080"
else
    echo -e "${RED}Service failed to start. Check logs with:${NC}"
    echo "  sudo journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi
