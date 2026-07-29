package main

import (
	"log"
	"os"

	"fan-curve-server/web"

	"github.com/kardianos/service"
)

var version = "1.0.0"

type Program struct {
	stop chan struct{}
}

func (p *Program) Start(s service.Service) error {
	p.stop = make(chan struct{})

	go func() {
		web.StartTempServer()
	}()

	return nil
}

func (p *Program) Stop(s service.Service) error {
	close(p.stop)

	log.Println("service stopped")

	return nil
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
		log.Fatal(err)
	}
	// Handle service control actions (install, uninstall, start, stop)
	if len(os.Args) > 1 {
		err := service.Control(s, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
