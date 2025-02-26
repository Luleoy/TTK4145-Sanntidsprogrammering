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

// getorderincurrentdirection
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

func OrderCompletedatCurrentFloor(floor int, direction Direction, OrderMatrix Orders) [][]int {
	//lage liste over det vi skal fjerne
	var completedOrdersList [][]int //kolonne 1 er floor, kolonne 2 er button
	completedOrdersList.append(floor, elevio.BT_CAB)
	switch direction {
	case elevio.MD_Up:
		completedOrdersList.append(floor, elevio.BT_HallUp)
	case elevio.MD_Down:
		completedOrdersList.append(floor, elevio.BT_HallDown)
	case elevio.MD_Stop:
		completedOrdersList.append(floor, elevio.BT_HallUp)
		completedOrdersList.append(floor, elevio.BT_HallDown)
	}
	return completedOrdersList
}
