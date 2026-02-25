package message_sync

import (
	"fmt"
	"time"
	"../elevator"
	"Network_Driver/bcast"
	"Network_Driver/peers"
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
var elevator_network_list Elev_List_t = [0 0 0]

type Elev_List_t []bool
type Cyclic_Counter_t int

//Data type structs that include the data and a barrier
type Request_Cyclic_Counter_t struct{
	value Cyclic_Counter_t
	barrier Elev_List_t
}
type Is_Alive_Data_t struct{
	value bool
	barrier Elev_List_t
}
type Is_Able_Data_t struct{
	value bool
	barrier Elev_List_t
}
type Floor_Data_t struct{
	value int
	barrier Elev_List_t
}
type Elevator_Behaviour_Data_t struct{
	value elevator.Elevator_Behaviour_t
	barrier Elev_List_t
}
type Motor_Direction_Data_t struct{
	value elevator.Motor_Direction_t
	barrier Elev_List_t
}

type Elevator_Data_t struct {
	Id int
	Is_Alive Is_Alive_Data_t
	Is_Able Is_Able_Data_t
	Floor Floor_Data_t
	Elevator_Behaviour Elevator_Behaviour_Data_t
	Motor_Direction Motor_Direction_Data_t
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

func Message_Sync_Server(
	from_network_data <-chan System_Data_t, //channel for recieving new system data
	get_system_data <-chan Get_System_Data_t, //channel for other routines to get the current system data
	from_fsm_data <-chan System_Data_t
	peersReciever chan peers.PeerUpdate
	){
	var system_data System_Data_t
	var confirmed_system_data System_Data_t
	// Network variables
	var activePeers []string
	networkReciever := make(chan StateMsg)
	networkTransmitter := make(chan StateMsg)
	
	// Go routines from Network_Driver
	go bcast.Receiver(bcastPort, networkReciever)
	go bcast.Transmitter(bcastPort, networkTransmitter)

	drv_buttons := make(chan elevator.ButtonEvent_t)

	go elevator.PollButtons(drv_buttons)
	

	for {
		select{
		case reg := <- get_system_data:
			reg.Reply = system_data

		case fresh_data := <- from_network_data:
			system_data = On_Recieved_Fresh_Data(system_data, confirmed_system_data, fresh_data)

		case fresh_data := <- from_fsm_data:
			system_data.Elevator_Data[fresh_data.Id] = fresh_data.Elevator_Data[fresh_data.Id]
			
		case btn := <-drv_buttons:
			if btn.Button == elevator.BT_Cab {
				var tmp_cab_request Request_Cyclic_Counter_t = Request_Cyclic_Counter_t{value: CC_Unconfirmed, barrier: make(Elev_List_t, N_ELEVATORS)} //blind copy

			}


		}
		case peersUpdate := <-peersReciever:
			activePeers = peersUpdate.Peers
	}

}


