package requestassigner
/* map over data that is being syncronized
-----------------------------------
Elevator States:
[ 	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[4]],
	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[4]],
	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[4]]	]

Hall Requests:
Hall_Request_Data[N_FLOORS][2]

Every piece of data have a list with the elevators that also agree with the iformation.
If this list == elevator_network_list then we send this data to the HSA
-----------------------------------
*/

//import something to use System_Data_t

import (
	"TTK4145-26V/elevator"
	"TTK4145-26V/message_sync"
	"strconv"
)

func Generating_RA_System_Data(confirmed_system_data message_sync.System_Data_t) RA_System_Data{
	ra_system := RA_System_Data{}
	ra_system.HallRequests = make([][N_HALL_CALLS]bool,elevator.N_FLOORS)
	for floor:=0; floor < elevator.N_FLOORS;floor++{
		for button:= 0; button < N_HALL_CALLS; button++{
			ra_system.HallRequests[floor][button] = CC_To_Bool(confirmed_system_data.Hall_Request_Data[floor][button].Value)
		}
	}

	ra_system.States = make(map[string]RA_Local_Elevator_State)
	for _,elev := range confirmed_system_data.Elevator_Data{

		if !(elev.Is_Alive) || !(elev.Is_Able){
			continue
		}

		cab_bools := make([]bool,elevator.N_FLOORS)
		for floor:= 0; floor< elevator.N_FLOORS; floor++{
			cab_bools[floor] = CC_To_Bool(elev.Cab_Requests[floor].Value)
		}

		Id_str := strconv.Itoa(elev.Id)

		ra_system.States[Id_str] = RA_Local_Elevator_State{
			Behavior:    elevator.Elevator_behaviour_to_string(elev.Elevator_Behaviour),
			Floor:       elev.Floor,
			Direction:   elevator.Elevator_dirn_to_string(elev.Motor_Direction),
			CabRequests: cab_bools,
		}
	}
	return ra_system
}