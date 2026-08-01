package utils

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GetCurrentSystemTemp reads the current CPU temperature from the system
// Returns temperature in Celsius
func GetCurrentSystemTemp() (float64, error) {
	// Check if ipmitool is available
	tmp, err := getServerTemps()
	if err == nil {
		return tmp, nil
	}
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

func getServerTemps() (float64, error) {
	cpu1, err := getSensorByNameExec("CPU1 Temp")
	if err != nil {
		slog.Error("error happened", "err", err)
		return 0, err
	}
	cpu2, err := getSensorByNameExec("CPU2 Temp")
	if err != nil {
		slog.Error("error happened", "err", err)
		return 0, err
	}
	// average the two temps
	avgTemp := (cpu1 + cpu2) / 2.0
	return avgTemp, nil
}

func getSensorByNameExec(name string) (float64, error) {
	execCmd := exec.Command("ipmitool", "sensor", "get", name)
	output, err := execCmd.Output()
	if err != nil {
		slog.Error("error happened", "err", err)
		return 0, err
	}
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		if strings.Contains(line, "Sensor Reading") {
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}
			readingPart := strings.TrimSpace(parts[1])
			readingParts := strings.Split(readingPart, " ")
			if len(readingParts) < 1 {
				continue
			}
			var temp float64
			_, err := fmt.Sscanf(readingParts[0], "%f", &temp)
			if err != nil {
				slog.Error("error happened", "err", err)
				return 0, err
			}
			return temp, nil
		}
	}
	return 0, fmt.Errorf("sensor not found")
}
