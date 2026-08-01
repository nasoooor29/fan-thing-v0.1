package web

import (
	"fmt"
	"log/slog"
	"net/http"

	"fan-curve-server/models"
	"fan-curve-server/utils"
)

func handleGetFanSpeed(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	// Placeholder implementation
	w.Header().Set("Content-Type", "text/plain")
	// get current temp
	config, err := utils.LoadConfig[models.FanCurveConfig](models.CONFIG_FILE)
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}
	temp, err := utils.GetCurrentSystemTemp()
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	fanSpeed := utils.CalculateFanSpeed(float64(temp), config)
	fmt.Fprintf(w, "%v", int(fanSpeed))
}

func handleGetCurrentTemp(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	temp, err := utils.GetCurrentSystemTemp()
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	fmt.Fprintf(w, "%v", int(temp))
}
