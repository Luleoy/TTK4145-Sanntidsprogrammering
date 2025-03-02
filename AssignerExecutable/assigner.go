package main

import (
	"TTK4145-Heislab/communication"
	"TTK4145-Heislab/single_elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

// legger til alt fra assigner filen inn i en funksjon slik at den kan kontinuerlig kalles på
func Assigner(commonstate communication.CommonState, id int) single_elevator.OrderMatrix { //skal det stå single elevator her eller ikke?

	stateMap := make(map[string]HRAElevState)
	for i, v := range commonstate.State {
		//sjekke om heisen er tilgjengelig - hvis den ikke er det continue
		//else:
		stateMap[strconv.Itoa(i)] = HRAElevState{
			Behaviour:   v.State.Behaviour.ToString(),
			Floor:       v.State.Floor,
			Direction:   v.State.Direction.ToString(),
			CabRequests: v.CabRequests,
		}
	}

	hraInput := HRAInput{commonstate.HallRequests, stateMap}

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	jsonBytes, err := json.Marshal(hraInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		//return
	}

	ret, err := exec.Command("executables/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		//return
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		//return
	}

	//fmt.Printf("output: \n")
	//for k, v := range *output {
	//fmt.Printf("%6v :  %+v\n", k, v)
	//}

	//må returnere ID siden vi skal bestemme hvilken heis som skal ta orderen
	return (*output)[strconv.Itoa(id)]

	//convert output to matrix sånn at dette kan tas rett inn i order manager
}

/*OUTPUT FRA HALL ASSIGNER - må sendes til order manager. Order manager må legge sammen egen matrise med matrise fra hall assigner. 1+1 skal ikke bli 2.
må sende hver av linjene til riktig heis. i order manager konverterer vi fra string til ordermatrix
{
    "0": [[false, true, false], [true, false, false], [false, false, true], [false, false, false]],
    "1": [[false, false, false], [false, false, true], [true, false, false], [false, true, false]],
    "2": [[true, false, false], [false, true, false], [false, false, true], [false, false, false]]
}*/

//main func removed

//verdensbilde eks
//valid state. broadcaste
//alle har samme verdensbilde, alle kjører samme algoritmen
//knappetrykk som order
//UDP broadcast example
//NEED TO GENERALIZE
/*input := HRAInput{
	HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
	States: map[string]HRAElevState{
		"one": HRAElevState{
			Behavior:    "moving",
			Floor:       2,
			Direction:   "up",
			CabRequests: []bool{false, false, false, true},
		},
		"two": HRAElevState{
			Behavior:    "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: []bool{false, false, false, false},
		},
	},
}*/
