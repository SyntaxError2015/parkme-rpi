package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
)

var shouldExit = false

// Sensor part

type Sensor struct {
	// TrigPin        int
	// EchoPin        int
	TrigPin  embd.DigitalPin
	EchoPin  embd.DigitalPin
	Code     string
	Occupied bool
}

var sensors = []Sensor{
	Sensor{nil, nil, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
	// Sensor{27, 22, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
	// Sensor{5, 6, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
}

const distLimit = 7

func readSensor(sensor Sensor) int {
	// digitalPin, _ := embd.NewDigitalPin(sensor.EchoPin)
	// digitalPin.SetDirection(embd.In)
	// digitalPin.PullDown()

	time.Sleep(5 * time.Millisecond)

	//Start Ranging
	// embd.DigitalWrite(sensor.TrigPin, embd.Low)
	sensor.TrigPin.Write(embd.Low)
	time.Sleep(2 * time.Microsecond)

	// embd.DigitalWrite(sensor.TrigPin, embd.High)
	sensor.TrigPin.Write(embd.High)
	time.Sleep(10 * time.Microsecond)

	// embd.DigitalWrite(sensor.TrigPin, embd.Low)
	sensor.TrigPin.Write(embd.Low)

	// pulseDuration, _ := digitalPin.TimePulse(embd.High)
	pulseDuration, _ := sensor.EchoPin.TimePulse(embd.High)

	distance := pulseDuration.Nanoseconds() / 58000

	// digitalPin.Close()
	time.Sleep(20 * time.Millisecond)

	if distance > 255 || distance <= 0 {
		return 255
	}

	return int(distance)
}

func initSensors() {
	// for _, sensor := range sensors {
	// 	embd.SetDirection(sensor.TrigPin, embd.Out)
	// 	// embd.SetDirection(sensor.EchoPin, embd.In)
	// }

	trigPin, _ := embd.NewDigitalPin(4)
	echoPin, _ := embd.NewDigitalPin(17)

	echoPin.SetDirection(embd.In)
	echoPin.PullDown()

	trigPin.SetDirection(embd.Out)

	sensors[0].TrigPin, sensors[0].EchoPin = trigPin, echoPin
}

// func treatSensor(sensor Sensor) {

// }

func main() {
	// embd.CloseGPIO()
	// time.Sleep(500 * time.Millisecond)

	go listenForStopSignals()
	defer func() {
		if r := recover(); r != nil {
			log.Println("Prepare to close!")
			embd.CloseGPIO()
			log.Println("Error Closed!!!", r)
		}
	}()

	embd.InitGPIO()

	initSensors()

	// toggle := false

	for !shouldExit {

		// if toggle {
		// 	embd.DigitalWrite(4, embd.High)
		// 	toggle = false
		// } else {
		// 	embd.DigitalWrite(4, embd.Low)
		// 	toggle = true
		// }

		dist := readSensor(sensors[0])
		log.Println("d=", dist)

		time.Sleep(500 * time.Millisecond)

	}

	embd.CloseGPIO()
	log.Println("Closed!!!")
}

func listenForStopSignals() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

	shouldExit = true
}
