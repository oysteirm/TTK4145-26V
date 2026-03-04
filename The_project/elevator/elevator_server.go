package elevator

import (
	"os"
	"the_project/Network_Driver/peers"
	"the_project/messageSync"
	"time"
)

const inactivityTimeout = 9 * time.Second
const obstructionTimeout = 5 * time.Second

func ElevatorServerMain(
	DataToFSM chan<- messageSync.ElevatorData_t,
	localID int,
	peersReceiver <-chan peers.PeerUpdate,
) {
	commands := make(chan Command_t)

	floorSensor := make(chan int)
	obstructionSwitch := make(chan bool)
	doorTimerStart := make(chan time.Duration)
	doorTimerStop := make(chan struct{})
	doorTimerTimeout := make(chan struct{})

	go ElevatorServer(commands)

	go PollFloorSensor(floorSensor)
	go PollObstructionSwitch(obstructionSwitch)

	go DoorTimer(doorTimerStart, doorTimerStop, doorTimerTimeout)

	doorTimer := time.NewTimer(0)
	doorTimer.Stop()
	obstructionTimer := time.NewTimer(obstructionTimeout)
	obstructionTimer.Stop()
	inactivityTimer := time.NewTimer(inactivityTimeout)
	inactivityTimer.Stop()

	OnInitBetweenFloors(commands)

	isObstructed := false
	var activePeers []string

	e_state := GetState(commands)
	DataToFSM <- BuildElevatorData(localID, e_state)

	for {
		select {

		case newFloor := <-floorSensor:
			SetFloorIndicator(newFloor)
			commands <- SetFloor_t{Floor: newFloor}
			e_state := GetState(commands)

			if e_state.ElevatorBehaviour == EB_Moving {
				if RequestsShouldStop(e_state) {
					SetMotorDirection(MD_Stop)
					SetDoorOpenLamp(true)

				e_state = RequestsClearAtCurrentFloor(e_state)
				commands <- SetState_t{ElevatorState: e_state}
				commands <- SetMotorDirection_t{MotorDirection: MD_Stop}
				commands <- SetElevatorBehaviour_t{ElevatorBehaviour: EB_DoorOpen}
					doorTimerStop <- struct{}{}
					doorTimerStart <- e_state.DoorOpenDuration

					ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
				}
			}
			UpdateFunctionalStatus(commands)

		case isObstructed = <-obstructionSwitch:
			e_state := GetState(commands)
			if e_state.ElevatorBehaviour == EB_DoorOpen {
				ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
			}
			UpdateFunctionalStatus(commands)

		case <-doorTimerTimeout:
			e_state := GetState(commands)
			if e_state.ElevatorBehaviour == EB_DoorOpen && !isObstructed {
				pair := RequestsChooseDirection(e_state)
				commands <- SetMotorDirection_t{MotorDirection: pair.MotorDirection}
				commands <- SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}

				e_state = GetState(commands)

				switch e_state.ElevatorBehaviour {
				case EB_DoorOpen:
					ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
				case EB_Idle:
					SetDoorOpenLamp(false)
					SetMotorDirection(e_state.MotorDirection)
					doorTimer.Stop()
					inactivityTimer.Stop()
				case EB_Moving:
					inactivityTimer.Reset(inactivityTimeout)
					SetDoorOpenLamp(false)
					SetMotorDirection(e_state.MotorDirection)
					doorTimer.Stop()
				}
			}
			UpdateFunctionalStatus(commands)

		case <-obstructionTimer.C:
			if len(activePeers) > 1 { // If we are alone, we cannot reset on obstruction or inactivity without loosing orders
				os.Exit(1) // 
			} else {
				obstructionTimer.Reset(obstructionTimeout)
			}

		case <-inactivityTimer.C:
			if len(activePeers) > 1 {
				os.Exit(2) // 
			} else {
				inactivityTimer.Reset(inactivityTimeout)
			}

		case peersUpdate := <-peersReceiver:
			activePeers = peersUpdate.Peers
		}

		select {
		case DataToFSM <- BuildElevatorData(localID, GetState(commands)):
		default:
		}
	}
}

func ResetTimers(isObstructed bool, obstructionTimer *time.Timer, doorTimer *time.Timer, inactivityTimer *time.Timer, doorOpenDuration time.Duration) {
	if isObstructed {
		obstructionTimer.Reset(obstructionTimeout)
	} else {
		obstructionTimer.Stop()
	}
	doorTimer.Reset(doorOpenDuration)
}

// Helper function to convert ElevatorState_t to Elevator_Data_t for message sync
func BuildElevatorData(localID int, e_state ElevatorState_t) messageSync.ElevatorData_t {
	// Initialize barrier with this elevator marked as sender
	barrier := make(messageSync.ElevList_t, messageSync.N_ELEVATORS)
	barrier[localID] = true

	// Convert cab requests from ElevatorState_t.Requests
	cabRequests := make([]messageSync.RequestCyclicCounter_t, N_FLOORS)
	for floor := 0; floor < N_FLOORS; floor++ {
		if e_state.Requests[floor][BT_Cab] {
			cabRequests[floor] = messageSync.RequestCyclicCounter_t{
				Value:   messageSync.CC_Unconfirmed,
				Barrier: barrier,
			}
		} else {
			cabRequests[floor] = messageSync.RequestCyclicCounter_t{
				Value:   messageSync.CC_Uninit,
				Barrier: barrier,
			}
		}
	}

	elevatorData := messageSync.ElevatorData_t{
		Id:          localID,
		MsgCounter: 0,
		IsAlive: messageSync.IsAliveData_t{
			Value:   true,
			Barrier: barrier,
		},
		IsFunctional: messageSync.IsFunctionalData_t{
			Value:   e_state.IsFunctional,
			Barrier: barrier,
		},
		Floor: messageSync.FloorData_t{
			Value:   e_state.Floor,
			Barrier: barrier,
		},
		ElevatorBehaviour: messageSync.ElevatorBehaviourData_t{
			Value:   e_state.ElevatorBehaviour,
			Barrier: barrier,
		},
		MotorDirection: messageSync.MotorDirectionData_t{
			Value:   e_state.MotorDirection,
			Barrier: barrier,
		},
		CabRequests: cabRequests,
	}

	return elevatorData
}
