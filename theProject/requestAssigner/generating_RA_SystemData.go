package requestAssigner

/* map over data that is being syncronized
-----------------------------------
Elevator States:
[ 	[ID		IsAlive 	IsFunctional	FLOOR	EB		MD	CabRequests[4]],
	[ID		IsAlive 	IsFunctional	FLOOR	EB		MD	CabRequests[4]],
	[ID		IsAlive 	IsFunctional	FLOOR	EB		MD	CabRRequests[4]]	]

Hall Requests:
HallRequestData[N_FLOORS][2]

Every piece of data have a list with the elevators that also agree with the iformation.
If this list == elevatorNetworkList then we send this data to the HSA
-----------------------------------
*/

import (
	"strconv"
	"theProject/config"
	"theProject/converters"
	"theProject/messageSync"
)

func Generating_RA_SystemData(confirmedSystemData messageSync.SystemData_t) RA_SystemData_t {
	RA_systemData := RA_SystemData_t{}
	RA_systemData.HallRequests = make([][config.N_UP_DOWN]bool, config.N_FLOORS)
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_UP_DOWN; button++ {
			RA_systemData.HallRequests[floor][button] = converters.CC_ToBool(confirmedSystemData.HallRequestData[floor][button].Value)
		}
	}

	RA_systemData.States = make(map[string]RA_LocalElevatorState_t)
	for _, elevator := range confirmedSystemData.ElevatorData {

		if !(elevator.IsAlive) || !(elevator.IsFunctional) || (elevator.Floor == -1) {
			continue
		}

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

	if len(RA_systemData.States) == 0 {
		for _, elevator := range confirmedSystemData.ElevatorData {
			if !(elevator.IsAlive) || (elevator.Floor == -1) {
				continue
			}

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
