package requestAssigner

//Example terminal_input to test compiled hallRequestAssigner:
//./tools/hallRequestAssigner-i '{"hallRequests":[[false,false],[true,false],[false,false],[false,true]],"states":{"one":{"behaviour":"moving","floor":2,"direction":"up","cabRequests":[false,false,false,true]},"two":{"behaviour":"idle","floor":0,"direction":"stop","cabRequests":[false,false,false,false]}}}'





import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type RA_LocalElevatorState struct {
    Behavior    string      `json:"behaviour"`// "moving", "doorOpen", "idle"
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}
//REMEMBER TO MOVE N_CAB_CALLS, ex into elevator_io
const N_HALL_CALLS = 2

type RA_SystemData  struct {
    HallRequests    [][N_HALL_CALLS]bool                   `json:"hallRequests"`
    States          map[string]RA_LocalElevatorState     `json:"states"`
}

//GIVE BETTER NAMES?
type RA_Output map[string][][]bool



func AssignRequests(elevatorSystem RA_SystemData) RA_Output {

	//ENCODING SYSTEM
	input, err := json.Marshal(elevatorSystem)
	if err != nil {
		fmt.Println("Marshal error in AssignRequests:", err)
		return nil
	}

	//EXECUTING COMPILED "hallRequestAssigner", fetched from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3
	output, err := exec.Command(
		"./tools/hallRequestAssignerDir/hallRequestAssigner",
		"--includeCab",
		"-i", string(input),
	).CombinedOutput()
	if err != nil {
		fmt.Println("Exec error in AssignRequests:", err)
		fmt.Println("AssignRequests output:", string(output))
		return nil
	}

	//DECODING STRING
	var result RA_Output
	err = json.Unmarshal(output, &result)
	if err != nil {
		fmt.Println("Unmarshal error in AssignRequests:", err)
		return nil
	}

	return result
}


