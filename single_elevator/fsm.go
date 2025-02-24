package single_elevator

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"fmt"
	"time"
)

type State struct { //the elevators current state
	Floor      int
	Direction  Direction
	Obstructed bool
	Motorstop  bool
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

func Elevator(
	newOrderChannel <-chan *Orders, //receiving new orders
	OrderDeliveredChannel chan<- elevio.ButtonEvent, //sending information about completed orders
	newLocalStateChannel chan<- State, //sending information about the elevators current state
) {
	//creating channels for communication
	doorOpenChannel := make(chan bool, 16)   //creating channel to check if door is open
	doorClosedChannel := make(chan bool, 16) //creating channel to check if door is closed
	floorEnteredChannel := make(chan int)    //tells which floor elevator is at
	obstructedChannel := make(chan bool, 16)
	motorChannel := make(chan bool, 16) //motorstatus

	//starting go-routines for foor and floorsensor
	go Door(doorClosedChannel, doorOpenChannel, obstructedChannel)
	go elevio.PollFloorSensor(floorEnteredChannel)

	//initializing elevator to go down
	elevio.SetMotorDirection(elevio.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	//må egt vite hvilken etasje heisen er i når den starter

	//var OrderMatrix Orders //matrix for orders

	OrderMatrix := new(Orders)

	motorTimer := time.NewTimer(configuration.WatchdogTime) //creating a watchdog timer
	motorTimer.Stop()

	/*
		SELECT CASES: listening to several channels at the same time. When receiving signal on one of the channels - do the corresponding case
		The elevator has to react to outside signals no matter what state the elevator is in - therefore the select cases are outside the switch cases (fsm) inside
		checking channels for communication: doorOpenChannel, doorClosedChannel, floorEnteredChannel, obstructedChannel, motorChannel
		1. case: doorClosedChannel. the door has closed
			switch: going through state behaviours: Idle, Moving and DoorOpen
			case 1: door is open. DoorOpen
				switch
				case 1 order in current direction. Moving
				case 2 order in opposite direction. Moving
				case 3 no orders. Idle
				default Idle
			default panic

		2. case: floorEnteredChannel
			switch: going through state behaviours: Idle, Moving and DoorOpen
			case 1: Elevator is moving to a new floor. Moving
				switch
				case 1 completes orders on current floor and direction. DoorOpen
				case 2 completes orders on current floor and direction - CAB. DoorOpen
				case 3 completes orders on current floor and direction - CAB and no orders in opposite direction. DoorOpen
				case 4 checking order in current direction. Moving
				case 5 checking if order in opposite direction from current floor. completes order in current floor. DoorOpen
				case 6 checking if order in opposite direction from current floor. Moving
				default Motor stops. Idle
			default panic

		3. case: newOrderChannel
			switch going through state behaviours: Idle, Moving and DoorOpen
			case 1: Idle
				switch
				case 1 orders at current floor and direction OR CAB. DoorOpen
				case 2 orders at current floor and opposite direction. DoorOpen
				case 3 orders in current direction. Moving
				case 4 orders in opposite direction from current floor. Moving
				default
			case 2: DoorOpen
				switch
				case 1 orders current floor and CAB OR orders at current floor and direction. Complete order.
			case 3: Moving
			default panic

		4. case: motorTimer.C. Lost motor power

		5. case: ObstructedChannel. Checking if door is obstructed

		6. case: motorChannel. Regained motor power

	*/

	for {
		select {

		case <-doorClosedChannel:
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction):
					elevio.SetMotorDirection(state.Direction.convertMD())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(configuration.WatchdogTime)
					motorChannel <- false
					newLocalStateChannel <- state

				case OrderMatrix[state.Floor][state.Direction.invertMD()]:
					doorOpenChannel <- true
					state.Direction = state.Direction.invertMD()
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					newLocalStateChannel <- state

				case OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction.invertMD()):
					state.Direction = state.Direction.invertMD()
					elevio.SetMotorDirection(state.Direction.convertMD())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(configuration.WatchdogTime)
					motorChannel <- false
					newLocalStateChannel <- state

				default:
					state.Behaviour = Idle
					newLocalStateChannel <- state
				}
			default:
				panic("DoorClosedChannel received in wrong state")
			}
		case state.Floor = <-floorEnteredChannel:
			elevio.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorChannel <- false
			switch state.Behaviour {
			case Moving:
				switch {
				case (*OrderMatrix)[state.Floor][state.Direction]:
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenChannel <- true
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					state.Behaviour = DoorOpen

				case (*OrderMatrix)[state.Floor][elevio.BT_Cab] && OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenChannel <- true
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					state.Behaviour = DoorOpen

				case (*OrderMatrix)[state.Floor][elevio.BT_Cab] && !OrderMatrix[state.Floor][state.Direction.invertMD()]:
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenChannel <- true
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					state.Behaviour = DoorOpen

				case OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction):
					motorTimer = time.NewTimer(configuration.WatchdogTime)
					motorChannel <- false

				case (*OrderMatrix)[state.Floor][state.Direction.invertMD()]:
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenChannel <- true
					state.Direction = state.Direction.invertMD()
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					state.Behaviour = DoorOpen

				case OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction.invertMD()):
					state.Direction = state.Direction.invertMD()
					elevio.SetMotorDirection(state.Direction.convertMD())
					motorTimer = time.NewTimer(configuration.WatchdogTime)
					motorChannel <- false

				default:
					elevio.SetMotorDirection(elevio.MD_Stop)
					state.Behaviour = Idle
				}
			default:
				panic("FloorEnteredChannel received in wrong state")
			}
			newLocalStateChannel <- state

		case <-newOrderChannel: //newOrders needed to receive data from the channeø and update ordermatrix. receive neworders from the channel, copy its values into ordermatrix, use ordermatrix for the rest of the code
			newOrders := <-newOrderChannel
			*OrderMatrix = *newOrders

			switch state.Behaviour {
			case Idle:
				switch {
				case (*OrderMatrix)[state.Floor][state.Direction] || (*OrderMatrix)[state.Floor][elevio.BT_Cab]:
					doorOpenChannel <- true
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					state.Behaviour = DoorOpen
					newLocalStateChannel <- state
				case (*OrderMatrix)[state.Floor][state.Direction.invertMD()]:
					doorOpenChannel <- true
					state.Direction = state.Direction.invertMD()
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
					state.Behaviour = DoorOpen
					newLocalStateChannel <- state
				case OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction):
					elevio.SetMotorDirection(state.Direction.convertMD())
					state.Behaviour = Moving
					newLocalStateChannel <- state
					motorTimer = time.NewTimer(configuration.WatchdogTime)
					motorChannel <- false
				case OrderMatrix.OrderinCurrentDirection(state.Floor, state.Direction.invertMD()):
					state.Direction = state.Direction.invertMD()
					elevio.SetMotorDirection(state.Direction.convertMD())
					state.Behaviour = Moving
					newLocalStateChannel <- state
					motorTimer = time.NewTimer(configuration.WatchdogTime)
					motorChannel <- false
				default:
				}
			case DoorOpen:
				switch {
				case (*OrderMatrix)[state.Floor][elevio.BT_Cab] || (*OrderMatrix)[state.Floor][state.Direction]:
					doorOpenChannel <- true
					OrderCompleted(state.Floor, state.Direction, OrderMatrix, OrderDeliveredChannel)
				}

			case Moving:
			default:
				panic("OrderMatrix received in wrong state")
			}
		case <-motorTimer.C:
			if !state.Motorstop {
				fmt.Println("Lost connection to motor")
				state.Motorstop = true
				newLocalStateChannel <- state
			}
		case obstruction := <-obstructedChannel:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				newLocalStateChannel <- state
			}
		case motor := <-motorChannel:
			if state.Motorstop {
				fmt.Println("Connection to motor restored")
				state.Motorstop = motor
				newLocalStateChannel <- state
			}
		}
	}
}
