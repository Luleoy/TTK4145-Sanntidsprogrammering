package main

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"TTK4145-Heislab/single_elevator"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Elevator System Starting...")

	// Initialize elevator hardware
	numFloors := configuration.NumFloors
	elevio.Init("localhost:15657", numFloors)

	// Global OrderMatrix (bruker peker for å unngå kopiering)
	orderMatrix := &single_elevator.Orders{}

	// Communication channels
	newOrderChannel := make(chan *single_elevator.Orders, configuration.Buffer)
	orderDeliveredChannel := make(chan elevio.ButtonEvent, configuration.Buffer)
	newLocalStateChannel := make(chan single_elevator.State, configuration.Buffer)

	// Polling channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Start FSM
	go single_elevator.Elevator(newOrderChannel, orderDeliveredChannel, newLocalStateChannel)

	// Start polling inputs
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Ensure Elevator Starts at a Valid Floor
	go func() {
		select {
		case floor := <-drv_floors:
			elevio.SetFloorIndicator(floor)
			fmt.Println("Elevator initialized at floor:", floor)
		default:
			// Hvis vi ikke vet startposisjon, gå ned til vi finner en etasje
			for {
				select {
				case floor := <-drv_floors:
					elevio.SetFloorIndicator(floor)
					fmt.Println("Elevator initialized at floor:", floor)
					return
				default:
					elevio.SetMotorDirection(elevio.MD_Down)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()

	// Monitor Elevator State for Debugging
	go func() {
		for state := range newLocalStateChannel {
			fmt.Printf("State Update: Floor %d, Direction %v, Behaviour %v\n",
				state.Floor, state.Direction, state.Behaviour)
		}
	}()

	// Handle Button Presses
	go func() {
		for btn := range drv_buttons {
			fmt.Printf("Button Pressed: Floor %d, Button %v\n", btn.Floor, btn.Button)

			// Oppdater eksisterende OrderMatrix
			orderMatrix[btn.Floor][btn.Button] = true

			// Send den oppdaterte OrderMatrix til FSM
			newOrderChannel <- orderMatrix
		}
	}()

	// Handle Order Completion
	go func() {
		for completedOrder := range orderDeliveredChannel {
			fmt.Printf("Order Completed: Floor %d, Button %v\n", completedOrder.Floor, completedOrder.Button)
		}
	}()

	// Handle Obstruction
	go func() {
		for obstructed := range drv_obstr {
			fmt.Printf("Obstruction: %v\n", obstructed)
			// Forward obstruction event to the FSM
			// Dette krever at FSM har en obstruction-kanal implementert
		}
	}()

	// Handle Stop Button
	go func() {
		for range drv_stop {
			fmt.Println("Stop Button Pressed")
			elevio.SetMotorDirection(elevio.MD_Stop)

			// Nullstill alle ordre i OrderMatrix
			*orderMatrix = single_elevator.Orders{} // Setter alle verdier til false
			newOrderChannel <- orderMatrix
		}
	}()

	// Keep main alive
	select {}
}
