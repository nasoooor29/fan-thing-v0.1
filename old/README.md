# Fan Curve Server

A Go-based fan control server with web interface for managing fan curves based on system temperature.

## Quick Install (Linux)

One-line install command:
```bash
curl -sSL https://raw.githubusercontent.com/nasoooor29/fan-thing-v0.1/main/quick-install.sh | sudo bash
```

Or install specific version:
```bash
curl -sSL https://raw.githubusercontent.com/nasoooor29/fan-thing-v0.1/main/quick-install.sh | sudo bash -s v1.0.0
```

## Manual Installation

### Download Pre-built Binaries

Download the appropriate release for your platform from the [releases page](https://github.com/nasoooor29/fan-thing-v0.1/releases).

Available platforms:
- Linux (amd64, arm64, armv7)
- macOS (amd64/Intel, arm64/Apple Silicon)
- Windows (amd64)

### Linux Service Installation

For Linux systems, you can install as a systemd service:

```bash
sudo ./install-service.sh
```

This will:
- Download the latest release from GitHub
- Install to `/opt/fan-curve-server/`
- Create and enable a systemd service
- Auto-start on boot

To install a specific version:
```bash
sudo ./install-service.sh v1.0.0
```

### Service Management

```bash
# Check status
sudo systemctl status fan-curve-server

# Stop service
sudo systemctl stop fan-curve-server

# Start service
sudo systemctl start fan-curve-server

# Restart service
sudo systemctl restart fan-curve-server

# View logs
sudo journalctl -u fan-curve-server -f
```

### Uninstall Service

```bash
sudo ./uninstall-service.sh
```

## Manual Usage

Simply run the binary:

```bash
# Linux/macOS
./fan-curve-server

# Windows
fan-curve-server.exe
```

Then open http://localhost:8080 in your browser.

## Building from Source

### Prerequisites
- Go 1.25.5 or later

### Build for current platform
```bash
go build -o fan-curve-server .
```

### Build for all platforms
```bash
./build-release.sh v1.0.0
```

This creates release archives for all supported platforms in the `dist/` directory.

## Creating a Release

The build script can automatically create a GitHub release:

```bash
# Build and create release
./build-release.sh v1.0.0

# When prompted, choose 'y' to create GitHub release
```

Requirements for automatic release:
- GitHub CLI (`gh`) installed
- Git repository with remote configured
- Proper GitHub authentication

## Features

- Web-based fan curve configuration
- Multiple interpolation modes
- Automatic configuration saving
- ESP32 integration support
- System temperature monitoring
- RESTful API

## Configuration

Configuration files are stored in the working directory:
- `config.json` - Fan curve configuration
- `curve.json` - Generated curve data

## API Endpoints

- `GET /api/config` - Get current configuration
- `POST /api/generate-curve` - Generate and save fan curve
- `GET /api/getFanSpeed` - Get current fan speed based on temperature

## Development

### Project Structure
```
.
├── main.go           # Main server and HTTP handlers
├── esp.go            # ESP32 communication
├── temp.go           # Temperature monitoring
├── storage.go        # Configuration persistence
├── types.go          # Type definitions
├── assets/           # Web UI assets
└── scripts/          # Build and install scripts
```

## License

See LICENSE file for details.
