package message_sync

import (
	"time"
	"the_project/elevator"
	"the_project/Network_driver/peers"
	"the_project/Network_Driver/bcast"
)	
/* map over data that is being syncronized
-----------------------------------
Elevator States:
[ 	[ID		ALIVE 	IS_Functional		FLOOR	EB		MD	Cab_Requests[N_FLOORS]],
	[ID		ALIVE 	IS_Functional	    FLOOR	EB		MD	Cab_Requests[N_FLOORS]], 
	[ID		ALIVE 	IS_Functional		FLOOR	EB		MD	Cab_Requests[N_FLOORS]]	]

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

const N_ELEVATORS int = 3
const btns_UP_and_Down int = 2

type Elev_List_t []bool
type Cyclic_Counter_t int

//Data type structs that include the data and a barrier
type Request_Cyclic_Counter_t struct{
	Value Cyclic_Counter_t
	Barrier Elev_List_t
}
type Is_Alive_Data_t struct{
	Value bool
	Barrier Elev_List_t
}
type Is_Functional_Data_t struct{
	Value bool
	Barrier Elev_List_t
}
type Floor_Data_t struct{
	Value int
	Barrier Elev_List_t
}
type Elevator_Behaviour_Data_t struct{
	Value elevator.Elevator_Behaviour_t
	Barrier Elev_List_t
}
type Motor_Direction_Data_t struct{
	Value elevator.Motor_Direction_t
	Barrier Elev_List_t
}
//Datatype for elevator states with barriers
type Elevator_Data_t struct {
	Id int
	//Msg_counter uint64
	Is_Alive Is_Alive_Data_t
	Is_Functional Is_Functional_Data_t
	Floor Floor_Data_t
	Elevator_Behaviour Elevator_Behaviour_Data_t
	Motor_Direction Motor_Direction_Data_t
	Cab_Requests []Request_Cyclic_Counter_t
}
//Datatype for multi elevator states and hall requests
type System_Data_t struct {
	Id int
	Elevator_Data []Elevator_Data_t
	Hall_Request_Data [][2]Request_Cyclic_Counter_t
}

type Get_System_Data_t struct{
	Reply System_Data_t
}

func Message_Sync_Server(
	from_network_data <-chan System_Data_t, //channel for recieving new system data
	get_system_data <-chan Get_System_Data_t, //channel for other routines to get the current system data
	from_fsm_data <-chan Elevator_Data_t, //channel for recieving elevator data from fsm
	peersReciever <-chan peers.PeerUpdate,
	local_id int,
	){
	// 
	var system_data System_Data_t
	var confirmed_system_data System_Data_t
	system_data, confirmed_system_data = Init_System_Data(local_id)
	var is_confirmed_data_updated bool = false

	// Network variables
	var activePeers []string
	networkReciever := make(chan System_Data_t)
	networkTransmitter := make(chan System_Data_t)
	bcastPort := 1234 //TODO: change this to a correct value

	// Go routines from Network_Driver
	go bcast.Receiver(bcastPort, networkReciever)
	go bcast.Transmitter(bcastPort, networkTransmitter)
	
	// Timer for broadcasting
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Go routine for button polling
	drv_buttons := make(chan elevator.ButtonEvent_t)
	go elevator.PollButtons(drv_buttons)
	

	for {
		select{
		case reg := <- get_system_data:
			reg.Reply = system_data

		case fresh_data := <- from_network_data:
			system_data, confirmed_system_data, is_confirmed_data_updated = On_Recieved_Fresh_Data(system_data, confirmed_system_data, fresh_data)

			if is_confirmed_data_updated{
				//TODO: send confirmed_data til elev_FSM
			}
			//use confirmed_data for light contract 
			//TODO: write these functions and place them in elev_server
			Light_Cab_Lights(confirmed_system_data.Elevator_Data[local_id].Cab_Requests)
			Light_Hall_Lights(confirmed_system_data.Hall_Request_Data)

		case fresh_data := <- from_fsm_data:
			system_data.Elevator_Data[local_id] = Update_Elevator_Data_About_Self(system_data.Elevator_Data[local_id], fresh_data, local_id)
			
		//buttonpress tries to change the CC to unconfirmed
		case btn := <-drv_buttons:
			if btn.Button == elevator.BT_Cab {
				var tmp_cab_request Request_Cyclic_Counter_t = Request_Cyclic_Counter_t{Value: CC_Unconfirmed, Barrier: make(Elev_List_t, N_ELEVATORS)} //blind copy

			}
		case //broadcast timer timeout
			system_data.Elevator_Data[local_id].Msg_counter ++


		case peersUpdate := <-peersReciever:
			//TODO: format peersupdate to a bool list 
			activePeers = peersUpdate.Peers

		}
	}
}


