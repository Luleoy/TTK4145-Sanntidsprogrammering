package single_elevator

import (
	"TTK4145-Heislab/driver-go/elevio"
	"time"
)

//DOOR skal håndtere obstructions og timer

type DoorState int

// doorstates
const (
	Closed DoorState = iota //iota tildeler Open int-verdien 0, inkrementerer med 1 for hver påfølgende konstant
	Obstructed
	InCountdown
)

// <- before chan means receiving channel
// <- after chan means sending channel
func Door(obstructedChannel chan<- bool,
	timerOutChannel <-chan bool,
	resetTimerChannel <-chan bool) { //doorObstructedChannel: sending information about whether or not the door is blocked

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
				doorstate = Closed            //changing STATE to closed
			}
			obstructedChannel <- obstruction //updating channel with obstruction status //++?
		case <-timerOutChannel:
			if doorstate != InCountdown {
				panic("Door state not implemented")
			}
			if obstruction {
				doorstate = Obstructed
			} else {
				elevio.SetDoorOpenLamp(false)
				doorstate = Closed
			}
		}
	}
}
//go routine for timer 
3 sek 
timeroutchannel <- true 

/*
case 1: if not obstruction and STATE is obstructed - close door and change STATE to closed
case 2: if there is a signal on the channel and if obstruction -
	switch
	case 2.1: if STATE is closed - open door, start timer and change STATE to InCountdown
	case 2.2 if STATE is InCountdown - reset timer
case 3: picking up signal from timer - if obstruction - change STATE to obstructed, else close door and change STATE to closed
*/

//må vi ha noe default?
