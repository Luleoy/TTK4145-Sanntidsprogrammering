package configuration

import (
	"time"
	

)

const (
	NumFloors    = 4
	NumElevators = 3
	NumButtons   = 3
	Buffer       = 1024

	DisconnectTime   = 1 * time.Second
	DoorOpenDuration = 3 * time.Second
	WatchdogTime     = 5 * time.Second
	SendWVTimer      = 20 * time.Second
)

type RequestState int

const (
	None RequestState = iota
	Order
	Confirmed
	Complete
)

//legge typen i configuration. Lage kanalene de skal sendes på i main.g. structuren på hva som blir sendt på kanalen

//hva skla detligge i const, struct?

/*
type CommunicationChannels struct{
	
	elevatorID string
	peerUpdateChannel <- chan peers.PeerUpdate
	NewlocalElevatorChannel <- chan single_elevator.State
	peerTXEnableChannel chan<- bool
	
}
	//assignedRequestsChannel chan<- map[string][][2]bool, 
	// trenger worldwiew også

/*
		ElevatorID               string                         // Identifikator for heisen
		PeerUpdateChannel        <-chan peers.PeerUpdate         // Mottar oppdateringer om peers
		NewLocalElevatorChannel  <-chan single_elevator.State    // Mottar oppdateringer for den lokale heisen
		PeerTXEnableChannel      chan<- bool                    // Sender signal for å aktivere sending til peers
		AssignedRequestsChannel  chan<- map[string][][2]bool      // Sender tilordnede hallordrer (hvis du trenger dette)
		// Eventuelt flere kanaler:
		MessageTxChannel         chan<- WorldView               // Sender ut WorldView
		MessageRxChannel         <-chan WorldView               // Mottar WorldView fra andre heiser
		OrderChannel             <-chan elevio.ButtonEvent      // Mottar knappetrykk (ordrer)
		CompleteOrderChannel     <-chan fsm.DoneOrder           // Mottar signal om fullførte ordrer
		ConfirmedOrderChannel    chan WorldView                 // Sender bekreftede ordre (om nødvendig)
		MergeFSMChannel          chan<- elevio.ButtonEvent      // Sender ordrer direkte til FSM hvis heisen er alene
*/

type CommunicationChannels struct {
    ElevatorID               string
    PeerUpdateChannel        <-chan peers.PeerUpdate
    NewLocalElevatorChannel  <-chan single_elevator.State
    PeerTXEnableChannel      chan<- bool
    AssignedRequestsChannel  chan<- map[string][][2]bool
    // Eventuelt flere kanaler som:
    MessageTxChannel         chan<- WorldView
    MessageRxChannel         <-chan WorldView
    OrderChannel             <-chan elevio.ButtonEvent
    CompleteOrderChannel     <-chan fsm.DoneOrder
    ConfirmedOrderChannel    chan WorldView
    MergeFSMChannel          chan<- elevio.ButtonEvent
}