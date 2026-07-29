package utils

import (
	"log/slog"

	"fan-curve-server/models"
)

func SendCurveToESP32() {
	// read the esp32 ip from config.json
	config, err := LoadConfig[models.FanCurveConfig](models.CONFIG_FILE)
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	tmp, err := GetCurrentSystemTemp()
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	speed := CalculateFanSpeed(tmp, config)
	slog.Info("whatever", "speed", speed)
}
