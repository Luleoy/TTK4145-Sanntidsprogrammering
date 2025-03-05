package worldview

//TRENGER FRA COMMUNICATION: numPeers

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"TTK4145-Heislab/single_elevator"
	"time"
)

// REQUESTSTATE INT - lage en funksjon som konverterer fra WorldView til OrderMatrix som vi kan bruke
type RequestState int

const (
	None RequestState = iota
	Order
	Confirmed
	Complete
)

// oppdaterer newlocalstate i single_elevator FSM
type WorldView struct {
	Counter int
	ID      string
	Acklist []string

	ElevatorStatusList map[string]single_elevator.State //legger inn all info om hver heis; floor, direction, obstructed, behaviour
	HallOrderStatus    [][configuration.NumButtons - 1]configuration.RequestState

	//burde vi broadcaste dette? ja
	CabRequests [configuration.NumFloors]bool //hvis vi broadcaster cab buttons her, slipper vi å lagre alt over i en fil
	//om heisen er i live eller ikke/unavailable
}

func InitializeWorldView(elevatorID string) WorldView {
	message := WorldView{
		Counter:            0,
		ID:                 elevatorID,
		Acklist:            make([]string, 0),
		ElevatorStatusList: map[string]single_elevator.State{elevatorID: single_elevator.State},
		HallOrderStatus:    InitializeHallOrderStatus(),
		//CabRequests?
	}
	return message
}

// funksjon som skal initialisere hallorderstatus. skal etterhvert ha false på alle utenom confirmed med true
func InitializeHallOrderStatus() [][configuration.NumButtons - 1]configuration.RequestState {
	hallOrderStatus := make([][configuration.NumButtons - 1]configuration.RequestState, configuration.NumFloors)
	for floor := range hallOrderStatus {
		for button := range hallOrderStatus[floor] {
			hallOrderStatus[floor][button] = configuration.None
		}
	}
	return hallOrderStatus
}

func ResetAckList(localWorldView *WorldView) {
	localWorldView.Acklist = make([]string, 0)
	localWorldView.Acklist = append(localWorldView.Acklist, localWorldView.ID)
}

func WorldViewHandler(
	// channels som må tas inn
	elevatorID string,
	WorldViewTXChannel chan<- WorldView, //WorldView transmitter
	WorldViewRXChannel <-chan WorldView, //WorldView receiver
	buttonPressedChannel <-chan elevio.ButtonEvent,
	mergeChannel chan<- elevio.ButtonEvent,
	//newOrderChannel chan<- Orders - fra OrderManager
	completedOrderChannel <-chan elevio.ButtonEvent,
	numPeersChannel <-chan int,

) {

	//initialize local world view to send on message channel
	initLocalWorldView := InitializeWorldView(elevatorID)
	localWorldView := &initLocalWorldView //bruke localworldview i casene fremover - kopiere worldview

	//timer for når Local World View skal oppdateres
	SendLocalWorldViewTimer := time.NewTimer(time.Duration(configuration.SendWVTimer) * time.Millisecond)
	numPeers:=0

	for {
	OuterLoop: //break ut av hele for-loopen
		select {
		case num:= <-numPeersChannel:
			numPeers = num
		
		case <-SendLocalWorldViewTimer.C: //local world view updates
			localWorldView.ID = elevatorID
			WorldViewTXChannel <- *localWorldView
			SendLocalWorldViewTimer.Reset(time.Duration(configuration.SendWVTimer) * time.Millisecond)

		case buttonPressed := <-buttonPressedChannel: //knappetrykk. tar inn button events. Dette er neworder. Må skille fra Neworderchannel i single_elevator. sjekk ut hvor den skal defineres etc
			//1 heis - sende til single elevator DENNE MÅ OPPDATERES NÅR
			if numPeers == 1 {
				OrderMatrix := [configuration.NumFloors][configuration.NumButtons]bool{}
				for {
					select {
					case buttonPressed := <-buttonPressedChannel: //har ikke tatt inn alt fra ordermanager - må revisit
						OrderMatrix[buttonPressed.Floor][buttonPressed.Button] = true
						SetLights(OrderMatrix)
						newOrderChannel <- OrderMatrix
					case ordercompletedbyfsm := <-completedOrderChannel:
						OrderMatrix[ordercompletedbyfsm.Floor][ordercompletedbyfsm.Button] = false
						SetLights(OrderMatrix)
						newOrderChannel <- OrderMatrix
					}
				}
			}
			localWorldView.HallOrderStatus[buttonPressed.Floor][int(buttonPressed.Button)] = configuration.Order
			localWorldView.Counter++
			ResetAckList(localWorldView)

		//MESSAGE SYSTEM
		case updatedWorldView := <-WorldViewRXChannel: //meldingssystemet - connection med Network
			if localWorldView.Counter >= updatedWorldView.Counter {
				if localWorldView.Counter == updatedWorldView.Counter && len(localWorldView.Acklist) < len(updatedWorldView.Acklist) {
					localElevatorStatus := localWorldView.ElevatorStatusList[elevatorID]
					localWorldView = &updatedWorldView
					localWorldView.ElevatorStatusList[elevatorID] = localElevatorStatus
				} else {
					break Outerloop
				}
			}
			//lys??
			if len(updatedWorldView.Acklist) == numPeers {
				for floor := 0; floor < configuration.NumFloors; floor++ {
					for button := 0; button < configuration.NumButtons-1; button++ {
						switch {
						case updatedWorldView.HallOrderStatus[floor][button] == configuration.Order: //updating state in hallorderstatus
							localWorldView.HallOrderStatus[floor][button] = configuration.Confirmed
							localWorldView.Counter = updatedWorldView.Counter
							localWorldView.Counter++
							ResetAckList(localWorldView)
						case updatedWorldView.HallOrderStatus[floor][button] == configuration.Confirmed && !HallOrderDistributed[floor][button]:
							//case må fylles inn
						case updatedWorldView.HallOrderStatus[floor][button] == configuration.Complete:
							localWorldView.HallOrderStatus[floor][button] = configuration.None
							HallOrderDistributed[floor][button] = false
							localWorldView.Counter++
						}
					}
				}
			}
			//case complete := <-completedOrderChannel: //kommer fra Single Elevator order completed channel?? LINK
			//endre navn i single elevator til completedSingleOrderChannel? - må ha noe ordercompleted channel

		}
	}
}

//buttonpressedchannel - hvordan hente ut
//updating states in FSM??
