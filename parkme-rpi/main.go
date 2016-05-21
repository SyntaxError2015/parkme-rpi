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
	TrigPin  embd.DigitalPin
	EchoPin  embd.DigitalPin
	Code     string
	Occupied bool
}

var sensors = []Sensor{
	Sensor{nil, nil, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
	Sensor{nil, nil, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
	Sensor{nil, nil, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
}

type SensorPins struct {
	TrigPin int
	EchoPin int
}

var sensorPins = []SensorPins{
	SensorPins{4, 17},
	SensorPins{27, 22},
	SensorPins{5, 6},
}

const distLimit = 7

func pulseIn(pin embd.DigitalPin, timeout int) int {
	var myTimeout int

	// for val, _ := pin.Read(); val == embd.Low && myTimeout < timeout; val, _ = pin.Read() {
	// 	myTimeout++
	// 	time.Sleep(1 * time.Microsecond)
	// }

	initTime := time.Now()

	for {
		val, _ := pin.Read()
		if val == embd.High {
			break
		}

		if time.Since(initTime).Nanoseconds() > int64(timeout)*1000 {
			return -1
		}
	}

	// log.Println("low=", time.Since(initTime).Nanoseconds()/1000)

	startTime := time.Now() // Record time when ECHO goes high

	// myTimeout = 0
	// for val, _ := pin.Read(); val == embd.High && myTimeout < timeout; val, _ = pin.Read() {
	// 	myTimeout++
	// 	time.Sleep(1 * time.Microsecond)
	// }

	for {
		val, _ := pin.Read()
		if val == embd.Low {
			break
		}

		if time.Since(initTime).Nanoseconds() > int64(timeout)*1000 {
			return -1
		}
	}

	myTimeout = int(time.Since(startTime).Nanoseconds() / 1000)
	// log.Println("high=", myTimeout)

	return myTimeout
}

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
	// pulseDuration, _ := sensor.EchoPin.TimePulse(embd.High)
	pulseDuration := pulseIn(sensor.EchoPin, 30000)

	// distance := pulseDuration.Nanoseconds() / 58000
	distance := pulseDuration / 58

	// digitalPin.Close()
	time.Sleep(20 * time.Millisecond)

	if distance > 255 || distance < 0 {
		return 255
	}

	return int(distance)
}

func initSensors() {
	for index, sensor := range sensorPins {
		trigPin, _ := embd.NewDigitalPin(sensor.TrigPin)
		echoPin, _ := embd.NewDigitalPin(sensor.EchoPin)

		echoPin.SetDirection(embd.In)
		echoPin.PullDown()

		trigPin.SetDirection(embd.Out)

		sensors[index].TrigPin, sensors[index].EchoPin = trigPin, echoPin
	}

	// trigPin, _ := embd.NewDigitalPin(4)
	// echoPin, err := embd.NewDigitalPin(17)

	// log.Println(err)

	// echoPin.SetDirection(embd.In)
	// echoPin.PullDown()

	// trigPin.SetDirection(embd.Out)

	// sensors[0].TrigPin, sensors[0].EchoPin = trigPin, echoPin
}

// func treatSensor(sensor Sensor) {

// }

func main() {
	go listenForStopSignals()

	defer func() {
		if r := recover(); r != nil {
			log.Println("Prepare to close!", r)
			embd.CloseGPIO()
			log.Println("Error Closed!!!", r)
		}
	}()

	embd.InitGPIO()

	initSensors()

	for !shouldExit {

		for index, sensor := range sensors {
			dist := readSensor(sensor)
			log.Println("d[", index, "]=", dist)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Println("")
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
