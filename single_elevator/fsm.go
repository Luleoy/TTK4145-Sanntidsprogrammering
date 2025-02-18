package single_elevator

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"time"
	//"fmt"
)

type State struct {
	Floor      int
	Direction  Direction
	Obstructed bool
	Motorstop  bool
	Behaviour  Behaviour
}

type Behaviour int

const (
	Idle Behaviour = iota
	Moving
	DoorOpen
)

// can print out behaviour of elevator
func (behaviour Behaviour) ToString() string {
	switch behaviour {
	case Idle:
		return "Idle"
	case Moving:
		return "Moving"
	case DoorOpen:
		return "DoorOpen"
	default:
		return "Unknown"
	}
}

func Elevator(newOrderChannel <-chan Orders, //receiving orders
	OrderDeliveredChannel chan<- elevio.ButtonEvent,
	newLocalStateChannel chan<- State,
) {
	doorOpenChannel := make(chan bool, 16)   //creating channel to check if door is open
	doorClosedChannel := make(chan bool, 16) //creating channel to check if door is closed
	floorEnteredChannel := make(chan int)
	obstructedChannel := make(chan bool, 16)
	motorChannel := make(chan bool, 16)

	go Door(doorClosedChannel, doorOpenChannel, obstructedChannel) //starting go-routine for door
	go elevio.PollFloorSensor(floorEnteredChannel)

	elevio.SetMotorDirection(elevio.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var orders Orders

	motorTimer := time.NewTimer(configuration.WatchdogTime)
	motorTimer.Stop()

}
