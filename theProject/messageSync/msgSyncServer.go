package messageSync

import (
	"fmt"
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/networkDriver/bcast"
	"theProject/networkDriver/peers"
	"time"
)

/* 
-----------------------------------
Functionality: 
	- MessageSyncServer receives elevator data and done requests from the elevator FSM,
	  and receives system data from other elevators via UDP bcasts.
	- It also UDP bcasts its own system data to its peers. It does not receive its own bcasts.
	- Every state and request of the whole system is stored in the SystemData_t struct.
	- Every elevator state and request has a list (barrier) of the elevators who have seen this information,
	  if this list == AcitvePeers list, then this data has reached consensus on the network.
	- For requests, this list is used to transition from unconfirmed -> confirmed and done -> no. 
	- For elevator states, it is used to update confirmed data with new elevator states that have reached consensus.
-----------------------------------
Map over the SystemData_t struct that is being synchronized

	ID

	Elevator States:
	   [N_ELEVATORS]{
	   ID,	
	   IsAlive,	
	   IsFunctinal,	
	   Floor,		
	   EB,		
	   MD,		
	   E_Barrier, 	
	   Cab_Requests[N_FLOORS]{Value, Barrier}},
		
	Hall Requests:
		HallRequestData[N_FLOORS][N_UP_DOWN] {Value,	Barrier}
-----------------------------------
*/

/* 
List containing info about our network peers
1: part of network
0: not part of network
Compared with barriers with func checkBarrier()
Updated through peerUpdates 
*/
var ActivePeers [config.N_ELEVATORS]bool


/* 
Datatype for cyclic counter states
-1: uninitialized
 0: no request
 1: unconfirmed request
 2: confirmed request
 3: request done 
*/
type CyclicCounter_t int

const (
	CC_Uninit      CyclicCounter_t = -1
	CC_No          CyclicCounter_t = 0
	CC_Unconfirmed CyclicCounter_t = 1
	CC_Confirmed   CyclicCounter_t = 2
	CC_Done        CyclicCounter_t = 3
)

// Datatype for cyclic counter value with a corresponding barrier 
type RequestCyclicCounter_t struct {
	Value   CyclicCounter_t
	Barrier [config.N_ELEVATORS]bool
}

type ElevatorData_t struct {
	ID                int
	IsAlive           bool
	IsFunctional      bool
	Floor             int
	ElevatorBehaviour elevator_IO.ElevatorBehaviour_t
	MotorDirection    elevator_IO.MotorDirection_t
	ElevatorBarrier   [config.N_ELEVATORS]bool
	CabRequests       [config.N_FLOORS]RequestCyclicCounter_t
}

type SystemData_t struct {
	ID              int
	ElevatorData    [config.N_ELEVATORS]ElevatorData_t
	HallRequestData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t
}

func MessageSyncServer(
	elevatorDataFromFSM <-chan ElevatorData_t, 		
	requestsFrom_FSM <-chan []elevator_IO.ButtonEvent_t, 
	dataToFSM chan<- SystemData_t, 							
	peersReciever <-chan peers.PeerUpdate, 			
	localID int, 		
) {

	// Variables used to sync data
	var systemData SystemData_t 
	var confirmedSystemData SystemData_t
	systemData, confirmedSystemData = InitSystemData(localID)
	var isConfirmedDataUpdated bool = false
	ActivePeers = [config.N_ELEVATORS]bool{}
	ActivePeers[localID] = true

	// Broadcasting with our peers
	networkReceiver := make(chan SystemData_t)
	networkTransmitter := make(chan SystemData_t)
	bcastPort := config.BCAST_PORT
	go bcast.Receiver(bcastPort, networkReceiver)
	go bcast.Transmitter(bcastPort, networkTransmitter)

	// Ticker for periodic broadcasting
	bcastTicker := time.NewTicker(config.BCAST_PERIOD)
	defer bcastTicker.Stop()

	// Go routine for request buttons polling
	drvButtons := make(chan elevator_IO.ButtonEvent_t)
	go elevator_IO.PollButtons(drvButtons)

	// Loop reacting to incoming data on channels
	for {
		select {

		case freshSystemData := <-networkReceiver:

			// Filtering out own messages
			if freshSystemData.ID != localID {

				systemData, confirmedSystemData, isConfirmedDataUpdated = onReceivedFreshData(systemData, confirmedSystemData, freshSystemData)

				if isConfirmedDataUpdated {
					// Filtering out unitialized floor value
					if confirmedSystemData.ElevatorData[localID].Floor != -1 {
						fmt.Println("Sending new confirmed data to FSM")
						systemPrintHorizontal(confirmedSystemData)
						dataToFSM <- confirmedSystemData
						isConfirmedDataUpdated = false
					}
				}
			}

		case freshRequestsToDone := <-requestsFrom_FSM:
			currentFloor := systemData.ElevatorData[localID].Floor

			for _, btnEvnt := range freshRequestsToDone {
				// Filtering out requests based on floor
				if btnEvnt.Floor != currentFloor {
					continue
				}
				if btnEvnt.Button == elevator_IO.BT_Cab {
					tempCabRequests := RequestCyclicCounter_t{Value: CC_Done, Barrier: [config.N_ELEVATORS]bool{}}
					tempCabRequests.Barrier[localID] = true
					systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor] = update_CC(systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor], tempCabRequests, localID)
				} else {
					tempHallRequests := RequestCyclicCounter_t{Value: CC_Done, Barrier: [config.N_ELEVATORS]bool{}}
					tempHallRequests.Barrier[localID] = true
					systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button] = update_CC(systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button], tempHallRequests, localID)
				}
			}

			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				systemPrintHorizontal(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		case freshData := <-elevatorDataFromFSM:
			
			systemData.ElevatorData[localID] = updateElevatorDataAboutSelf(systemData.ElevatorData[localID], freshData, localID)

			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				systemPrintHorizontal(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		// Buttonpress for a request, tries to change CC to uncomfirmed
		case btn := <-drvButtons:
			if btn.Button == elevator_IO.BT_Cab {
				var tmpCabRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: [config.N_ELEVATORS]bool{}}
				tmpCabRequest.Barrier[localID] = true
				systemData.ElevatorData[localID].CabRequests[btn.Floor] = update_CC(systemData.ElevatorData[localID].CabRequests[btn.Floor], tmpCabRequest, localID)
			} else {
				var tmpHallRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: [config.N_ELEVATORS]bool{}}
				tmpHallRequest.Barrier[localID] = true
				systemData.HallRequestData[btn.Floor][btn.Button] = update_CC(systemData.HallRequestData[btn.Floor][btn.Button], tmpHallRequest, localID)
			}

			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				systemPrintHorizontal(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		case <-bcastTicker.C:
			networkTransmitter <- systemData

		case peersUpdate := <-peersReciever:
			// Use the new information to set the new ActivePeers list, making sure we are alive
			ActivePeers = fromPeersUpdateToActivePeers(peersUpdate)
			ActivePeers[localID] = true

			systemData = update_CC_ForCurrentPeers(systemData, localID)

			for i := 0; i < config.N_ELEVATORS; i++ {
				// If there is new info, new barrier
				if systemData.ElevatorData[i].IsAlive != ActivePeers[i] {
					systemData.ElevatorData[i].IsAlive = ActivePeers[i]
					systemData.ElevatorData[i].ElevatorBarrier = [config.N_ELEVATORS]bool{}
					systemData.ElevatorData[i].ElevatorBarrier[localID] = true
				}
			}

			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				systemPrintHorizontal(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}
		}
	}
}