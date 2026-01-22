package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
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
