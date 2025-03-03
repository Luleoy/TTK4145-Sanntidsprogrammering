package communication

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/single_elevator"
)

// oppdaterer newlocalstate i single_elevator FSM
type WorldView struct {
	State       single_elevator.State
	CabRequests [configuration.NumFloors]bool
}
