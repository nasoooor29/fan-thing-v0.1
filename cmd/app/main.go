package main

import (
	"log"
	"os"
	"time"

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

	log.Println("service stopped")

	return nil
}

func (p *Program) run() {
	go web.StartWebApp()

	log.Println("service started:", version)

	workTicker := time.NewTicker(5 * time.Second)

	defer workTicker.Stop()

	for {
		select {

		case <-workTicker.C:
			// Your actual work
			log.Printf("doing work with version: %v\n", version)

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
		log.Fatal(err)
	}
	// Handle service control actions (install, uninstall, start, stop)
	if len(os.Args) > 0 {
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
