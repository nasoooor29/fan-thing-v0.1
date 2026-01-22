package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// GetCurrentSystemTemp reads the current CPU temperature from the system
// Returns temperature in Celsius
func GetCurrentSystemTemp() (float64, error) {
	// Read from thermal zone 0 (typically the CPU temperature)
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		slog.Error("error happened", "err", err)
		return 100, fmt.Errorf("failed to read temperature file: %w", err)
	}

	// The temperature is in millidegrees Celsius
	tempStr := strings.TrimSpace(string(data))
	tempMilliC, err := strconv.ParseInt(tempStr, 10, 64)
	if err != nil {
		slog.Error("error happened", "err", err)
		return 100, fmt.Errorf("failed to parse temperature: %w", err)
	}

	// Convert from millidegrees to degrees Celsius
	tempC := float64(tempMilliC) / 1000.0
	return tempC, nil
}
