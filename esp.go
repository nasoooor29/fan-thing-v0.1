package main

import "fmt"

func SendToEsp(fanSpeed int) error {
	// should be using serial
	fmt.Println("Sending to ESP32:", fanSpeed)
	return nil
}
