package main

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"TTK4145-Heislab/single_elevator"
	"fmt"
)

func main() {
	fmt.Println("Started yeah")

	//port
	//heis ID

	numFloors := configuration.NumFloors
	elevio.Init("localhost:15657", numFloors)

	newOrderChannel := make(chan single_elevator.Orders, configuration.Buffer)
	OrderDeliveredChannel := make(chan elevio.ButtonEvent, configuration.Buffer)
	newLocalStateChannel := make(chan single_elevator.State, configuration.Buffer)

	// Polling channels (from `driver_main`)
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go single_elevator.Elevator(newOrderChannel, OrderDeliveredChannel, newLocalStateChannel)
	fmt.Println("Go routine started yeah")

	// Start polling inputs (from `driver_main`)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Handle inputs from sensors and buttons
	go func() {
		for {
			select {
			case a := <-drv_buttons:
				fmt.Printf("Button Pressed: %+v\n", a)
				var orders single_elevator.Orders
				orders[a.Floor][a.Button] = true
				newOrderChannel <- orders
				elevio.SetButtonLamp(a.Button, a.Floor, true)

			case floor := <-drv_floors:
				fmt.Printf("Floor Sensor Triggered: %+v\n", floor)
				elevio.SetFloorIndicator(floor)

			case obstructed := <-drv_obstr:
				fmt.Printf("Obstruction Switch: %+v\n", obstructed)
				if obstructed {
					elevio.SetMotorDirection(elevio.MD_Stop)
				}

			case stop := <-drv_stop:
				fmt.Printf("Stop Button Pressed: %+v\n", stop)
				for f := 0; f < numFloors; f++ {
					for b := elevio.ButtonType(0); b < 3; b++ {
						elevio.SetButtonLamp(b, f, false)
					}
				}
			}
		}
	}()

	// Monitor elevator state updates
	go func() {
		for state := range newLocalStateChannel {
			fmt.Printf("Elevator State Updated: Floor %d, Direction %v, Behaviour %v\n",
				state.Floor, state.Direction, state.Behaviour.ToString())
		}
	}()

	// Prevent `main.go` from exiting
	select {}
}
