package single_elevator

import (
	"TTK4145-Heislab/driver-go/elevio"
)

type Direction int

const (
	Down Direction = iota
	Up
)

func (d Direction) convertMD() elevio.MotorDirection {
	switch d {
	case Down:
		return elevio.MD_Down
	case Up:
		return elevio.MD_Up
	default:
		return elevio.MD_Stop
	}
}

func (d Direction) convertBT() elevio.ButtonType {
	switch d {
	case Down:
		return elevio.BT_HallDown
	case Up:
		return elevio.BT_HallUp
	default:
		return elevio.BT_Cab
	}
}

// invert motordirection to get opposite direction
func (d Direction) invertMD() Direction {
	switch d {
	case Down:
		return Up
	case Up:
		return Down
	default:
		return Down
	}
}
