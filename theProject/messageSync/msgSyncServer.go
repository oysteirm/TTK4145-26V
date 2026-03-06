package messageSync

import (
	"time"
	"theProject/elevator"
	"theProject/networkDriver/peers"
	"theProject/networkDriver/bcast"
)	

/* map over data that is being syncronized
-----------------------------------
Elevator States:
[ 	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[N_FLOORS]],
	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[N_FLOORS]], 
	[ID		ALIVE 	IS_ABLE		FLOOR	EB		MD	Cab_Requests[N_FLOORS]]	]

Hall Requests:
HallRequestData[N_FLOORS][N_HALL_CALLS]

Every piece of data have a list with the elevators who agree with the information. 
If this list == elevator_network_list then we send this data have reached consensus and is put in confirmed data which is sent to HSA
-----------------------------------
*/

const (
	CC_Uninit 		CyclicCounter_t = -1
	CC_No 			CyclicCounter_t	= 0
	CC_Unconfirmed 	CyclicCounter_t	= 1
	CC_Confirmed 	CyclicCounter_t	= 2
	CC_Done 		CyclicCounter_t	= 3
)

const N_ELEVATORS int = 3

// List containing info about our network peers
// 1: part of network
// 0: not part of network
var activePeers []bool

type CyclicCounter_t int

//Data type structs that include the data and a Barrier
type RequestCyclicCounter_t struct{
	Value CyclicCounter_t
	Barrier []bool
}
/*
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
*/

//Datatype for elevator states with barrier
type ElevatorData_t struct {
	ID int
	IsAlive bool
	IsFunctional bool
	Floor int
	ElevatorBehaviour elevator.ElevatorBehaviour_t
	MotorDirection elevator.MotorDirection_t
	ElevatorBarrier []bool
	CabRequests []RequestCyclicCounter_t
}

//Datatype for multi elevator states and hall requests
type SystemData_t struct {
	ID int
	ElevatorData []ElevatorData_t
	HallRequestData [][2]RequestCyclicCounter_t
}

type GetSystemData_t struct{
	Reply SystemData_t
}

func MessageSyncServer(
	getSystemData <-chan GetSystemData_t, 	//channel for other routines to get the current system data
	dataFromFSM <-chan SystemData_t, 		//channel for recieving elevator data from elevator FSM
	dataToFSM chan<- SystemData_t, 			//channel for sending confirmed data to FSM
	peersReciever <-chan peers.PeerUpdate,	//channel for updating activePeers list
	localID int,							//ID of local elevator 
	){

	// Variables used to sync data
	var systemData SystemData_t
	var confirmedSystemData SystemData_t
	systemData, confirmedSystemData = InitSystemData(localID)
	var isConfirmedDataUpdated bool = false

	// Network channels and variable
	networkReceiver := make(chan SystemData_t)
	networkTransmitter := make(chan SystemData_t)
	bcastPort := 1234 //TODO: change this to a correct value

	// Go routines for communicating with other elevators
	go bcast.Receiver(bcastPort, networkReceiver)
	go bcast.Transmitter(bcastPort, networkTransmitter)
	
	// Ticker for periodic broadcasting 100Hz
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Go routine for button polling
	drvButtons := make(chan elevator.ButtonEvent_t)
	go elevator.PollButtons(drvButtons)
	
	for {
		select{
		//Someone needs the system data
		case reg := <- getSystemData:
			reg.Reply = systemData

		//We recieve new data from the network
		case freshData := <- networkReceiver:
			systemData, confirmedSystemData, isConfirmedDataUpdated = onReceivedFreshData(systemData, confirmedSystemData, freshData)

			//If we have new confirmed data, we sent it to the elevator FSM
			if isConfirmedDataUpdated {
				dataToFSM <- confirmedSystemData
			}

			//TODO: place them in elev_server
			lightCabLights(confirmedSystemData.ElevatorData[localID].CabRequests)
			lightHallLights(confirmedSystemData.HallRequestData)

		//We recieve data from the elevator FSM
		case freshData := <- dataFromFSM:
			systemData.ElevatorData[localID] = updateElevatorDataAboutSelf(systemData.ElevatorData[localID], freshData.ElevatorData[localID], localID)
			systemData.HallRequestData = updateHallRequestData(systemData.HallRequestData, freshData.HallRequestData, localID)
			
		//new buttonpress tries to change the CC to unconfirmed
		case btn := <-drvButtons:
			if btn.Button == elevator.BT_Cab {
				var tmpCabRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: make([]bool, N_ELEVATORS)}
				systemData.ElevatorData[localID].CabRequests[btn.Floor] = update_CC(systemData.ElevatorData[localID].CabRequests[btn.Floor], tmpCabRequest, localID)
			} else {
				var tmpHallRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: make([]bool, N_ELEVATORS)}
				systemData.HallRequestData[btn.Floor][btn.Button] = update_CC(systemData.HallRequestData[btn.Floor][btn.Button], tmpHallRequest, localID)
			}

		//broadcast timer timeout
		case <-ticker.C: 
			//TODO: check broadcast System data is correct
			networkTransmitter <- systemData 

		//Updates on the active peers list
		case peersUpdate := <-peersReciever:
			activePeers = fromPeersUpdateToActivePeers(peersUpdate)
			for i := 0; i < N_ELEVATORS; i++{
				//if there is new info, new barrier
				if systemData.ElevatorData[i].IsAlive != activePeers[i]{
					systemData.ElevatorData[i].IsAlive = activePeers[i]
					systemData.ElevatorData[i].ElevatorBarrier = make([]bool, N_ELEVATORS)
					systemData.ElevatorData[i].ElevatorBarrier[localID] = true
				}
			}
		}
	}
}


