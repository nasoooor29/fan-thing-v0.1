package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"time"
)

//go:embed all:assets
var assetsFS embed.FS

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
	Save(CONFIG_FILE, &req)

	curveData := GenerateChartData(&req)

	response := map[string]any{
		"curveData":     curveData,
		"controlPoints": req.Points,
	}
	err := Save(CURVE_FILE, response)
	if err != nil {
		slog.Error("could not save the curve file", "err", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	go SendCurveToESP32()
}

// handleGetConfig returns the saved configuration
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	config, err := LoadConfig[FanCurveConfig]()
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
		Save(CONFIG_FILE, config)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func handleGetFanSpeed(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	// Placeholder implementation
	w.Header().Set("Content-Type", "text/plain")
	// get current temp
	config, err := LoadConfig[FanCurveConfig]()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}
	temp, err := GetCurrentSystemTemp()
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	fanSpeed := CalculateFanSpeed(float64(temp), config)
	fmt.Fprintf(w, "%v", int(fanSpeed))
}

func main() {
	go func() {
		for {
			SendCurveToESP32()
			time.Sleep(INTREVAL_MS * time.Millisecond)
		}
	}()
	// Get the embedded assets filesystem
	assetsSubFS, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		log.Fatal("Failed to load embedded assets:", err)
	}

	// API endpoints (must come before static files)
	http.HandleFunc("POST /api/generate-curve", handleGenerateCurve)
	http.HandleFunc("GET /api/config", handleGetConfig)
	http.HandleFunc("GET /api/getFanSpeed", handleGetFanSpeed)

	// Serve static files from embedded assets
	http.Handle("/", http.FileServer(http.FS(assetsSubFS)))

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	fmt.Println("Configuration auto-saves to ./config.json")
	fmt.Println("Curve Points auto-saves to ./curve.json")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
