package elevatorStateGuardian

import (
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/messageSync"
)

/*
-----------------------------------
Functionality: 
	- Owns the local systemState used to control the local elevator FSM
	- Communicates new elevator data and done requests to messageSync
	- guardianCommands is used by elevatorSever for message passing to get and set systemData Values.
	  It is used by sending the correct struct type to the channel. 
-----------------------------------
*/

// Channel type for communicating with the elevatorStateGuardian
type GuardianCommands_t interface{}

// Get types
/*
Example of usage: 
elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
*/
type GetElevatorData_t struct {
	Reply chan messageSync.ElevatorData_t
}
type GetAssignedRequests_t struct {
	Reply chan elevator_IO.AssignedRequests_t
}

// Set types
/*
Example of usage: 
guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
*/
type SetSystemData_t struct {
	SystemData messageSync.SystemData_t
}
type SetElevatorData_t struct {
	ElevatorData messageSync.ElevatorData_t
}
type SetIsFunctional_t struct {
	IsFunctional bool
}
type SetRequestsDone_t struct {
	RequestsToClear []elevator_IO.ButtonEvent_t
}
type SetAssignedRequest_t struct {
	AssignedRequests elevator_IO.AssignedRequests_t
}


func ElevatorStateGuardian(
	guardianCommands chan GuardianCommands_t, 					// Channel for get / set systemData
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t, 	// Channel for sending data to messageSyncServer
	requestsToMsgSync chan<- []elevator_IO.ButtonEvent_t, 		// Channel for sending done request CC to messageSyncServer
	localID int) { 												// ID of the local elevator

	// Initialize the system data
	var systemData messageSync.SystemData_t
	systemData, _ = messageSync.InitSystemData(localID)
	var elevatorDataChanged bool = false

	requests_temp := make([][]bool, config.N_FLOORS)
	for floor := 0;  floor < config.N_FLOORS; floor++ {
		requests_temp[floor] = make([]bool, config.N_BUTTONS)
	}
	var assignedRequests elevator_IO.AssignedRequests_t = requests_temp

	// Loop that continuesly decodes guardianCommands type and executes it. 
	for cmd := range guardianCommands {
		switch c := cmd.(type) {

		case GetElevatorData_t:
			c.Reply <- systemData.ElevatorData[localID]

		case GetAssignedRequests_t:
			c.Reply <- assignedRequests

		// Used by msg sync to set the new confirmed system data
		// But Not letting potentially old data overwrite the local state
		case SetSystemData_t:
			tmpFloor := systemData.ElevatorData[localID].Floor
			tmpElevatorBehaviour := systemData.ElevatorData[localID].ElevatorBehaviour
			tmpMotorDirection := systemData.ElevatorData[localID].MotorDirection
			tmpIsFunctional := systemData.ElevatorData[localID].IsFunctional

			systemData = c.SystemData

			systemData.ElevatorData[localID].Floor = tmpFloor
			systemData.ElevatorData[localID].MotorDirection = tmpMotorDirection
			systemData.ElevatorData[localID].ElevatorBehaviour = tmpElevatorBehaviour
			systemData.ElevatorData[localID].IsFunctional = tmpIsFunctional

		// Used by the FSM to set the local elevator state
		case SetElevatorData_t:
			old := systemData.ElevatorData[localID]

			changed  := 
					old.IsFunctional 		!= c.ElevatorData.IsFunctional 		||
        			old.Floor 				!= c.ElevatorData.Floor 			||
        			old.ElevatorBehaviour 	!= c.ElevatorData.ElevatorBehaviour ||
        			old.MotorDirection 		!= c.ElevatorData.MotorDirection
	
			if changed {
			systemData.ElevatorData[localID].IsFunctional 		= c.ElevatorData.IsFunctional
			systemData.ElevatorData[localID].Floor 				= c.ElevatorData.Floor
			systemData.ElevatorData[localID].ElevatorBehaviour 	= c.ElevatorData.ElevatorBehaviour
			systemData.ElevatorData[localID].MotorDirection 	= c.ElevatorData.MotorDirection

			systemData.ElevatorData[localID].ElevatorBarrier 		  	= [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID] 	= true
			elevatorDataChanged = true
			}

		case SetIsFunctional_t:
			systemData.ElevatorData[localID].IsFunctional 				= c.IsFunctional
			systemData.ElevatorData[localID].ElevatorBarrier 			= [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID] 	= true
			elevatorDataChanged = true


		// Sending requests that the local FSM have executed to messageSync
		// We filter out requests that is not on the same floor
		case SetRequestsDone_t:
			currentFloor := systemData.ElevatorData[localID].Floor
			filteredRequests := make([]elevator_IO.ButtonEvent_t, 0, len(c.RequestsToClear))

			for _, req := range c.RequestsToClear {
				if req.Floor == currentFloor {
					filteredRequests = append(filteredRequests, req)
				}
			}
			if len(filteredRequests) > 0 {
				requestsToMsgSync <- filteredRequests
			}

		// The assigned reqests from RA
		case SetAssignedRequest_t:
			assignedRequests = c.AssignedRequests
		}

		// If data in the elevator state was changed by FSM, then we send it to messageSync
		if elevatorDataChanged {
			elevatorDataToMsgSync <- systemData.ElevatorData[localID]
			elevatorDataChanged = false
		}
	}
}


// Functions for using the get functionallity with guardianCommands
func GetElevatorData(guardianCommands chan GuardianCommands_t) messageSync.ElevatorData_t {
	reply := make(chan messageSync.ElevatorData_t)
	guardianCommands <- GetElevatorData_t{Reply: reply}
	return <-reply
}
func GetAssignedRequests(guardianCommands chan GuardianCommands_t) elevator_IO.AssignedRequests_t {
	reply := make(chan elevator_IO.AssignedRequests_t)
	guardianCommands <- GetAssignedRequests_t{Reply: reply}
	return <-reply
}