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

//import something to use N_FLOORS, ex from elevator_io.go

import (
	"strconv"
)

func Generating_RA_System_Data(confirmed_system_data System_Data_t) RA_System_Data{
	ra_system := RA_System_Data{}
	ra_system.HallRequests = make([][N_HALL_CALLS]bool,N_FLOORS)
	for i:=0; i < N_FLOORS;i++{
		for j:= 0; j < N_HALL_CALLS; j++{
			ra_system.HallRequests[i][j] = Counter_To_Bool(confirmed_system_data.Hall_Request_Data[i][j])
		}
	}

	ra_system.States = make(map[string]RA_Local_Elevator_State)
	for _,elevator := range confirmed_system_data.Elevator_Data{

		if !(elevator.Is_Alive.value) || !(elevator.Is_Able.value){
			continue
		}

		cab_bools := make([]bool,N_FLOORS)
		for i:= 0; i<N_FLOORS; i++{
			cab_bools[i] = Counter_To_Bool(elevator.Cab_Requests[i])
		}

		Id_str := strconv.Itoa(elevator.Id)

		ra_system.States[Id_str] = RA_Local_Elevator_State{
			Behavior:    elevator_behaviour_to_string(elevator.Elevator_Behaviour.value),
			Floor:       elevator.Floor.value,
			Direction:   elevator_dirn_to_string(elevator.Motor_Direction.value),
			CabRequests: cab_bools,
		}
	}
	return ra_system
}