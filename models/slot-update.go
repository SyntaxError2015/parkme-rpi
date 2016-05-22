package models

import "gopkg.in/mgo.v2/bson"

type SlotUpdate struct {
	ParkID bson.ObjectId `json:"parkID"`
	Slot   Slot          `json:"slot"`
}
