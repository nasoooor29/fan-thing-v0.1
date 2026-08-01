# Fan Thingy

Fan Thingy controls server fan speeds from a temperature-to-speed curve. It consists of a central controller and a lightweight temperature agent installed on every monitored server.

The controller polls each agent for its current temperature, calculates a fan speed from the configured curve, and publishes that speed to [Super Fanzy](https://github.com/qulxizer/superfanzy) over MQTT.

## Architecture

| Component | Runs on | Responsibility | Port |
| --- | --- | --- | --- |
| Main controller | One host | Serves the web UI, runs the MQTT broker, polls agents, and publishes fan commands | HTTP `8080`, MQTT `1883` |
| Temperature agent | Each monitored server | Reads the current temperature using `ipmitool` and exposes it over HTTP | HTTP `8081` |

The controller maps the order of device addresses in the configuration to Fanzy fan topics:

| Device position | MQTT topic |
| --- | --- |
| First address | `/fanctl/control/fan/1/PWM` |
| Second address | `/fanctl/control/fan/2/PWM` |
| Nth address | `/fanctl/control/fan/N/PWM` |

Fan speeds are configured as percentages (`0` to `100`) and sent to Fanzy on its `0` to `1000` scale.

## Requirements

- Go `1.25.12` or newer for local builds
- Linux hosts with `systemd`-compatible service support
- `ipmitool` on every host running the temperature agent
- SSH access as `root` to deployment targets
- Super Fanzy configured to connect to the controller's MQTT broker

The bundled MQTT broker currently permits unauthenticated connections. Keep port `1883` on a trusted network.

## Quick Start

1. Configure Super Fanzy on the fan-controller device and point it at the controller host on MQTT port `1883`.
2. Update the `IPS` arrays in the scripts for your environment.
   - `scripts/authorize-devices.sh`: every target host.
   - `scripts/deploy-temp-app.sh`: every server that will report a temperature.
   - `scripts/deploy-main-app.sh`: the one host that will run the controller.
3. Install your SSH key on the target hosts:

   ```bash
   bash ./scripts/authorize-devices.sh
   ```

4. Build and deploy the temperature agent to every monitored server:

   ```bash
   bash ./scripts/deploy-temp-app.sh
   ```

5. Build and deploy the main controller:

   ```bash
   bash ./scripts/deploy-main-app.sh
   ```

6. Open `http://<controller-host>:8080`, add the temperature-agent addresses in fan order, and draw or edit the fan curve. Changes are saved automatically to `config.json`.

## Configuration

The controller stores its settings in `config.json` in its working directory. The web UI is the recommended way to manage it. A configuration looks like this:

```json
{
  "points": [
    { "temperature": 30, "fanSpeed": 25 },
    { "temperature": 60, "fanSpeed": 50 },
    { "temperature": 80, "fanSpeed": 100 }
  ],
  "interpolationMode": "gradual",
  "ips": [
    "192.168.100.101:8081",
    "192.168.100.102:8081"
  ]
}
```

- `points`: temperature in degrees Celsius and fan speed as a percentage.
- `interpolationMode`: `gradual` linearly interpolates between points; `hardcut` uses the speed of the highest threshold at or below the current temperature.
- `ips`: temperature-agent addresses, including port `8081`, ordered to match Super Fanzy fan numbers.

The controller polls all configured agents every five seconds. It also writes `curve.json`, which contains the generated chart data.

## Temperature Agent API

After deployment, verify an agent directly:

```bash
curl http://192.168.100.101:8081/api/getCurrentTemp
```

The endpoint returns a plain-text Celsius value, for example `49`.

The agent first reads the `CPU1 Temp` and `CPU2 Temp` IPMI sensors and returns their average. If IPMI is unavailable, it falls back to `/sys/class/thermal/thermal_zone0/temp`.

## Local Development

Run the controller:

```bash
go run ./cmd/app
```

Run a temperature agent in a separate terminal:

```bash
go run ./cmd/temp
```

Both programs use service-management arguments such as `install`, `start`, `stop`, and `uninstall` when deployed. The deployment scripts build the binaries and register them as services on the target hosts.

## Project Layout

```text
cmd/app/       Main controller entry point
cmd/temp/      Temperature-agent entry point
web/           HTTP handlers, embedded UI, and MQTT setup
utils/         Temperature reading and curve calculation
scripts/       SSH authorization and deployment helpers
```
