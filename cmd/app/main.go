package main

import (
	"log/slog"
	"os"
	"time"

	"fan-curve-server/models"
	"fan-curve-server/utils"
	"fan-curve-server/web"

	"github.com/kardianos/service"
)

var version = "1.0.0"

type Program struct {
	stop chan struct{}
}

func (p *Program) Start(s service.Service) error {
	p.stop = make(chan struct{})

	go p.run()

	return nil
}

func (p *Program) Stop(s service.Service) error {
	close(p.stop)

	slog.Info("service stopped")

	return nil
}

func (p *Program) run() {
	go web.StartWebApp()

	slog.Info("service started")

	workTicker := time.NewTicker(1 * time.Second)

	defer workTicker.Stop()

	for {
		select {

		case <-workTicker.C:
			// Your actual work

			config, err := utils.LoadConfig[models.FanCurveConfig](models.CONFIG_FILE)
			if err != nil {
				slog.Error("error happened", "err", err)
				continue
			}
			utils.GetCalcSend(config)
		case <-p.stop:
			return
		}
	}
}

func main() {
	config := &service.Config{
		Name:        "FanThingService",
		DisplayName: "FanThing Service",
		Description: "A service to control server fans based on ipmi temp using ipmitool and mqtt.",
	}

	prg := &Program{}

	s, err := service.New(prg, config)
	if err != nil {
		slog.Error("an error occured", "err", err)
		os.Exit(1)
	}
	// Handle service control actions (install, uninstall, start, stop)
	if len(os.Args) > 1 {
		err := service.Control(s, os.Args[1])
		if err != nil {
			slog.Error("an error occured", "err", err)
			os.Exit(1)
		}
		return
	}

	err = s.Run()
	if err != nil {
		slog.Error("an error occured", "err", err)
		os.Exit(1)
	}
}
