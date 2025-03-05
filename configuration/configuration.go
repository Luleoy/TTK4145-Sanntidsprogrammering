package configuration

import (
	"TTK4145-Heislab/worldview"
	"time"
)

const (
	NumFloors    = 4
	NumElevators = 3
	NumButtons   = 3
	Buffer       = 1024

	DisconnectTime   = 1 * time.Second
	DoorOpenDuration = 3 * time.Second
	WatchdogTime     = 5 * time.Second
	SendWVTimer      = 20 * time.Second
)

type RequestState int

const (
	None RequestState = iota
	Order
	Confirmed
	Complete
)
/*
// legge typen i configuration. Lage kanalene de skal sendes på i main.g. structuren på hva som blir sendt på kanalen
type ElvatorSystem struct {
	ElevtorID        string
	localWorldView   *worldview.WorldView
	SendMessageTimer *time.Timer
	numPeers         int
}
*/