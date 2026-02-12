package message_sync

import (
	"../elevator"
	"fmt"
	"time"
)
/* map over data that is being syncronized
-----------------------------------
[ID		ALIVE 	IS_ABLE	]struct		
[FLOOR	EB		MD		]
[H_U	H_D		CAB		]4
[H_U	H_D		CAB		]3
[H_U	H_D		CAB		]2
[H_U	H_D		CAB		]1

[ID		ALIVE 	IS_ABLE	]struct		
[FLOOR	EB		MD		]
[H_U	H_D		CAB		]4
[H_U	H_D		CAB		]3
[H_U	H_D		CAB		]2
[H_U	H_D		CAB		]1

[ID		ALIVE 	IS_ABLE	]struct		
[FLOOR	EB		MD		]
[H_U	H_D		CAB		]4
[H_U	H_D		CAB		]3
[H_U	H_D		CAB		]2
[H_U	H_D		CAB		]1
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

type Cyclic_Counter_t int
type Elev_Alive_List_t []int

type Request_Cyclic_Counter_t struct{
	Cyclic_Counter Cyclic_Counter_t
	barrier Elev_Alive_List_t
}

type Requests_Data_t [][]Request_Cyclic_Counter_t

type Elevator_Data_t struct {
	Id int
	Is_Alive bool
	Is_Able bool
	Floor int
	Elevator_Behaviour elevator.Elevator_Behaviour_t
	Motor_Direction elevator.Motor_Direction_t
}

type System_Data_t struct {
	Elevator_Data []Elevator_Data_t
	Requests_Data []Requests_Data_t
}

func Message_Sync_Server(){}

func Update_CC(){}
func Check_Barrier(){}

