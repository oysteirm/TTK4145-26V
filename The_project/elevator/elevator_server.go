package elevator

import (
	"os"
	"the_project/Network_Driver/peers"
	"the_project/message_sync"
	"time"
)

const inactivityTimeout = 9 * time.Second
const obstructionTimeout = 5 * time.Second

func Elevator_Server_main(
	to_fsm_data chan<- message_sync.Elevator_Data_t,
	local_id int,
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
	go Elevator_Server(commands)

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
	to_fsm_data <- BuildElevatorData(local_id, e_state)

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

		// Send state to msg_sync (non-blocking)
		select {
		case to_fsm_data <- BuildElevatorData(local_id, GetState(commands)):
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
func BuildElevatorData(local_id int, e_state ElevatorState_t) message_sync.Elevator_Data_t {
	// Initialize barrier with this elevator marked as sender
	barrier := make(message_sync.Elev_List_t, message_sync.N_ELEVATORS)
	barrier[local_id] = true

	// Convert cab requests from ElevatorState_t.Requests
	cab_requests := make([]message_sync.Request_Cyclic_Counter_t, N_FLOORS)
	for floor := 0; floor < N_FLOORS; floor++ {
		if e_state.Requests[floor][BT_Cab] {
			cab_requests[floor] = message_sync.Request_Cyclic_Counter_t{
				Value:   message_sync.CC_Unconfirmed,
				Barrier: barrier,
			}
		} else {
			cab_requests[floor] = message_sync.Request_Cyclic_Counter_t{
				Value:   message_sync.CC_Uninit,
				Barrier: barrier,
			}
		}
	}

	elevator_data := message_sync.Elevator_Data_t{
		Id:          local_id,
		Msg_counter: 0,
		Is_Alive: message_sync.Is_Alive_Data_t{
			Value:   true,
			Barrier: barrier,
		},
		Is_Functional: message_sync.Is_Functional_Data_t{
			Value:   e_state.IsFunctional,
			Barrier: barrier,
		},
		Floor: message_sync.Floor_Data_t{
			Value:   e_state.Floor,
			Barrier: barrier,
		},
		Elevator_Behaviour: message_sync.Elevator_Behaviour_Data_t{
			Value:   e_state.ElevatorBehaviour,
			Barrier: barrier,
		},
		Motor_Direction: message_sync.Motor_Direction_Data_t{
			Value:   e_state.MotorDirection,
			Barrier: barrier,
		},
		Cab_Requests: cab_requests,
	}

	return elevator_data
}
