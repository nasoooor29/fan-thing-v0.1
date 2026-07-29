package models

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

// type Getty interface {
// 	GetTemp(endpoint string) float64
// 	SendTemp(endpoint string) float64
// }

type Getty struct {
	brokerIp   string
	deviceAdrr string
}

func (g Getty) GetTemp() (float64, error) {
	endpoint := fmt.Sprintf("http://%v/api/getCurrentTemp", g.deviceAdrr)
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
	return nil
}

func GenerateGettys(brokerIp string, deviceAdrrs []string) []Getty {
	gettys := make([]Getty, len(deviceAdrrs))
	for i, addr := range deviceAdrrs {
		gettys[i] = Getty{
			brokerIp:   brokerIp,
			deviceAdrr: addr,
		}
	}
	return gettys
}
