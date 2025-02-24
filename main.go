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
		for {
			select {
			case floor := <-drv_floors:
				elevio.SetFloorIndicator(floor)
				fmt.Println("Elevator initialized at floor:", floor)
				return
			default:
				elevio.SetMotorDirection(elevio.MD_Down) // Move down until reaching a floor
				time.Sleep(100 * time.Millisecond)
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

	// Keep main alive
	select {}
}
