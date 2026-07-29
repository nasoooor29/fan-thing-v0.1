package web

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"fan-curve-server/models"
	"fan-curve-server/utils"
)

// handleGenerateCurve generates the full fan curve data for visualization
func handleGenerateCurve(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	var req models.FanCurveConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Auto-save configuration
	utils.Save(models.CONFIG_FILE, &req)

	curveData := utils.GenerateChartData(&req)

	response := map[string]any{
		"curveData":     curveData,
		"controlPoints": req.Points,
	}
	err := utils.Save(models.CURVE_FILE, response)
	if err != nil {
		slog.Error("could not save the curve file", "err", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	go utils.SendCurveToESP32()
}

// handleGetConfig returns the saved configuration
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	config, err := utils.LoadConfig[models.FanCurveConfig](models.CONFIG_FILE)
	if err != nil {
		// Return empty/default config if none exists
		config = &models.FanCurveConfig{
			Points: []models.FanCurvePoint{
				{Temperature: 30, FanSpeed: 25},
				{Temperature: 60, FanSpeed: 50},
				{Temperature: 80, FanSpeed: 100},
			},
			InterpolationMode: "gradual",
		}
		utils.Save(models.CONFIG_FILE, config)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
