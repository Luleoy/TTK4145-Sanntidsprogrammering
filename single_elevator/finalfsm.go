package single_elevator

import (
	"TTK4145-Heislab/driver-go/elevio"
)

type State struct { //the elevators current state
	Floor      int
	Direction  Direction //directions: Up, Down
	Obstructed bool
	Behaviour  Behaviour //behaviours: Idle, Moving and DoorOpen
}

type Behaviour int

const (
	Idle Behaviour = iota
	Moving
	DoorOpen //completing current order at requested floor
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

//WHERE TO UPDATE FLOOR *********
//resetting av TIMER ved dooropen ++++

func SingleElevator(
	newOrderChannel <-chan Orders, //receiving new orders FROM ORDER MANAGER
	OrderDeliveredChannel chan<- elevio.ButtonEvent, //sending information about completed orders TO ORDER MANAGER
	newLocalStateChannel chan<- State, //sending information about the elevators current state TO ORDER MANAGER
) {

	//creating channels for communication
	resetTimerChannel := make(chan bool)
	timerOutChannel := make(chan bool)
	floorEnteredChannel := make(chan int) //tells which floor elevator is at
	obstructedChannel := make(chan bool, 16)

	//starting go-routines for foor and floorsensor
	go Door(obstructedChannel, timerOutChannel, resetTimerChannel) //MÅ FIKSES OPP I
	go elevio.PollFloorSensor(floorEnteredChannel)

	//initializing elevator to go down
	elevio.SetMotorDirection(elevio.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var OrderMatrix Orders //matrix for orders

	for {
		//Watchdog??
		select {
		case <-timerOutChannel:
			switch state.Behaviour {
			case DoorOpen:
				elevio.SetDoorOpenLamp(false)
				//hvis på vei nedover og ser at order above, kommer den nå til å utføre det.
				if ordersAbove(OrderMatrix, state.Floor) || state.Direction == Up {
					elevio.SetMotorDirection(1)
					state.Behaviour = Moving
				}
				if ordersBelow(OrderMatrix, state.Floor) || state.Direction == Down {
					elevio.SetMotorDirection(-1)
					state.Behaviour = Moving
				}
				if orderHere(OrderMatrix, state.Floor) {
					state.Behaviour = DoorOpen
				} else {
					state.Behaviour = Idle
				}
			case Moving:
				// what? crash program???
			}

		case obstr := <-obstructedChannel:
			// updatedState = gotNewObstruction(state, obstr);
			state.Obstructed = obstr
			switch state.Behaviour {
			case Moving:
				continue
			case DoorOpen:
				if obstr {
					resetTimerChannel <- true
				}
			case Idle:
				continue
			}

		}
	}
}
