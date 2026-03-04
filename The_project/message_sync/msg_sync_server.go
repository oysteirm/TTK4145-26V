package message_sync

import (
	"TTK4145-26V/elevator"
)
/* map over data that is being syncronized
-----------------------------------
Elevator States:
[ 	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[N_FLOORS]],
	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[N_FLOORS]], 
	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[N_FLOORS]]	]

Hall Requests:
Hall_Request_Data[N_FLOORS][N_HALL_CALLS]

Every piece of data have a list with the elevators who agree with the information. 
If this list == elevator_network_list then we send this data have reached consensus and is put in confirmed data which is sent to HSA
-----------------------------------
*/

const (
	CC_Uninit Cyclic_Counter_t 	= -1
	CC_No 						= 0
	CC_Unconfirmed 				= 1
	CC_Confirmed 				= 2
	CC_Done 					= 3
)

var N_ELEVATORS int = 3


type Elev_List_t []bool
type Cyclic_Counter_t int8

//Data type structs that include the data and a barrier
type Request_Cyclic_Counter_t struct{
	Value Cyclic_Counter_t
	barrier Elev_List_t
}
// type Is_Alive_Data_t struct{
// 	Value bool
// 	barrier Elev_List_t
// }
// type Is_Able_Data_t struct{
// 	Value bool
// 	barrier Elev_List_t
// }
// type Floor_Data_t struct{
// 	Value int
// 	barrier Elev_List_t
// }
// type Elevator_Behaviour_Data_t struct{
// 	Value elevator.ElevatorBehaviour_t
// 	barrier Elev_List_t
// }
// type Motor_Direction_Data_t struct{
// 	Value elevator.MotorDirection_t
// 	barrier Elev_List_t
// }

type Elevator_Data_t struct {
	Id int
	Is_Alive bool
	Is_Able bool
	Floor int
	Elevator_Behaviour elevator.ElevatorBehaviour_t
	Motor_Direction elevator.MotorDirection_t
	Cab_Requests []Request_Cyclic_Counter_t
}

type System_Data_t struct {
	Id int
	Elevator_Data []Elevator_Data_t
	Hall_Request_Data [][2]Request_Cyclic_Counter_t
}

type Get_System_Data_t struct{
	Reply System_Data_t
}


