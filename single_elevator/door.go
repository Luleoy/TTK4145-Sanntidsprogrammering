package single_elevator

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/driver-go/elevio"
	"time"
)

type DoorState int

// doorstates
const (
	//fjernet Open state
	Closed DoorState = iota //iota tildeler Open int-verdien 0, inkrementerer med 1 for hver påfølgende konstant
	Obstructed
	InCountdown
)

// <- before chan means receiving channel
// <- after chan means sending channel
func Door(doorClosedChannel chan<- bool,
	doorOpenChannel <-chan bool,
	doorObstructedChannel chan<- bool) { //doorObstructedChannel: sending information about whether or not the door is blocked

	elevio.SetDoorOpenLamp(false)
	obstructionChannel := make(chan bool)               //creating a channel to check if door is obstructed
	go elevio.PollObstructionSwitch(obstructionChannel) //starting go-routine - sending true or false to obstructionChannel at all times

	obstruction := false                    //no obstruction when initializing elevator
	timeCounter := time.NewTimer(time.Hour) //creating a timer and setting to 1 hour
	var doorstate DoorState = Closed        //door is closed when initializing
	timeCounter.Stop()                      //++

	for {
		select {

		case obstruction = <-obstructionChannel: //checking if obstruction is true or false by reading from channel
			if !obstruction && doorstate == Obstructed { //if not obstruction and STATE is obstructed
				elevio.SetDoorOpenLamp(false) //close door
				doorClosedChannel <- true
				doorstate = Closed //changing STATE to closed
			}
			doorObstructedChannel <- obstruction //updating channel with obstruction status //++?

		case <-doorOpenChannel: //checking if door is open by reading from channel (if true)
			if obstruction {
				obstructionChannel <- true //send on channel
			}
			switch doorstate {
			case Closed:
				elevio.SetDoorOpenLamp(true)
				timeCounter = time.NewTimer(configuration.DoorOpenDuration)
				doorstate = InCountdown

			case InCountdown:
				timeCounter = time.NewTimer(configuration.DoorOpenDuration)

			case Obstructed:
				timeCounter = time.NewTimer(configuration.DoorOpenDuration)
				doorstate = InCountdown

			default:
				panic("Door state not implemented")
			}
		case <-timeCounter.C: //checking if time is up by reading from channel
			if doorstate != InCountdown {
				panic("Door state not implemented")
			}
			if obstruction {
				doorstate = Obstructed
			} else {
				elevio.SetDoorOpenLamp(false)
				doorClosedChannel <- true
				doorstate = Closed
			}
		}
	}
}

/*
case 1: if not obstruction and STATE is obstructed - close door and change STATE to closed
case 2: if there is a signal on the channel and if obstruction -
	switch
	case 2.1: if STATE is closed - open door, start timer and change STATE to InCountdown
	case 2.2 if STATE is InCountdown - reset timer
case 3: picking up signal from timer - if obstruction - change STATE to obstructed, else close door and change STATE to closed
*/

//må vi ha noe default?
