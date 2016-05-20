package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
)

var shouldExit = false

func main() {
	go listenForStopSignals()

	embd.InitGPIO()
	defer embd.CloseGPIO()
	var led = 4

	embd.SetDirection(led, embd.Out)
	var toggle = true

	for !shouldExit {
		// embd.LEDToggle("GPIO2")

		if toggle {
			embd.DigitalWrite(led, embd.High)
			toggle = false
		} else {
			embd.DigitalWrite(led, embd.Low)
			toggle = true
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func listenForStopSignals() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	<-quit

	shouldExit = true
}
