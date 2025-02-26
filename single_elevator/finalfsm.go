package single_elevator

import (
	"TTK4145-Heislab/configuration"
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

var timerOutChannel = make(chan bool)
func startTimer(duration configuration.DoorOpenDuration) {
	go func() {
		time.Sleep(duration)
		timerOutChannel <- true
	}()
}


//WHERE TO UPDATE FLOOR - updating on channel at all times.
//resetting av TIMER ved dooropen ++++

func SingleElevator(
	newOrderChannel <-chan Orders, //receiving new orders FROM ORDER MANAGER
	completedOrderChannel chan<- elevio.ButtonEvent, //sending information about completed orders TO ORDER MANAGER
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
	var completedOrdersList [][]int //kolonne 1 er floor, kolonne 2 er button 

	//timeroutchannel - må sende en verdi til den noe sted 

	for {
		//Watchdog??
		select {
		case <-timerOutChannel:
			switch state.Behaviour {
			case DoorOpen:
				elevio.SetDoorOpenLamp(true)
				startTimer(configuration.DoorOpenDuration)
				//hvis på vei nedover og ser at order above, kommer den nå til å utføre det.
				if ordersAbove(OrderMatrix, state.Floor) || state.Direction == Up {
					elevio.SetMotorDirection(elevio.MD_Up)
					state.Behaviour = Moving
				}
				if ordersBelow(OrderMatrix, state.Floor) || state.Direction == Down {
					elevio.SetMotorDirection(elevio.MD_Down)
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
			//OBSTRUCTION MÅ HÅNDTERE ALT 
		case obstr := <-obstructedChannel:
			// updatedState = gotNewObstruction(state, obstr);
			state.Obstructed = obstr
			switch state.Behaviour {
			case Moving:
				continue
			case DoorOpen:
				if obstr {
					elevio.SetDoorOpenLamp(true)
					startTimer(configuration.DoorOpenDuration)
				}
			case Idle:
				continue
			}
			//CASE OBSTRUCTED 
		case state.Floor = <-floorEnteredChannel: //if order at current floor
			elevio.SetFloorIndicator(state.Floor)
			//sjekker om vi har bestillinger i state.Floor, if yes. stop. and clear floor orders

			switch state.Behaviour {
			case Moving:
				if orderHere(OrderMatrix, state.Floor) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					completedOrdersList = OrderCompletedatCurrentFloor(state.Floor, Direction(state.Direction.convertMD()))
					Door() //WE NEED TO FIX THE DOOR
					//requests cleared 
					state.Behaviour = DoorOpen
					//oppdatere sånn at vi kan sende på kanalen at completedorder 
					for completedOrder in completedOrdersList {
						completedOrderChannel <- completedOrder
					}
				}
			default:
			}
		case newOrder := <-newOrderChannel:
			//her håndterer vi hvordan motoren skal kjøre og i hvilken retning etc. Buttonpressed maybe baby
			//få heisen til å gå til order fra order matrix - se hva fra timer vi kan bruke i denne istedenfor 
		}
	}
}


//DOOR - SETTE PÅ DOOROPEN LAMP OG STARTE EN TIMER 
//OBSTRUCTION I FSM MÅ HÅNDTERE DET 