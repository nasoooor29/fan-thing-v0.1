package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sort"
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
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
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

// CalculateFanSpeedGradual performs linear interpolation between control points
func CalculateFanSpeedGradual(temperature float64, config *FanCurveConfig) float64 {
	if len(config.Points) == 0 {
		return 0
	}

	// Sort points by temperature
	sortedPoints := make([]FanCurvePoint, len(config.Points))
	copy(sortedPoints, config.Points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].Temperature < sortedPoints[j].Temperature
	})

	// Handle edge cases
	if len(sortedPoints) == 1 {
		return sortedPoints[0].FanSpeed
	}

	if temperature <= sortedPoints[0].Temperature {
		return sortedPoints[0].FanSpeed
	}

	if temperature >= sortedPoints[len(sortedPoints)-1].Temperature {
		return sortedPoints[len(sortedPoints)-1].FanSpeed
	}

	// Find the two points to interpolate between
	for i := 0; i < len(sortedPoints)-1; i++ {
		p1 := sortedPoints[i]
		p2 := sortedPoints[i+1]

		if temperature >= p1.Temperature && temperature <= p2.Temperature {
			// Linear interpolation
			ratio := (temperature - p1.Temperature) / (p2.Temperature - p1.Temperature)
			return p1.FanSpeed + ratio*(p2.FanSpeed-p1.FanSpeed)
		}
	}

	return 0
}

// CalculateFanSpeedHardCut performs step-based interpolation
func CalculateFanSpeedHardCut(temperature float64, config *FanCurveConfig) float64 {
	if len(config.Points) == 0 {
		return 0
	}

	// Sort points by temperature
	sortedPoints := make([]FanCurvePoint, len(config.Points))
	copy(sortedPoints, config.Points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].Temperature < sortedPoints[j].Temperature
	})

	// Find the highest temperature threshold that is <= input temperature
	var lastPoint *FanCurvePoint
	for i := range sortedPoints {
		if sortedPoints[i].Temperature <= temperature {
			lastPoint = &sortedPoints[i]
		} else {
			break
		}
	}

	if lastPoint != nil {
		return lastPoint.FanSpeed
	}

	return 0
}

// CalculateFanSpeed dispatches to the appropriate calculation method
func CalculateFanSpeed(temperature float64, config *FanCurveConfig) float64 {
	if config.InterpolationMode == "hardcut" {
		return CalculateFanSpeedHardCut(temperature, config)
	}
	return CalculateFanSpeedGradual(temperature, config)
}

// GenerateChartData generates the full fan curve data for visualization
func GenerateChartData(config *FanCurveConfig) []CurveDataPoint {
	curveData := make([]CurveDataPoint, 0, 101)
	for temp := 0; temp <= 100; temp++ {
		fanSpeed := CalculateFanSpeed(float64(temp), config)
		curveData = append(curveData, CurveDataPoint{
			X: float64(temp),
			Y: fanSpeed,
		})
	}
	return curveData
}
