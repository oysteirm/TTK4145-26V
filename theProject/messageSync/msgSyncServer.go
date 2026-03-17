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
Functionallity: 
	- MessageSyncServer recieves system data from other elevators from UDP bcast and, 
	  elevator data and done requests from the elevator FSM.
	- It also UDP bcast it's own systemdata to its peers. It does not recieve it's own bcasts.
	- Every state and requests of the whole system is stored in the SystemData_t struct.
	- Every elevator state and request have a list (barrier) with the elevators who have seen this information,
	  if this list == AcitvePeers list then this data have reached consensus on the network.
	- For requests this list is used to transition from unconfirmed -> confirmed and done -> no. 
	- For Elevator states it is used to update confirmed data with new elevator states that have reached consensus.

-----------------------------------
The Utilities:
	Update functions: 
		- Input / output functions that return an updated variable based on the data on input.
		- Uses ID, barriers and cyclic counter (CC) logic to decide what data to update.
	Helper functions:
		-Converting functions for peers.
		-Union and ActivePeers check for barriers.

-----------------------------------
Map over the SystemData_t struct that is being syncronized

	ID

	Elevator States:
	   [N_ELEVATORS]{ID,	IsAlive,	IS_Functinal,	Floor,		EB,		
	   MD,		E_Barrier, 	Cab_Requests[N_FLOORS]{Value, Barrier}},
		

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
-1: uninitialized,
0: no request
1: unconfirmed request
2: confirmed requests
3: requests done 
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

// Datatype for elevator states with barrier
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

// Datatype for multi elevator system,
type SystemData_t struct {
	ID              int
	ElevatorData    [config.N_ELEVATORS]ElevatorData_t
	HallRequestData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t
}


//Go rountinge used for syncronizing system data and sending data with consensus to the elevator FSM
func MessageSyncServer(
	elevatorDataFromFSM <-chan ElevatorData_t, 				//channel for recieving elevator data from elevator FSM
	requestsFrom_FSM <-chan []elevator_IO.ButtonEvent_t, 	//channel for recieving done requests from elevator FSM
	dataToFSM chan<- SystemData_t, 							//channel for sending confirmed data to FSM
	peersReciever <-chan peers.PeerUpdate, 					//channel for updating ActivePeers list
	localID int, 											//ID of local elevator
) {

	// Variables used to sync data
	var systemData SystemData_t 
	var confirmedSystemData SystemData_t
	systemData, confirmedSystemData = InitSystemData(localID)
	var isConfirmedDataUpdated bool = false
	ActivePeers = [config.N_ELEVATORS]bool{}
	ActivePeers[localID] = true

	// Network channels and variable
	networkReceiver := make(chan SystemData_t)
	networkTransmitter := make(chan SystemData_t)
	bcastPort := config.BCAST_PORT

	// Go routines for communicating with other elevators
	go bcast.Receiver(bcastPort, networkReceiver)
	go bcast.Transmitter(bcastPort, networkTransmitter)

	// Ticker for periodic broadcasting
	ticker := time.NewTicker(config.BCAST_PERIOD)
	defer ticker.Stop()

	// Go routine for request buttons polling
	drvButtons := make(chan elevator_IO.ButtonEvent_t)
	go elevator_IO.PollButtons(drvButtons)

	// Loop reacting to incoming data on channels
	for {
		select {

		// We recieve new data from the network
		case freshSystemData := <-networkReceiver:

			// Filtering out own messages
			if freshSystemData.ID != localID {

				// Update based on the newly received data
				systemData, confirmedSystemData, isConfirmedDataUpdated = onReceivedFreshData(systemData, confirmedSystemData, freshSystemData)

				// If we have new confirmed data, we sent it to the elevator FSM
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

		// We receive requests that elevator FSM have done
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
			// Update confirmed data and send to elevator FSM if we did
			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				systemPrintHorizontal(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		// We recieve data from the elevator FSM
		case freshData := <-elevatorDataFromFSM:
			
			systemData.ElevatorData[localID] = updateElevatorDataAboutSelf(systemData.ElevatorData[localID], freshData, localID)
			// Update confirmed data and send to elevator FSM if we did
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
				systemData.ElevatorData[localID].CabRequests[btn.Floor] = update_CC(systemData.ElevatorData[localID].CabRequests[btn.Floor], tmpCabRequest, localID)
			} else {
				var tmpHallRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: [config.N_ELEVATORS]bool{}}
				systemData.HallRequestData[btn.Floor][btn.Button] = update_CC(systemData.HallRequestData[btn.Floor][btn.Button], tmpHallRequest, localID)
			}
			// Update confirmed data and send to elevator FSM if we did
			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				systemPrintHorizontal(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		// Time to broadcast the systemData
		case <-ticker.C:
			networkTransmitter <- systemData

		// We recieve updates on the active peers list
		case peersUpdate := <-peersReciever:
			// Use the new information to set the new AvtivePeers list, 
			// making sure we are alive
			ActivePeers = fromPeersUpdateToActivePeers(peersUpdate)
			ActivePeers[localID] = true
			// Update CC for requests with the new ActivePeers list
			systemData = update_CC_ForCurrentPeers(systemData, localID)

			for i := 0; i < config.N_ELEVATORS; i++ {
				// If there is new info, new barrier
				if systemData.ElevatorData[i].IsAlive != ActivePeers[i] {
					systemData.ElevatorData[i].IsAlive = ActivePeers[i]
					systemData.ElevatorData[i].ElevatorBarrier = [config.N_ELEVATORS]bool{}
					systemData.ElevatorData[i].ElevatorBarrier[localID] = true
				}
			}
			// Update confirmed data and send to elevator FSM if we did
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

