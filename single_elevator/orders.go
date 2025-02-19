package single_elevator

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
)

type Orders [configuration.NumFloors][configuration.NumButtons]bool //creating matrix to take orders. floors*buttons

// checking to see if there are any orders on the direction elevator is moving
func (OrderMatrix Orders) OrderinCurrentDirection(floor int, direction Direction) bool {
	switch direction {
	case Up:
		for f := floor + 1; f < configuration.NumFloors; f++ {
			if OrderMatrix[f][elevio.BT_HallUp] || OrderMatrix[f][elevio.BT_Cab] {
				return true
			}
		}
	case Down:
		for f := 0; f < floor; f++ {
			if OrderMatrix[f][elevio.BT_HallUp] || OrderMatrix[f][elevio.BT_Cab] {
				return true
			}
		}
	}
	return false
}

func OrderCompleted(floor int, direction Direction, OrderMatrix Orders, orderCompletedChannel chan<- elevio.ButtonEvent) bool {
	if OrderMatrix[floor][elevio.BT_Cab] {
		orderCompletedChannel <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}
	if OrderMatrix[floor][direction] {
		orderCompletedChannel <- elevio.ButtonEvent{Floor: floor, Button: direction.convertBT()}
	}
	return false
}
