package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"parkme-rpi/models"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
)

var shouldExit = false

// Sensor struct
type Sensor struct {
	TrigPin  embd.DigitalPin
	EchoPin  embd.DigitalPin
	Code     string
	Occupied bool
}

//SensorPins for sensor pins
type SensorPins struct {
	TrigPin int
	EchoPin int
}

//Sensor specific
var sensors = []Sensor{
	Sensor{nil, nil, "775e77a1-d8be-4fe4-bde7-276f0f3a8f1e", false},
	Sensor{nil, nil, "97e2d498-af63-4146-b3b2-68916ffc30b2", false},
	Sensor{nil, nil, "c835d1d6-1cc0-4a08-ace7-bcaefd4e384d", false},
}

var sensorPins = []SensorPins{
	SensorPins{4, 17},
	SensorPins{27, 22},
	SensorPins{5, 6},
}

//constants
const distLimit = 15
const apiURL = ""

// read the HIGH pulse from a specific pin
func pulseIn(pin embd.DigitalPin, timeout int) int {

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

	for {
		val, _ := pin.Read()
		if val == embd.Low {
			break
		}

		if time.Since(initTime).Nanoseconds() > int64(timeout)*1000 {
			return -1
		}
	}

	myTimeout := int(time.Since(startTime).Nanoseconds() / 1000)
	// log.Println("high=", myTimeout)

	return myTimeout
}

func readSensor(sensor Sensor) int {

	time.Sleep(5 * time.Millisecond)

	//Start Ranging
	sensor.TrigPin.Write(embd.Low)
	time.Sleep(2 * time.Microsecond)

	sensor.TrigPin.Write(embd.High)
	time.Sleep(10 * time.Microsecond)

	sensor.TrigPin.Write(embd.Low)

	pulseDuration := pulseIn(sensor.EchoPin, 30000)

	distance := pulseDuration / 58

	time.Sleep(20 * time.Millisecond)

	if distance > 255 || distance < 0 {
		return 255
	}

	return int(distance)
}

//Send to the server the parking slot state
func sendParkingState(sensor *Sensor) {
	log.Println(sensor.Code, sensor.Occupied)
}

func changeParkingState(sensor *Sensor, occupied bool) {
	if sensor.Occupied != occupied {
		sensor.Occupied = occupied

		sendParkingState(sensor)
	}
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

}

// HTTP part
func test() {

	var park models.Park
	park.ID = bson.NewObjectId()
	park.AppUserID = bson.NewObjectId()
	park.Slots = []models.Slot{
		models.Slot{
			ID: bson.NewObjectId(),
		},
	}

	js, _ := SerializeJSON(park)

	log.Println(string(js))

	// var jsonStr = []byte(`{"id":"park-1"}`)
	resp, err := http.Post("http://vpn.nbi.ninja:1234/api1/parks/Register", "application/json", bytes.NewBuffer(js))
	log.Println(resp, err)
}

//End HTTP

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

	test()

	for !shouldExit {

		for index, sensor := range sensors {
			dist := readSensor(sensor)
			//log.Println("d[", index, "]=", dist)

			changeParkingState(&sensors[index], dist < distLimit)
		}
		time.Sleep(1000 * time.Millisecond)
		//log.Println("")
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
