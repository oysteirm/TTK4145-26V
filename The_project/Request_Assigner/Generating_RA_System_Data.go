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
	for i:=0; i < elevator.N_FLOORS;i++{
		for j:= 0; j < N_HALL_CALLS; j++{
			ra_system.HallRequests[i][j] = CC_To_Bool(confirmed_system_data.Hall_Request_Data[i][j].Value)
		}
	}

	ra_system.States = make(map[string]RA_Local_Elevator_State)
	for _,elev := range confirmed_system_data.Elevator_Data{

		if !(elev.Is_Alive.Value) || !(elev.Is_Able.Value){
			continue
		}

		cab_bools := make([]bool,elevator.N_FLOORS)
		for i:= 0; i< elevator.N_FLOORS; i++{
			cab_bools[i] = CC_To_Bool(elev.Cab_Requests[i].Value)
		}

		Id_str := strconv.Itoa(elev.Id)

		ra_system.States[Id_str] = RA_Local_Elevator_State{
			Behavior:    elevator.Elevator_behaviour_to_string(elev.Elevator_Behaviour.Value),
			Floor:       elev.Floor.Value,
			Direction:   elevator.Elevator_dirn_to_string(elev.Motor_Direction.Value),
			CabRequests: cab_bools,
		}
	}
	return ra_system
}