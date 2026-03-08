package requestAssigner

//Example terminal_input to test compiled hallRequestAssigner:
//./tools/hallRequestAssigner-i '{"hallRequests":[[false,false],[true,false],[false,false],[false,true]],"states":{"one":{"behaviour":"moving","floor":2,"direction":"up","cabRequests":[false,false,false,true]},"two":{"behaviour":"idle","floor":0,"direction":"stop","cabRequests":[false,false,false,false]}}}'





import (
	"theProject/config"
	"encoding/json"
	"fmt"
	"os/exec"
)

type RA_LocalElevatorState_t struct {
    Behavior    string      `json:"behaviour"`// "moving", "doorOpen", "idle" (all lowercase)
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"` // "stop", "up" or "down" (all lowercase)
    CabRequests []bool      `json:"cabRequests"`
}


type RA_SystemData_t  struct {
    HallRequests    [][config.N_UP_DOWN]bool               `json:"hallRequests"`
    States          map[string]RA_LocalElevatorState_t     `json:"states"`
}

//GIVE BETTER NAMES?
type RA_Output_t map[string][][]bool



func AssignRequests(elevatorSystem RA_SystemData_t) RA_Output_t {

	//ENCODING SYSTEM
	input, err := json.Marshal(elevatorSystem)
	if err != nil {
		fmt.Println("Marshal error in AssignRequests:", err)
		return nil
	}

	//EXECUTING COMPILED "providedRequestAssigner" , fetched from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3
	output, err := exec.Command(
		"./requestAssigner/providedRequestAssigner",
		"--includeCab",
		"-i", string(input),
	).CombinedOutput()
	if err != nil {
		fmt.Println("Exec error in AssignRequests:", err)
		fmt.Println("AssignRequests output:", string(output))
		return nil
	}

	//DECODING STRING
	var result RA_Output_t
	err = json.Unmarshal(output, &result)
	if err != nil {
		fmt.Println("Unmarshal error in AssignRequests:", err)
		return nil
	}

	return result
}


