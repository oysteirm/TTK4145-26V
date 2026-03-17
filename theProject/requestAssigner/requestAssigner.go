package requestAssigner

// This code is inspired by provided code, fetched from https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner

import (
	"theProject/config"
	"encoding/json"
	"fmt"
	"os/exec"
)

/*
-----------------------------------
Functionallity: 
	- Converts internal system data to JSON format expected by the assigner
	- Executes the provided request assigner 
	- Parses and returns assigned requests back to the system
Design: 
	- The requestAssigner module does not implement assignment logic itself
	- It acts as a bridge between our Go code and the provided assigner
	- All communication with the assigner happens through JSON encoding/decoding

-----------------------------------
*/


// Local elevator state formatted for the request assigner
// Note that field names and values must match expected JSON format 
type RA_LocalElevatorState_t struct {
    Behavior    string      `json:"behaviour"`// "moving", "doorOpen", "idle" (all lowercase)
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"` // "stop", "up" or "down" (all lowercase)
    CabRequests []bool      `json:"cabRequests"`
}

//System data formatted for the request assigner
type RA_SystemData_t  struct {
    HallRequests    [][config.N_UP_DOWN]bool               `json:"hallRequests"`
    States          map[string]RA_LocalElevatorState_t     `json:"states"`
}


type RA_Output_t map[string][][]bool

//Sends current system state to external request assigner and returns assigned requests
func AssignRequests(elevatorSystem RA_SystemData_t) RA_Output_t {

	//Encoding system data to JSON string
	input, err := json.Marshal(elevatorSystem)
	if err != nil {
		fmt.Println("Marshal error in AssignRequests:", err)
		return nil
	}

	//Executing "providedRequestAssigner" , fetched from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3
	output, err := exec.Command(
		"./requestAssigner/providedRequestAssigner",
		"--includeCab",
		"-i", string(input),
	).CombinedOutput()

	// Print to debug info if execution fails
	if err != nil {
		fmt.Println("Exec error in AssignRequests:", err)
		fmt.Println("AssignRequests output:", string(output))
		fmt.Println("RA input JSON: ")
		fmt.Println(string(input))
		return nil
	}

	//Decoding JSON output into result struct
	var result RA_Output_t
	err = json.Unmarshal(output, &result)
	if err != nil {
		fmt.Println("Unmarshal error in AssignRequests:", err)
		return nil
	}

	return result
}


