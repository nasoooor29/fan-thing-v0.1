package web

import (
	"log/slog"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func CreateMqtt() (*mqtt.Server, error) {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	// Development only: allow every client and topic.
	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		slog.Error("failed to add auth hook", "err", err)
		return nil, err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: ":1883",
	})

	if err := server.AddListener(tcp); err != nil {
		slog.Error("failed to add TCP listener", "err", err)
		return nil, err
	}
	return server, nil
}
