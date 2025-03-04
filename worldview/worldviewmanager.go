package worldview

//TRENGER FRA COMMUNICATION: numPeers

import (
	"TTK4145-Heislab/driver-go/elevio"
	//sirkelkomposisjon med communication
	"TTK4145-Heislab/single_elevator"
)

func WorldViewHandler(
	// channels som må tas inn
	elevatorID string,
	WorldViewTXChannel chan<- WorldView, //WV transmitter
	WorldViewRXChannel <-chan WorldView, //WV receiver
	buttonPressedChannel <-chan elevio.ButtonEvent,
	mergeChannel chan<- elevio.ButtonEvent,

) {

	for {
	OuterLoop: //break ut av hele for-loopen
		select {
		case buttonPressed := <-buttonPressedChannel: //knappetrykk. tar inn button events. Dette er neworder. Må skille fra Neworderchannel i single_elevator. sjekk ut hvor den skal defineres etc
			//1 heis - sende til single elevator
			if numPeers == 1 {
				single_elevator.OrderManager(newOrderChannel, completedOrderChannel, buttonPressed) //må endre på single elevator - Ikke neworderchannel og completedorderchannel 
		

			}
		//CAB - sende til single elevator
		//Hall - Assigner - localWorldView, counter økes
		//Acklist nullstilles

		case updatedWorldView := <-WorldViewRXChannel: //meldingssystemet - connection med Network
			//case complete := <-completedOrderChannel: //kommer fra Single Elevator order completed channel?? LINK
			//endre navn i single elevator til completedSingleOrderChannel

		}
	}
}


//buttonpressedchannel - hvordan hente ut
//updating states in FSM??


order manager for single elevator
func OrderManager(newOrderChannel chan<- Orders,
	completedOrderChannel <-chan elevio.ButtonEvent, //sende-kanal
	//newLocalStateChannel <-chan State, //sende-kanal - NÅR SKAL DENNE BRUKES?
	buttonPressedChannel <-chan elevio.ButtonEvent) { //kun lesing av kanal
	OrderMatrix := [configuration.NumFloors][configuration.NumButtons]bool{}
	for {
		select {
		case buttonPressed := <-buttonPressedChannel:
			OrderMatrix[buttonPressed.Floor][buttonPressed.Button] = true
			SetLights(OrderMatrix)
			newOrderChannel <- OrderMatrix
		case ordercompletedbyfsm := <-completedOrderChannel:
			OrderMatrix[ordercompletedbyfsm.Floor][ordercompletedbyfsm.Button] = false
			SetLights(OrderMatrix)
			newOrderChannel <- OrderMatrix
		}
	}