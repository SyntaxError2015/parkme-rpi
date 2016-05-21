package models

import "gopkg.in/mgo.v2/bson"

// Slot reprents a square in a parking lot, where the car is parked
type Slot struct {
	ID         bson.ObjectId `json:"id"`
	Park       Park          `json:"park"`
	Position   Point         `json:"position"`
	IsOccupied bool          `json:"isOccupied"`
}
