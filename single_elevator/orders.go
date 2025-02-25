package single_elevator

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
)

func SetLights(orderMatrix Orders) { //skru av og på lys
	for f := 0; f < configuration.NumFloors; f++ {
		for b := 0; b < configuration.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, orderMatrix[f][b])
		}
	}
}

type Orders [configuration.NumFloors][configuration.NumButtons]bool //creating matrix to take orders. floors*buttons

/*
func (OrderMatrix Orders) OrderinCurrentDirection(floor int, direction Direction) bool {
	switch direction {
	case Up:
		for f := floor + 1; f < configuration.NumFloors; f++ {
			for b := 0; b < configuration.NumButtons; b++ {
				if OrderMatrix[f][b] {
					return true
				}
			}
		}
		return false
	case Down:
		for f := 0; f < floor; f++ {
			for b := 0; b < configuration.NumButtons; b++ {
				if OrderMatrix[f][b] {
					return true
				}
			}
		}
		return false
	default:
		panic("Invalid direction")
	}

}
*/

func OrderCompleted(floor int, direction Direction, OrderMatrix Orders, orderCompletedChannel chan<- elevio.ButtonEvent) {
	if OrderMatrix[floor][elevio.BT_Cab] {
		orderCompletedChannel <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}
	if OrderMatrix[floor][direction] {
		orderCompletedChannel <- elevio.ButtonEvent{Floor: floor, Button: direction.convertBT()}
	}
	SetLights(OrderMatrix)
}

func orderHere(orders Orders, floor int) bool {
	for b := 0; b < configuration.NumButtons; b++ {
		if orders[floor][b] == true { // Hvis det finnes en aktiv forespørsel
			return true
		}
	}
	return false
}

func ordersAbove(orders Orders, floor int) bool {
	for f := floor + 1; f < configuration.NumFloors; f++ {
		if orderHere(orders, f) {
			return true
		}
	}
	return false
}

func ordersBelow(orders Orders, floor int) bool {
	for f := floor - 1; f >= 0; f-- {
		if orderHere(orders, f) {
			return true
		}
	}
	return false
}
