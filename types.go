package main

import "go.bug.st/serial"

const (
	CONFIG_FILE = "config.json"
	CURVE_FILE  = "curve.json"
	INTREVAL_MS = 1000
	ESP_URL     = "http://smthing.local/fan-speed"
)

var PORT serial.Port

// FanCurvePoint represents a temperature to fan speed mapping
type FanCurvePoint struct {
	Temperature float64 `json:"temperature"`
	FanSpeed    float64 `json:"fanSpeed"`
}

// FanCurveConfig represents a complete fan curve configuration
type FanCurveConfig struct {
	Points            []FanCurvePoint `json:"points"`
	InterpolationMode string          `json:"interpolationMode"` // "gradual" or "hardcut"
}

// CurveDataPoint represents a single point in the generated curve
type CurveDataPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
