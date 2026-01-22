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

	var req FanCurveConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Auto-save configuration
	storage.Save(CONFIG_FILE, &req)

	curveData := GenerateChartData(&req)

	response := map[string]any{
		"curveData":     curveData,
		"controlPoints": req.Points,
	}
	storage.Save(CURVE_FILE, response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetConfig returns the saved configuration
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
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
		storage.Save(CONFIG_FILE, config)
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
	http.HandleFunc("POST /api/generate-curve", handleGenerateCurve)
	http.HandleFunc("GET /api/config", handleGetConfig)

	// Serve static files from embedded assets
	http.Handle("/", http.FileServer(http.FS(assetsSubFS)))

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	fmt.Println("All assets are embedded in the binary!")
	fmt.Println("Configuration auto-saves to ./config.json")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
