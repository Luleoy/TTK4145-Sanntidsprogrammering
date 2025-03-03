package distributor

import (
	"TTK4145-Heislab/configuration"
	"TTK4145-Heislab/single_elevator"
)

type LocalState struct {
	State       single_elevator.State
	CabRequests [configuration.NumFloors]bool
}
