package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.bug.st/serial"
)

func SendToEsp(fanSpeed int) error {
	// should be using serial
	fmt.Println("Sending to ESP32:", fanSpeed)
	return nil
}

func SendCurveToESP32() {
	// read the esp32 ip from config.json
	config, err := LoadConfig[FanCurveConfig]()
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	tmp, err := GetCurrentSystemTemp()
	if err != nil {
		slog.Error("error happened", "err", err)
		return
	}
	speed := CalculateFanSpeed(tmp, config)
	err = SendToEsp(int(speed))
	if err != nil {
		slog.Error("could not send the post req", "err", err)
		return
	}
	slog.Info("Sending data to esp", "temp (C)", tmp, "speed (%)", speed)

	var port serial.Port

	port, err = connectToEsp()
	if err != nil {
		slog.Error("could not connect to esp32, retrying in 5 seconds...", "err", err)
		return
	}
	fmt.Fprintf(port, "%d\n", int(speed))
}

func GetEspPath() (string, error) {
	// loop over all ttyUSB devices and return the first one
	files, err := os.ReadDir("/dev")
	if err != nil {
		slog.Error("error happened", "err", err)
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		containsTTYUSB := strings.Contains(file.Name(), "ttyUSB")
		if !containsTTYUSB {
			continue
		}
		return "/dev/" + file.Name(), nil
	}
	return "", fmt.Errorf("no ttyUSB device found")
}

func connectToEsp() (serial.Port, error) {
	ttyPath, err := GetEspPath()
	if err != nil {
		slog.Error("error happened", "err", err)
		return nil, err
	}
	slog.Info("connected to esp device", "path", ttyPath)

	mode := &serial.Mode{
		BaudRate: 115200, // must match ESP32 Serial.begin()
	}

	port, err := serial.Open(ttyPath, mode)
	if err != nil {
		slog.Error("error happened", "err", err)
		return nil, err
	}
	return port, nil
}
