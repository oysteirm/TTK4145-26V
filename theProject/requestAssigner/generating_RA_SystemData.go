package requestAssigner

import (
	"strconv"
	"theProject/config"
	"theProject/converters"
	"theProject/messageSync"
)

/*
-----------------------------------
Functionality:
	- Converts confirmed system data into the format required by the request assigner
	- Filters out invalid or unusable elevators before sending data to the assigner
	- Acts as a translation layer between internal data structures and RA input format
	- Only valid elevators are included in assignment to avoid bad decisions
	- Falls back to including all alive elevators if no fully valid ones are found
-----------------------------------
*/

func Generating_RA_SystemData(confirmedSystemData messageSync.SystemData_t) RA_SystemData_t {

	RA_systemData := RA_SystemData_t{}

	// Convert hall requests to boolean format expected by the request assigner
	RA_systemData.HallRequests = make([][config.N_UP_DOWN]bool, config.N_FLOORS)
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_UP_DOWN; button++ {
			RA_systemData.HallRequests[floor][button] = converters.CC_ToBool(confirmedSystemData.HallRequestData[floor][button].Value)
		}
	}

	RA_systemData.States = make(map[string]RA_LocalElevatorState_t)
	
	// Adding valid elevators (alive, functional, and with known floor)
	for _, elevator := range confirmedSystemData.ElevatorData {

		if !(elevator.IsAlive) || !(elevator.IsFunctional) || (elevator.Floor == -1) {
			continue
		}

		// Convert cab requests to boolean slice
		cabBools := make([]bool, config.N_FLOORS)
		for floor := 0; floor < config.N_FLOORS; floor++ {
			cabBools[floor] = converters.CC_ToBool(elevator.CabRequests[floor].Value)
		}

		IdStr := strconv.Itoa(elevator.ID)

		RA_systemData.States[IdStr] = RA_LocalElevatorState_t{
			Behavior:    converters.ElevatorBehaviourToString(elevator.ElevatorBehaviour),
			Floor:       elevator.Floor,
			Direction:   converters.ElevatorDirnToString(elevator.MotorDirection),
			CabRequests: cabBools,
		}
	}
	// If no fully valid elevators were found, include all alive elevators to avoid empty input to request assigner
	if len(RA_systemData.States) == 0 {
		for _, elevator := range confirmedSystemData.ElevatorData {
			if !(elevator.IsAlive) || (elevator.Floor == -1) {
				continue
			}

			// Convert cab requests to boolean slice
			cabBools := make([]bool, config.N_FLOORS)
			for floor := 0; floor < config.N_FLOORS; floor++ {
				cabBools[floor] = converters.CC_ToBool(elevator.CabRequests[floor].Value)
			}

			IdStr := strconv.Itoa(elevator.ID)
			
			RA_systemData.States[IdStr] = RA_LocalElevatorState_t{
				Behavior:    converters.ElevatorBehaviourToString(elevator.ElevatorBehaviour),
				Floor:       elevator.Floor,
				Direction:   converters.ElevatorDirnToString(elevator.MotorDirection),
				CabRequests: cabBools,
			}
		}
	}

	return RA_systemData
}