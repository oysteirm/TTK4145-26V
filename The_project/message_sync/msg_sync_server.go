package message_sync

import (
	"time"
	"the_project/elevator"
	"the_project/Network_Driver/peers"
	"the_project/Network_Driver/bcast"
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
	CC_Uninit 		Cyclic_Counter_t 	= -1
	CC_No 			Cyclic_Counter_t	= 0
	CC_Unconfirmed 	Cyclic_Counter_t	= 1
	CC_Confirmed 	Cyclic_Counter_t	= 2
	CC_Done 		Cyclic_Counter_t	= 3
)

const N_ELEVATORS 		int = 3
const btns_UP_and_Down 	int = 2

// List containing info about our network peers
// 1: part of network
// 0: not part of network
var Active_Peers []bool

type Cyclic_Counter_t int

//Data type structs that include the data and a Barrier
type Request_Cyclic_Counter_t struct{
	Value Cyclic_Counter_t
	Barrier []bool
}
type Is_Alive_Data_t struct{
	Value bool
	Barrier []bool
}
type Is_Functional_Data_t struct{
	Value bool
	Barrier []bool
}
type Floor_Data_t struct{
	Value int
	Barrier []bool
}
type Elevator_Behaviour_Data_t struct{
	Value elevator.Elevator_Behaviour_t
	Barrier []bool
}
type Motor_Direction_Data_t struct{
	Value elevator.Motor_Direction_t
	Barrier []bool
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
	get_system_data <-chan Get_System_Data_t, 	//channel for other routines to get the current system data
	data_from_fsm <-chan Elevator_Data_t, 		//channel for recieving elevator data from elevator FSM
	data_to_fsm chan<- System_Data_t, 			//channel for sending confirmed data to FSM
	peersReciever <-chan peers.PeerUpdate,
	local_id int,
	){

	// Variables used to sync data
	var system_data System_Data_t
	var confirmed_system_data System_Data_t
	system_data, confirmed_system_data = Init_System_Data(local_id)
	var is_confirmed_data_updated bool = false

	// Network variables
	network_receiver := make(chan System_Data_t)
	network_transmitter := make(chan System_Data_t)
	bcastPort := 1234 //TODO: change this to a correct value

	// Go routines from Network_Driver
	go bcast.Receiver(bcastPort, network_receiver)
	go bcast.Transmitter(bcastPort, network_transmitter)
	
	// Timer for broadcasting
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Go routine for button polling
	drv_buttons := make(chan elevator.ButtonEvent_t)
	go elevator.PollButtons(drv_buttons)
	
	for {
		select{
		//Someone needs the system data
		case reg := <- get_system_data:
			reg.Reply = system_data

		//We recieve new data from the network
		case fresh_data := <- network_receiver:
			system_data, confirmed_system_data, is_confirmed_data_updated = On_Received_Fresh_Data(system_data, confirmed_system_data, fresh_data)

			//If we have new confirmed data, we sent it to the elevator FSM
			if is_confirmed_data_updated{
				data_to_fsm <- confirmed_system_data
			}

			//TODO: place them in elev_server
			Light_Cab_Lights(confirmed_system_data.Elevator_Data[local_id].Cab_Requests)
			Light_Hall_Lights(confirmed_system_data.Hall_Request_Data)

		//We recieve data from the elevator FSM
		case fresh_data := <- data_from_fsm:
			system_data.Elevator_Data[local_id] = Update_Elevator_Data_About_Self(system_data.Elevator_Data[local_id], fresh_data, local_id)
			
		//new buttonpress tries to change the CC to unconfirmed
		case btn := <-drv_buttons:
			if btn.Button == elevator.BT_Cab {
				var tmp_cab_request Request_Cyclic_Counter_t = Request_Cyclic_Counter_t{Value: CC_Unconfirmed, Barrier: make([]bool, N_ELEVATORS)}
				system_data.Elevator_Data[local_id].Cab_Requests[btn.Floor] = Update_CC(system_data.Elevator_Data[local_id].Cab_Requests[btn.Floor], tmp_cab_request, local_id)
			} else {
				var tmp_hall_request Request_Cyclic_Counter_t = Request_Cyclic_Counter_t{Value: CC_Unconfirmed, Barrier: make([]bool, N_ELEVATORS)}
				system_data.Hall_Request_Data[btn.Floor][btn.Button] = Update_CC(system_data.Hall_Request_Data[btn.Floor][btn.Button], tmp_hall_request, local_id)
			}

		//broadcast timer timeout
		case <-ticker.C: 
			//TODO: check broadcast System data is correct
			network_transmitter <- system_data 

		//Updates on the active peers list
		case Peers_Update := <-peersReciever:
			//TODO: format peersupdate to a bool list 
			Active_Peers = From_Peers_Update_To_Active_Peers(Peers_Update)
			for i := 0; i < N_ELEVATORS; i++{
				system_data.Elevator_Data[i].Is_Alive.Value = Active_Peers[i]
			}
		}
	}
}


