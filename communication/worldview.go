package communication

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/single_elevator"
	"time"
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

InitializeWorldView(elevatorID string) WorldView {
	message := WorldView{
		Counter: 0, 
		ID: elevatorID, 
		AckList: make([]string, 0), 
		ElevatorStatusList: map[string]single_elevator.State{elevatorID: single_elevator.State}
		HallOrderStatus

		//CabRequests? 
	}
	return message 
}

func WorldViewHandler(
	// channels som må tas inn
	elevatorID string,
	WorldViewTXChannel chan<- WorldView, //WV transmitter
	WorldViewRXChannel <-chan WorldView, //WV receiver
) {

	for {
	OuterLoop: //break ut av hele for-loopen
		select {
		
			case <-SendLocalWorldViewTimer.C: //local world view updates
			localWorldView.ID = elevatorID
			WorldViewTXChannel <- *localWorldView
			SendLocalWorldViewTimer.Reset(time.Duration(configuration.SendWVTimer) * time.Millisecond)

		case buttonPressed := <-buttonPressedChannel: //knappetrykk. tar inn button events. Dette er neworder. Må skille fra Neworderchannel i single_elevator. sjekk ut hvor den skal defineres etc
		case updatedWorldView := <-WorldViewRXChannel: //meldingssystemet - connection med Network
		//case complete := <-completedOrderChannel: //kommer fra Single Elevator order completed channel?? LINK
		//endre navn i single elevator til completedSingleOrderChannel
		
		}
	}
}


	//initialize local world view to send on message channel
	initLocalWorldView := InitializeWorldView(elevatorID)
	localWorldView := &initLocalWorldView //bruke localworldview i casene fremover - kopiere worldview

	//timer for når Local World View skal oppdateres
	SendLocalWorldViewTimer := time.NewTimer(time.Duration(configuration.SendWVTimer) * time.Millisecond)


