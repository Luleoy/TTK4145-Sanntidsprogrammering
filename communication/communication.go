package communication

func CommunicationHandler(
	// channels som må tas inn
	elevatorID string,
) {

	for {
	OuterLoop: //break ut av hele for-loopen
		select {

		case newLocalElevator := <-newLocalElevatorChannel: //heis forsvunnet og kommet tilbake på nettverket
		case peers := <-peerUpdateChannel: //oppdatere på hvilke heiser som er aktive
		}
	}
}
