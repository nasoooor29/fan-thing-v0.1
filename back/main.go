package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
)

// FanCurvePoint represents a temperature to fan speed mapping
type FanCurvePoint struct {
	Temperature float64 `json:"temperature"`
	FanSpeed    float64 `json:"fanSpeed"`
}

// GenerateCurveRequest is the request body for generating curve data
type GenerateCurveRequest struct {
	Points            []FanCurvePoint `json:"points"`
	InterpolationMode string          `json:"interpolationMode"` // "gradual" or "hardcut"
}

// CurveDataPoint represents a single point in the generated curve
type CurveDataPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// GenerateCurveResponse is the response containing the full curve data
type GenerateCurveResponse struct {
	CurveData     []CurveDataPoint `json:"curveData"`
	ControlPoints []FanCurvePoint  `json:"controlPoints"`
}

// enableCORS adds CORS headers to allow frontend access
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// calculateFanSpeedGradual performs linear interpolation
func calculateFanSpeedGradual(temperature float64, points []FanCurvePoint) float64 {
	if len(points) == 0 {
		return 0
	}

	// Sort points by temperature
	sortedPoints := make([]FanCurvePoint, len(points))
	copy(sortedPoints, points)
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

// calculateFanSpeedHardCut performs step-based interpolation
func calculateFanSpeedHardCut(temperature float64, points []FanCurvePoint) float64 {
	if len(points) == 0 {
		return 0
	}

	// Sort points by temperature
	sortedPoints := make([]FanCurvePoint, len(points))
	copy(sortedPoints, points)
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

// calculateFanSpeed dispatches to the appropriate calculation method
func calculateFanSpeed(temperature float64, points []FanCurvePoint, mode string) float64 {
	if mode == "hardcut" {
		return calculateFanSpeedHardCut(temperature, points)
	}
	return calculateFanSpeedGradual(temperature, points)
}

// handleGenerateCurve generates the full fan curve data for visualization
func handleGenerateCurve(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateCurveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate curve data for temperatures 0-100
	curveData := make([]CurveDataPoint, 0, 101)
	for temp := 0; temp <= 100; temp++ {
		fanSpeed := calculateFanSpeed(float64(temp), req.Points, req.InterpolationMode)
		curveData = append(curveData, CurveDataPoint{
			X: float64(temp),
			Y: fanSpeed,
		})
	}

	response := GenerateCurveResponse{
		CurveData:     curveData,
		ControlPoints: req.Points,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Serve static files from front-simple directory
	fs := http.FileServer(http.Dir("../front-simple"))
	http.Handle("/", fs)

	// API endpoint for generating curve data
	http.HandleFunc("/api/generate-curve", handleGenerateCurve)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
