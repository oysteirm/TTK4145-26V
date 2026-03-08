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
	
	"theProject/config"
	"theProject/elevatorStateGuardian"
	"theProject/messageSync"
	"strconv"
)

func Generating_RA_SystemData(confirmedSystemData messageSync.SystemData_t) RA_SystemData_t{
	RA_systemData := RA_SystemData_t{}
	RA_systemData.HallRequests = make([][config.N_UP_DOWN]bool, config.N_FLOORS)
	for floor:=0; floor < config.N_FLOORS;floor++{
		for button:= 0; button < config.N_UP_DOWN; button++{
			RA_systemData.HallRequests[floor][button] = CC_ToBool(confirmedSystemData.HallRequestData[floor][button].Value)
		}
	}

	RA_systemData.States = make(map[string]RA_LocalElevatorState_t)
	for _,elev := range confirmedSystemData.ElevatorData{

		if !(elev.IsAlive) || !(elev.IsFunctional){
			continue
		}

		cabBools := make([]bool, config.N_FLOORS)
		for floor:= 0; floor< config.N_FLOORS; floor++{
			cabBools[floor] = CC_ToBool(elev.CabRequests[floor].Value)
		}

		IdStr := strconv.Itoa(elev.ID)

		RA_systemData.States[IdStr] = RA_LocalElevatorState_t{
			Behavior:    elevatorStateGuardian.ElevatorBehaviourToString(elev.ElevatorBehaviour),
			Floor:       elev.Floor,
			Direction:   elevatorStateGuardian.ElevatorDirnToString(elev.MotorDirection),
			CabRequests: cabBools,
		}
	}
	return RA_systemData
}