#!/bin/bash
# uninstall-service.sh - Remove fan-curve-server systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="fan-curve-server"
INSTALL_DIR="/opt/${SERVICE_NAME}"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo -e "${GREEN}Uninstalling ${SERVICE_NAME} service${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

# Stop the service if running
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo -e "${YELLOW}Stopping service...${NC}"
    systemctl stop "$SERVICE_NAME"
fi

# Disable the service if enabled
if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo -e "${YELLOW}Disabling service...${NC}"
    systemctl disable "$SERVICE_NAME"
fi

# Remove service file
if [ -f "$SERVICE_FILE" ]; then
    echo -e "${YELLOW}Removing service file...${NC}"
    rm -f "$SERVICE_FILE"
fi

# Reload systemd daemon
echo -e "${YELLOW}Reloading systemd daemon...${NC}"
systemctl daemon-reload
systemctl reset-failed 2>/dev/null || true

# Remove installation directory
if [ -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Removing installation directory...${NC}"
    rm -rf "$INSTALL_DIR"
fi

echo -e "${GREEN}Service uninstalled successfully!${NC}"
echo ""
echo "Note: Configuration files (config.json, curve.json) were left in place."
echo "To remove them manually, run:"
echo "  rm -f $INSTALL_DIR/config.json $INSTALL_DIR/curve.json"
