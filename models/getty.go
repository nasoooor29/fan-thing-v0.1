package models

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	mqtt "github.com/mochi-mqtt/server/v2"
)

// type Getty interface {
// 	GetTemp(endpoint string) float64
// 	SendTemp(endpoint string) float64
// }

var MQTT *mqtt.Server

type Getty struct {
	DeviceAdrr string
	index      int
}

func (g Getty) GetTemp() (float64, error) {
	endpoint := fmt.Sprintf("http://%v/api/getCurrentTemp", g.DeviceAdrr)
	resp, err := http.Get(endpoint)
	if err != nil {
		slog.Error("error happened", "err", err)
		return 100, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error happened", "err", err)
		return 100, err
	}
	// convert bytes into float the response body is just a number
	temp, err := strconv.ParseFloat(string(body), 64)
	if err != nil {
		slog.Error("error happened", "err", err)
		return 100, err
	}
	return temp, nil
}

func (g Getty) SendSpeed(speed float64) error {
	// MQTT topic names are exact: this intentionally has no leading slash.

	correctSpeed := speed * 10 // cuz fanzy want from 1 to 1000
	mqttTopic := fmt.Sprintf("/fanctl/control/fan/%d/PWM", g.index)
	slog.Info("sending speed", "speed", correctSpeed, "topic", mqttTopic)
	// Retain the latest command so a subscriber that reconnects receives it.
	return MQTT.Publish(mqttTopic, []byte(fmt.Sprintf("%f", correctSpeed)), true, 0)
}

func GenerateGettys(deviceAdrrs []string) []Getty {
	gettys := make([]Getty, len(deviceAdrrs))
	for i, addr := range deviceAdrrs {
		gettys[i] = Getty{
			DeviceAdrr: addr,
			index:      i + 1,
		}
	}
	return gettys
}
