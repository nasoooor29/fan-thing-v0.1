package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:assets
var assetsFS embed.FS

var storage *Storage

// enableCORS adds CORS headers to allow frontend access
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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

	var req FanCurveConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Auto-save configuration
	storage.SaveConfig(req.Points, req.InterpolationMode)

	curveData := GenerateChartData(req.Points, req.InterpolationMode)
	response := GenerateCurveResponse{
		CurveData:     curveData,
		ControlPoints: req.Points,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetConfig returns the saved configuration
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := storage.LoadConfig()
	if err != nil {
		// Return empty/default config if none exists
		config = &FanCurveConfig{
			Points: []FanCurvePoint{
				{Temperature: 30, FanSpeed: 25},
				{Temperature: 60, FanSpeed: 50},
				{Temperature: 80, FanSpeed: 100},
			},
			InterpolationMode: "gradual",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func main() {
	// Initialize storage
	storage = NewStorage()

	// Get the embedded assets filesystem
	assetsSubFS, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		log.Fatal("Failed to load embedded assets:", err)
	}

	// API endpoints (must come before static files)
	http.HandleFunc("/api/generate-curve", handleGenerateCurve)
	http.HandleFunc("/api/config", handleGetConfig)

	// Serve static files from embedded assets
	http.Handle("/", http.FileServer(http.FS(assetsSubFS)))

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	fmt.Println("All assets are embedded in the binary!")
	fmt.Println("Configuration auto-saves to ./config.json")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
