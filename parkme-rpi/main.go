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

//Park
var piPark = models.Park{
	ID:        bson.ObjectIdHex("5740f150c6abb10a0bbaee17"),
	AppUserID: bson.ObjectIdHex("5740f150c6abb10a0bbaee18"),
	Slots:     []models.Slot{
	// models.Slot{
	// 	ID: bson.ObjectIdHex("5740f64ae679ffe299e6cf75"),
	// 	ID: bson.ObjectIdHex("5740f64ae679ffe299e6cf76"),
	// 	ID: bson.ObjectIdHex("5740f64ae679ffe299e6cf77"),
	// },
	},
}

// Sensor struct
type Sensor struct {
	TrigPin  embd.DigitalPin
	EchoPin  embd.DigitalPin
	Code     string
	Occupied bool
	X        float64
	Y        float64
}

//SensorPins for sensor pins
type SensorPins struct {
	TrigPin int
	EchoPin int
}

//Sensor specific
var sensors = []Sensor{
	Sensor{nil, nil, "5740f64ae679ffe299e6cf75", false, 45.754286, 21.198960},
	Sensor{nil, nil, "5740f64ae679ffe299e6cf76", false, 45.754299, 21.199141},
	Sensor{nil, nil, "5740f64ae679ffe299e6cf77", false, 45.754307, 21.199309},
}

var sensorPins = []SensorPins{
	SensorPins{4, 17},
	SensorPins{27, 22},
	SensorPins{5, 6},
}

//constants
const distLimit = 15
const apiURL = "http://vpn.nbi.ninja:1234/api1/"

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

func constructPark() {
	for _, sensor := range sensors {
		piPark.Slots = append(piPark.Slots, models.Slot{
			ID:         bson.ObjectIdHex(sensor.Code),
			IsOccupied: sensor.Occupied,
			Position:   models.Point{X: sensor.X, Y: sensor.Y},
		})
	}
}

func registerPark() {
	constructPark()

	serializedPark, _ := SerializeJSON(piPark)

	// log.Println(string(serializedPark))

	resp, err := http.Post(apiURL+"parks/Register", "application/json", bytes.NewBuffer(serializedPark))

	if err != nil {
		log.Println(resp, err)
	}
}

//Send to the server the parking slot state
func sendParkingState(sensor *Sensor) {
	//log.Println(sensor.Code, sensor.Occupied)

	updatedSlot := models.SlotUpdate{
		ParkID: piPark.ID,
		Slot: models.Slot{
			ID:         bson.ObjectIdHex(sensor.Code),
			IsOccupied: sensor.Occupied,
			Park:       models.Park{ID: piPark.ID},
			Position:   models.Point{X: sensor.X, Y: sensor.Y},
		},
	}

	serializedSlot, _ := SerializeJSON(updatedSlot)
	log.Println(string(serializedSlot))

	req, err := http.NewRequest("PUT", apiURL+"slots/UpdateSlot", bytes.NewBuffer(serializedSlot))
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Println(resp, err)
	}
}

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

	resp, err := http.Post(apiURL+"parks/Register", "application/json", bytes.NewBuffer(js))
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

	//Initialize GPIO stuff & sensors
	embd.InitGPIO()
	initSensors()

	//API call to register the park on the server
	registerPark()

	for !shouldExit {

		for index, sensor := range sensors {
			dist := readSensor(sensor)
			//log.Println("d[", index, "]=", dist)

			changeParkingState(&sensors[index], dist < distLimit)
		}
		time.Sleep(500 * time.Millisecond)
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
