package single_elevator // Definerer pakken order_manager

import (
	"TTK4145-Heislab/driver-go/elevio"
	"TTK4145-Heislab/configuration"
	"fmt"
)

// Definerer en struktur for å håndtere ordre
type OrderManager struct {
	OrderMatrix single_elevator.Orders      // En 2D-matrise som holder oversikt over bestillinger
	newOrderChannel  chan single_elevator.Orders // Kanal for å sende nye ordre til heisen
	CompletedCh chan elevio.ButtonEvent     // Kanal for å motta signal om at en ordre er fullført
	LocalState  chan single_elevator.State  // Kanal for å motta heisens nåværende tilstand
}


func NewOrderManager(newOrderChannel chan single_elevator.Orders, completedOrderChannel chan elevio.ButtonEvent, localState chan single_elevator.State) OrderManager {
	return OrderManager{
		OrderMatrix: single_elevator.Orders{}, // Starter med en tom ordrematrise
		newOrderChannel:  newOrderChannel,               // Setter kanalene for kommunikasjon
		completedOrderChannel: completedOrderChannel,
		LocalState:  localState,
	}
}

// **Starter Order Manager i en egen goroutine**
func (om OrderManager) Run() {
	for {
		select {
		case newOrder := <-om.ListenForNewOrders(): // Hvis en ny ordre oppdages
			om = om.AddOrder(newOrder) // Oppdater OrderManager med ny ordre
		case completedOrder := <-om.completedOrderChannel: // Hvis en ordre er fullført
			om = om.RemoveOrder(completedOrder) // Fjerner fullførte ordre
		case state := <-om.LocalState: // Hvis heisen sender en oppdatert tilstand
			om = om.HandleStateChange(state) // Håndterer endringer i heistilstand
		}
	}
}

// **Funksjon for å lytte etter nye bestillinger fra knappetrykk**
func (om OrderManager) ListenForNewOrders() <-chan elevio.ButtonEvent {
	newOrderChannel := make(chan elevio.ButtonEvent) // Oppretter en kanal for nye ordre
	go func() {
		for {
			for f := 0; f < configuration.NumFloors; f++ { // Går gjennom alle etasjer
				for b := 0; b < configuration.NumButtons; b++ { // Går gjennom alle knapper
					if elevio.GetButtonSignal(elevio.ButtonType(b), f) { // Sjekker om en knapp er trykket
						newOrderChannel <- elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}
					}
				}
			}
		}
	}()
	return newOrderChannel
}

// **Legger til en ny bestilling i ordrematrisen**
func (om OrderManager) AddOrder(order elevio.ButtonEvent) OrderManager {
	om.OrderMatrix[order.Floor][order.Button] = true // Setter ordren i matrisen
	om.NewOrderCh <- om.OrderMatrix // Sender oppdatert matrise til heisen
	fmt.Printf("Added order: Floor %d, Button %v\n", order.Floor, order.Button)
	single_elevator.SetLights(om.OrderMatrix) // Oppdaterer lysene i heisen
	return om // Returnerer oppdatert OrderManager
}

// **Fjerner en fullført ordre fra matrisen**
func (om OrderManager) RemoveOrder(order elevio.ButtonEvent) OrderManager {
	om.OrderMatrix[order.Floor][order.Button] = false // Sletter ordren fra matrisen
	fmt.Printf("Completed order: Floor %d, Button %v\n", order.Floor, order.Button)
	single_elevator.SetLights(om.OrderMatrix) // Oppdaterer lysene
	return om // Returnerer oppdatert OrderManager
}

// **Håndterer endringer i heisens tilstand**
func (om OrderManager) HandleStateChange(state single_elevator.State) OrderManager {
	// Hvis heisen er i en etasje og har en ordre der, fjern ordren
	if state.Behaviour == single_elevator.DoorOpen {
		completedOrders := single_elevator.OrderCompletedatCurrentFloor(state.Floor, state.Direction, om.OrderMatrix)
		for _, order := range completedOrders { // Går gjennom alle fullførte ordre
			om = om.RemoveOrder(elevio.ButtonEvent{Floor: order[0], Button: elevio.ButtonType(order[1])})
		}
	}
	return om // Returnerer