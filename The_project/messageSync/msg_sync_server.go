package messageSync

import (
	"time"
	"the_project/elevator"
	"the_project/Network_Driver/peers"
	"the_project/Network_Driver/bcast"
)	
/* map over data that is being syncronized
-----------------------------------
Elevator States:
[ 	[ID		ALIVE 	IsFunctional		FLOOR	EB		MD	CabRequests[N_FLOORS]],
	[ID		ALIVE 	IsFunctional	    FLOOR	EB		MD	CabRequests[N_FLOORS]], 
	[ID		ALIVE 	IsFunctional		FLOOR	EB		MD	CabRequests[N_FLOORS]]	]

Hall Requests:
Hall_Request_Data[N_FLOORS][N_HALL_CALLS]

Every piece of data have a list with the elevators who agree with the information. 
If this list == elevator_network_list then we send this data have reached consensus and is put in confirmed data which is sent to HSA
-----------------------------------
*/

const (
	CC_Uninit CyclicCounter_t 	= -1
	CC_No 						= 0
	CC_Unconfirmed 				= 1
	CC_Confirmed 				= 2
	CC_Done 					= 3
)

const N_ELEVATORS int = 3
const btns_UP_and_Down int = 2

type ElevList_t []bool
type CyclicCounter_t int

//Data type structs that include the data and a barrier
type RequestCyclicCounter_t struct{
	Value CyclicCounter_t
	Barrier ElevList_t
}
type IsAliveData_t struct{
	Value bool
	Barrier ElevList_t
}
type IsFunctionalData_t struct{
	Value bool
	Barrier ElevList_t
}
type FloorData_t struct{
	Value int
	Barrier ElevList_t
}
type ElevatorBehaviourData_t struct{
	Value elevator.ElevatorBehaviour_t
	Barrier ElevList_t
}
type MotorDirectionData_t struct{
	Value elevator.MotorDirection_t
	Barrier ElevList_t
}
//Datatype for elevator states with barriers
type ElevatorData_t struct {
	Id int
	//MsgCounter uint64
	IsAlive IsAliveData_t
	IsFunctional IsFunctionalData_t
	Floor FloorData_t
	ElevatorBehaviour ElevatorBehaviourData_t
	MotorDirection MotorDirectionData_t
	CabRequests []RequestCyclicCounter_t
}
//Datatype for multi elevator states and hall requests
type SystemData_t struct {
	Id int
	ElevatorData []ElevatorData_t
	HallRequestData [][2]RequestCyclicCounter_t
}

type GetSystemData_t struct{
	Reply SystemData_t
}

func MessageSyncServer(
	fromNetworkData <-chan SystemData_t, //channel for recieving new system data
	getSystemData <-chan GetSystemData_t, //channel for other routines to get the current system data
	fromFsmData <-chan ElevatorData_t, //channel for recieving elevator data from fsm
	peersReciever <-chan peers.PeerUpdate,
	localID int,
	){
	// 
	var systemData SystemData_t
	var confirmedSystemData SystemData_t
	systemData, confirmedSystemData = Init_systemData(localID)
	var iisConfirmedDataUpdated bool = false

	// Network variables
	var activePeers []string
	networkReciever := make(chan systemData_t)
	networkTransmitter := make(chan systemData_t)
	bcastPort := 1234 //TODO: change this to a correct value

	// Go routines from Network_Driver
	go bcast.Receiver(bcastPort, networkReciever)
	go bcast.Transmitter(bcastPort, networkTransmitter)
	
	// Timer for broadcasting
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Go routine for button polling
	drvButtons := make(chan elevator.ButtonEvent_t)
	go elevator.PollButtons(drvButtons)
	

	for {
		select{
		case reg := <- get_systemData:
			reg.Reply = systemData

		case freshData := <- fromNetworkData:
			systemData, confirmedSystemData, isConfirmedDataUpdated = OnRecievedFreshData(systemData, confirmedSystemData, freshData)

			if isConfirmedDataUpdated{
				//TODO: send confirmed_data til elev_FSM
			}
			//use confirmed_data for light contract 
			//TODO: write these functions and place them in elev_server
			LightCabLights(confirmedSystemData.ElevatorData[localID].CabRequests)
			LightHallLights(confirmedSystemData.HallRequestData)

		case freshData := <- fromFsmData:
			systemData.ElevatorData[localID] = UpdateElevatorDataAboutSelf(systemData.ElevatorData[localID], freshData, localID)
			
		//buttonpress tries to change the CC to unconfirmed
		case btn := <-drvButtons:
			if btn.Button == elevator.BT_Cab {
				var tmpCabRequest Request_Cyclic_Counter_t = Request_Cyclic_Counter_t{Value: CC_Unconfirmed, Barrier: make(Elev_List_t, N_ELEVATORS)} //blind copy

			}
		case //broadcast timer timeout
			systemData.ElevatorData[localID].MsgCounter


		case peersUpdate := <-peersReciever:
			//TODO: format peersupdate to a bool list 
			activePeers = peersUpdate.Peers

		}
	}
}


