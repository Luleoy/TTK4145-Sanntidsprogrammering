package communication

import (
	"fmt"
	"TTK4145-Heislab/Network-go/network/peers"
	"TTK4145-Heislab/single_elevator"
	"TTK4145-Heislab/configuration"
)

func CommunicationHandler(chans configuration.CommunicationChannels) {

	//initialisering
	numPeers := 0


	for {
	
		select {
		//case_ 5: Oppdateringer for den lokale heisen
		case newLocalElevator := <-newLocalElevatorChannel: 
			localWorldView.ElevatorList[elevatorID] = newLocalElevator
			cabReq := GetCabRequests(newLocalElevator)

			// Lagre bestillingene til en JSON-fil som en backup. mby
			//saveJSON.SaveCabButtonToFile(cabReq, "cabOrders.json")
		

		//Case 6:
		//oppdatere på hvilke heiser som er aktive ( når heiser kommer på og forsvinner fra nettverket)
		case peers := <-peerUpdateChannel: 
			
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peers.Peers)
			fmt.Printf("  New:      %q\n", peers.New)
			fmt.Printf("  Lost:     %q\n", peers.Lost)
			
			numPeers = len(peers.Peers)

			if len(peers.Lost) > 0 {
				if localWorldView.ElevatorList[peers.Lost[0]].Unavailable {
					AssignOrder(*localWorldView, ch_assignedRequests)
					peerTXEnableChannel <- true
				} else {
					for i, ack := range localWorldView.AckList {
						for _, lostPeer := range peers.Lost {
							delete(localWorldView.ElevatorList, lostPeer)
							if ack == lostPeer {
								localWorldView.AckList = append(localWorldView.AckList[:i], localWorldView.AckList[i+1:]...)
							}
						}
					}
					AssignOrder(*localWorldView, assignedRequestsChannel)
				}
		}
		}
	}
}
