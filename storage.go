package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

// Storage handles all data persistence and business logic
type Storage struct {
	mu sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{}
}

// SaveConfig saves the fan curve configuration to disk
func (s *Storage) Save(name string, config any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(name, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the configuration from disk
func (s *Storage) LoadConfig() (*FanCurveConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no saved configuration found")
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FanCurveConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
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
