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
	toFsmData chan<- messageSync.ElevatorData_t,
	localID int,
	peersReceiver <-chan peers.PeerUpdate,
) {
	// Create internal channels for state machine
	commands := make(chan Command_t)

	// Create event channels
	floorSensor := make(chan int)
	obstructionSwitch := make(chan bool)
	doorTimerStart := make(chan time.Duration)
	doorTimerStop := make(chan struct{})
	doorTimerTimeout := make(chan struct{})

	// Launch main elevator state machine (internal)
	go ElevatorServer(commands)

	// Launch polling goroutines
	go PollFloorSensor(floorSensor)
	go PollObstructionSwitch(obstructionSwitch)

	// Launch door timer
	go DoorTimer(doorTimerStart, doorTimerStop, doorTimerTimeout)

	// Initialize timers
	doorTimer := time.NewTimer(0)
	doorTimer.Stop()
	obstructionTimer := time.NewTimer(obstructionTimeout)
	obstructionTimer.Stop()
	inactivityTimer := time.NewTimer(inactivityTimeout)
	inactivityTimer.Stop()

	// Initialize elevator (move down between floors)
	OnInitBetweenFloors(commands)

	// Track obstruction state
	isObstructed := false
	var activePeers []string

	// Send initial state
	e_state := GetState(commands)
	toFsmData <- BuildElevatorData(localID, e_state)

	for {
		select {
		// Floor arrival
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

		// Obstruction switch
		case isObstructed = <-obstructionSwitch:
			e_state := GetState(commands)
			if e_state.ElevatorBehaviour == EB_DoorOpen {
				ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
			}
			UpdateFunctionalStatus(commands)

		// Door timeout
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

		// Obstruction timeout - safety check
		case <-obstructionTimer.C:
			if len(activePeers) > 1 {
				os.Exit(1) // Obstruction timeout - emergency exit
			} else {
				obstructionTimer.Reset(obstructionTimeout)
			}

		// Inactivity timeout - safety check
		case <-inactivityTimer.C:
			if len(activePeers) > 1 {
				os.Exit(2) // Inactivity timeout - emergency exit
			} else {
				inactivityTimer.Reset(inactivityTimeout)
			}

		// Peers update
		case peersUpdate := <-peersReceiver:
			activePeers = peersUpdate.Peers
		}

		// Send state to msgSync (non-blocking)
		select {
		case toFsmData <- BuildElevatorData(localID, GetState(commands)):
		default:
		}
	}
}

// ResetTimers resets the door and obstruction timers appropriately
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
