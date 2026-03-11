package messageSync

import (
	"fmt"
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/networkDriver/bcast"
	"theProject/networkDriver/peers"
	"time"
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
	CC_Uninit      CyclicCounter_t = -1
	CC_No          CyclicCounter_t = 0
	CC_Unconfirmed CyclicCounter_t = 1
	CC_Confirmed   CyclicCounter_t = 2
	CC_Done        CyclicCounter_t = 3
)

// List containing info about our network peers
// 1: part of network
// 0: not part of network

var ActivePeers [config.N_ELEVATORS]bool

type CyclicCounter_t int

// Data type structs that include the data and a Barrier
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

// Datatype for multi elevator states and hall requests
type SystemData_t struct {
	ID              int
	ElevatorData    [config.N_ELEVATORS]ElevatorData_t
	HallRequestData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t
}

func MessageSyncServer(
	elevatorDataFromFSM <-chan ElevatorData_t, //channel for recieving elevator data from elevator FSM
	requestsFrom_FSM <-chan []elevator_IO.ButtonEvent_t, //channel for recieving done requests from elevator FSM
	dataToFSM chan<- SystemData_t, //channel for sending confirmed data to FSM
	peersReciever <-chan peers.PeerUpdate, //channel for updating ActivePeers list
	localID int, //ID of local elevator
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
	bcastPort := config.B_CAST_PORT

	// Go routines for communicating with other elevators
	go bcast.Receiver(bcastPort, networkReceiver)
	go bcast.Transmitter(bcastPort, networkTransmitter)

	// Ticker for periodic broadcasting 100Hz
	ticker := time.NewTicker(config.B_CAST_PERIOD)
	defer ticker.Stop()

	// Go routine for button polling
	drvButtons := make(chan elevator_IO.ButtonEvent_t)
	go elevator_IO.PollButtons(drvButtons)

	for {
		select {

		//We recieve new data from the network
		case freshSystemData := <-networkReceiver:

			if freshSystemData.ID != localID {

				systemData, confirmedSystemData, isConfirmedDataUpdated = OnReceivedFreshData(systemData, confirmedSystemData, freshSystemData)

				//If we have new confirmed data, we sent it toupdate_CC the elevator FSM
				if isConfirmedDataUpdated {
					if confirmedSystemData.ElevatorData[localID].Floor != -1 {
						fmt.Println("Sending new confirmed data to FSM")
						ChatGPT_SystemPrint(confirmedSystemData)
						dataToFSM <- confirmedSystemData
						isConfirmedDataUpdated = false
					}
				}
			}

		case freshRequestsToDone := <-requestsFrom_FSM:
			currentFloor := systemData.ElevatorData[localID].Floor

			for _, btnEvnt := range freshRequestsToDone {
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
				ChatGPT_SystemPrint(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		//We recieve data from the elevator FSM
		case freshData := <-elevatorDataFromFSM:
			systemData.ElevatorData[localID] = UpdateElevatorDataAboutSelf(systemData.ElevatorData[localID], freshData, localID)

			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)

			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				ChatGPT_SystemPrint(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		//new buttonpress tries to change the CC to unconfirmed
		case btn := <-drvButtons:
			if btn.Button == elevator_IO.BT_Cab {
				var tmpCabRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: [config.N_ELEVATORS]bool{}}
				systemData.ElevatorData[localID].CabRequests[btn.Floor] = update_CC(systemData.ElevatorData[localID].CabRequests[btn.Floor], tmpCabRequest, localID)
			} else {
				var tmpHallRequest RequestCyclicCounter_t = RequestCyclicCounter_t{Value: CC_Unconfirmed, Barrier: [config.N_ELEVATORS]bool{}}
				systemData.HallRequestData[btn.Floor][btn.Button] = update_CC(systemData.HallRequestData[btn.Floor][btn.Button], tmpHallRequest, localID)
			}

			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)

			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				ChatGPT_SystemPrint(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}

		//broadcast timer timeout
		case <-ticker.C:
			//TODO: check broadcast System data is correct
			networkTransmitter <- systemData

		//Updates on the active peers list
		case peersUpdate := <-peersReciever:
			ActivePeers = fromPeersUpdateToActivePeers(peersUpdate)
			ActivePeers[localID] = true
			systemData = update_CC_ForCurrentPeers(systemData, localID)
			for i := 0; i < config.N_ELEVATORS; i++ {
				//if there is new info, new barrier
				if systemData.ElevatorData[i].IsAlive != ActivePeers[i] {
					systemData.ElevatorData[i].IsAlive = ActivePeers[i]
					systemData.ElevatorData[i].ElevatorBarrier = [config.N_ELEVATORS]bool{}
					systemData.ElevatorData[i].ElevatorBarrier[localID] = true
				}
			}
			confirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)

			if isConfirmedDataUpdated {
				fmt.Println("Sending new confirmed data to FSM")
				ChatGPT_SystemPrint(confirmedSystemData)
				dataToFSM <- confirmedSystemData
				isConfirmedDataUpdated = false
			}
		}
	}
}

// print from chatGPT
func ChatGPT_SystemPrint(systemData SystemData_t) {

	// Top line
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Print("  +--------------------+")
	}
	fmt.Println()

	// Elevator headers
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | Elevator: %-2d       |", i)
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  |IsAlive = %-9t |", systemData.ElevatorData[i].IsAlive)
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  |IsFunctional = %-2t |", systemData.ElevatorData[i].IsFunctional)
	}
	fmt.Println()

	// Floor
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | floor = %-2d         |", systemData.ElevatorData[i].Floor)
	}
	fmt.Println()

	// Direction
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | dirn  = %-10s |",
			ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection))
	}
	fmt.Println()

	// Behaviour
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | behav = %-10s |",
			ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour))
	}
	fmt.Println()

	// Button header
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Print("  |  | up  | dn  | cab |")
	}
	fmt.Println()

	// Floors
	for f := elevator_IO.N_FLOORS - 1; f >= 0; f-- {

		for i := 0; i < config.N_ELEVATORS; i++ {

			fmt.Printf("  | %d", f)

			for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {

				if (f == elevator_IO.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
					(f == 0 && btn == elevator_IO.BT_HallDown) {

					fmt.Print("|     ")

				} else {

					if btn == elevator_IO.BT_Cab {
						if CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
							fmt.Print("|  #  ")
						} else {
							fmt.Print("|  -  ")
						}
					} else {
						if CC_ToBool(systemData.HallRequestData[f][btn].Value) {
							fmt.Print("|  #  ")
						} else {
							fmt.Print("|  -  ")
						}
					}

				}
			}

			fmt.Print("|")
		}

		fmt.Println()
	}

	// Bottom line
	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Print("  +--------------------+")
	}
	fmt.Println()
}

func CC_ToBool(cc CyclicCounter_t) bool {
	switch cc {
	case CC_Confirmed, CC_Done:
		return true
	default:
		return false
	}
}

func ElevatorBehaviourToString(eb elevator_IO.ElevatorBehaviour_t) string {
	switch eb {
	case elevator_IO.EB_Idle:
		return "idle"
	case elevator_IO.EB_DoorOpen:
		return "doorOpen"
	case elevator_IO.EB_Moving:
		return "moving"
	default:
		return "UNDEFINED"
	}
}

func ElevatorDirnToString(d elevator_IO.MotorDirection_t) string {
	switch d {
	case elevator_IO.MD_Up:
		return "up"
	case elevator_IO.MD_Down:
		return "down"
	case elevator_IO.MD_Stop:
		return "stop"
	default:
		return "UNDEFINED"
	}
}
