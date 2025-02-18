package configuration

import (
	"time")

const( 
	NumFloors = 4
	NumElevators = 3
	NumButtons = 3 

	DisconnectTime = 1*time.Second
	DoorOpenDuration = 3*time.Second
	WatchdogTime = 5*time.Second
)