package models

import "gopkg.in/mgo.v2/bson"

// Park struct
type Park struct {
	ID        bson.ObjectId `json:"id"`
	AppUserID bson.ObjectId `json:"appUserID"`
	Address   string        `json:"address"`
	Status    int           `json:"status"`
	Position  Point         `json:"position"`
	Slots     []Slot        `json:"slots"`
}
