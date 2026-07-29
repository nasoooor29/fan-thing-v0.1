package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:assets
var assetsFS embed.FS

// enableCORS adds CORS headers to allow frontend access
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func StartWebApp() {
	port := 8080
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

	fmt.Printf("Server starting on http://localhost:%v\n", port)
	fmt.Println("Configuration auto-saves to ./config.json")
	fmt.Println("Curve Points auto-saves to ./curve.json")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func StartTempServer() {
	port := 8081

	http.HandleFunc("GET /api/getFanSpeed", handleGetFanSpeed)
	http.HandleFunc("GET /api/getCurrentTemp", handleGetCurrentTemp)

	fmt.Printf("Server starting on http://localhost:%v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
