package main

import (
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
)

func main() {
	embd.InitGPIO()
	defer embd.CloseGPIO()
	var led = 4

	embd.SetDirection(led, embd.Out)
	var toggle = true

	for {
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
