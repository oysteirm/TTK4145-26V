package requestassigner

//Example terminal_input to test compiled hall_request_assigner
//./tools/hall_request_assigner/hall_request_assigner -i '{"hallRequests":[[false,false],[true,false],[false,false],[false,true]],"states":{"one":{"behaviour":"moving","floor":2,"direction":"up","cabRequests":[false,false,false,true]},"two":{"behaviour":"idle","floor":0,"direction":"stop","cabRequests":[false,false,false,false]}}}'


import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type RA_Local_Elevator_State struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}
//REMEMBER TO MOVE N_CAB_CALLS, ex into elevator_io
const N_HALL_CALLS = 2

type RA_Elevator_States_and_Requests  struct {
    HallRequests    [][N_HALL_CALLS]bool                   `json:"hallRequests"`
    States          map[string]RA_Local_Elevator_State     `json:"states"`
}

//GIVE BETTER NAMES?
type RA_Output map[string][][]bool



func Assign_Orders(Elevator_System RA_Elevator_States_and_Requests) RA_Output {

	//ENCODING SYSTEM
	input, err := json.Marshal(Elevator_System)
	if err != nil {
		fmt.Println("Marshal error in Assign_Orders:", err)
		return nil
	}

	//EXECUTING COMPILED "hall_request_assigner", fetched from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3
	output, err := exec.Command(
		"./tools/hall_request_assigner/hall_request_assigner",
		"--includeCab",
		"-i", string(input),
	).CombinedOutput()
	if err != nil {
		fmt.Println("Exec error in Assign_Orders:", err)
		return nil
	}

	//DECODING STRING
	var result RA_Output
	err = json.Unmarshal(output, &result)
	if err != nil {
		fmt.Println("Unmarshal error in Assign_Orders:", err)
		return nil
	}

	return result
}