package single_elevator

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"fmt"
	"time"
)

// MÅ ENDRE NAVN FRA STATE TIL ELEVATOR: STATE ER MISVISENDE
type State struct { //the elevators current state
	Floor      int
	Direction  Direction //directions: Up, Down
	Obstructed bool
	Behaviour  Behaviour //behaviours: Idle, Moving and DoorOpen
	unavailable bool
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

func runTimer(duration time.Duration, timeOutChannel chan<- bool, resetTimerChannel <-chan bool) {
	deadline := time.Now().Add(100000 * time.Hour)
	is_running := false

	for {
		select {
		case <-resetTimerChannel:
			deadline = time.Now().Add(duration)
			is_running = true
		default:
			if is_running && time.Since(deadline) > 0 {
				timeOutChannel <- true
				is_running = false
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}

//WHERE TO UPDATE FLOOR - updating on channel at all times.

func SingleElevator(
	newOrderChannel <-chan Orders, //receiving new orders FROM ORDER MANAGER
	completedOrderChannel chan<- elevio.ButtonEvent, //sending information about completed orders TO ORDER MANAGER
	newLocalStateChannel chan<- State, //sending information about the elevators current state TO ORDER MANAGER
) {


	//Initialization of elevator
	fmt.Println("setting motor down")
	elevio.SetMotorDirection(elevio.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	//creating channels for communication
	floorEnteredChannel := make(chan int) //tells which floor elevator is at
	obstructedChannel := make(chan bool, 16)
	stopPressedChannel := make(chan bool, 16)

	//starting go-routines for foor and floorsensor
	go elevio.PollFloorSensor(floorEnteredChannel)

	timerOutChannel := make(chan bool)
	resetTimerChannel := make(chan bool)
	go runTimer(configuration.DoorOpenDuration, timerOutChannel, resetTimerChannel)
	// go startTimer(configuration.DoorOpenDuration, timerOutChannel)

	/*
		resetWatchdogChannel := make(chan bool)
		go func() {
			timeout := 3 * time.Second
			deadline := time.Now().Add(timeout)

			for {
				select {
				case <-resetWatchdogChannel:
					deadline = time.Now().Add(timeout)
				default:
					if time.Now().After(deadline) {
						fmt.Println("Watchdog timer expired! Restarting elevator")
						go SingleElevator(newOrderChannel, completedOrderChannel, newLocalStateChannel)
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}() */

	go elevio.PollObstructionSwitch(obstructedChannel)
	go elevio.PollStopButton(stopPressedChannel)

	var OrderMatrix Orders //matrix for orders

	for {
		//Watchdog??
		select {
		case <-timerOutChannel: //timeroutchannel - må sende en verdi til den noe sted!!
			//resetWatchdogChannel <- true
			switch state.Behaviour {
			case DoorOpen:
				DirectionBehaviourPair := ordersChooseDirection(state.Floor, state.Direction, OrderMatrix)
				state.Behaviour = DirectionBehaviourPair.Behaviour
				state.Direction = Direction(DirectionBehaviourPair.Direction)
				newLocalStateChannel <- state
				switch state.Behaviour {
				case DoorOpen:
					//start timer på nytt og rydd forespørsler i nåværende etasje
					resetTimerChannel <- true
					OrderCompletedatCurrentFloor(state.Floor, Direction(state.Direction.convertMD()), completedOrderChannel) //requests cleared
					newLocalStateChannel <- state
				case Moving, Idle:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(DirectionBehaviourPair.Direction)

				}
			case Moving:
				panic("timeroutchannel in moving")
			}
		case stopbuttonpressed := <-stopPressedChannel:
			//resetWatchdogChannel <- true
			if stopbuttonpressed {
				fmt.Println("StopButton is pressed")
				elevio.SetStopLamp(true)
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetStopLamp(false)
			}
		case state.Obstructed = <-obstructedChannel:
			//resetWatchdogChannel <- true
			switch state.Behaviour {
			case DoorOpen:
				resetTimerChannel <- true
				fmt.Println("Obstruction switch ON")
				newLocalStateChannel <- state //NEW LOCAL STATE MÅ OPPDATERES OVERALT
			case Moving, Idle:
				continue
			}
		case state.Floor = <-floorEnteredChannel: //if order at current floor
			//resetWatchdogChannel <- true
			fmt.Println("New floor: ", state.Floor)
			elevio.SetFloorIndicator(state.Floor)
			//sjekker om vi har bestillinger i state.Floor, if yes. stop. and clear floor orders
			switch state.Behaviour {
			case Moving:
				if orderHere(OrderMatrix, state.Floor) || state.Floor == 0 || state.Floor == configuration.NumFloors-1 {
					elevio.SetMotorDirection(elevio.MD_Stop)
					OrderCompletedatCurrentFloor(state.Floor, Direction(state.Direction.convertMD()), completedOrderChannel) //requests cleared
					resetTimerChannel <- true
					state.Behaviour = DoorOpen
					newLocalStateChannel <- state
					fmt.Println("New local state:", state)
				}
			default:
			}
		case OrderMatrix = <-newOrderChannel:
			//resetWatchdogChannel <- true
			fmt.Println("New orders :)")
			switch state.Behaviour {
			case Idle:
				state.Behaviour = Moving
				DirectionBehaviourPair := ordersChooseDirection(state.Floor, state.Direction, OrderMatrix)
				state.Behaviour = DirectionBehaviourPair.Behaviour
				state.Direction = Direction(DirectionBehaviourPair.Direction)
				newLocalStateChannel <- state
				//elevio.SetMotorDirection(DirectionBehaviourPair.Direction)
				switch state.Behaviour {
				case DoorOpen:
					//start timer på nytt og rydd forespørsler i nåværende etasje
					resetTimerChannel <- true
					OrderCompletedatCurrentFloor(state.Floor, Direction(state.Direction.convertMD()), completedOrderChannel) //requests cleared
					newLocalStateChannel <- state
				case Moving, Idle:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(DirectionBehaviourPair.Direction)
				}
			}
		}
		elevio.SetDoorOpenLamp(state.Behaviour == DoorOpen)
	}
}

/*
Hva må ryddes opp i:
watchdogtimer? - hakker, og går out of range
default/panic bør det implementeres over alt?
obstruction - ??
doesnt know its in between two floors when stopping in between two floors
printer new orders selv om vi ikke har noen orders?
*/
