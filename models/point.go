package models

// Point contains the left-upper and bottom-lower points of a rectangle in which an entire zone is located
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
