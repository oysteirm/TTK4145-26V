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
	"theProject/elevator"
	"theProject/messageSync"
	"strconv"
)

func Generating_RA_SystemData(confirmedSystemData messageSync.SystemData_t) RA_SystemData{
	raSystem := RA_SystemData{}
	raSystem.HallRequests = make([][elevator.N_UP_DOWN]bool,elevator.N_FLOORS)
	for floor:=0; floor < elevator.N_FLOORS;floor++{
		for button:= 0; button < elevator.N_UP_DOWN; button++{
			raSystem.HallRequests[floor][button] = CC_ToBool(confirmedSystemData.HallRequestData[floor][button].Value)
		}
	}

	raSystem.States = make(map[string]RA_LocalElevatorState)
	for _,elev := range confirmedSystemData.ElevatorData{

		if !(elev.IsAlive) || !(elev.IsFunctional){
			continue
		}

		cabBools := make([]bool,elevator.N_FLOORS)
		for floor:= 0; floor< elevator.N_FLOORS; floor++{
			cabBools[floor] = CC_ToBool(elev.CabRequests[floor].Value)
		}

		IdStr := strconv.Itoa(elev.ID)

		raSystem.States[IdStr] = RA_LocalElevatorState{
			Behavior:    elevator.ElevatorBehaviourToString(elev.ElevatorBehaviour),
			Floor:       elev.Floor,
			Direction:   elevator.ElevatorDirnToString(elev.MotorDirection),
			CabRequests: cabBools,
		}
	}
	return raSystem
}