package elevator

import (
	"fmt"
	"os"
	"strconv"
	"the_project/networkDriver/peers"
	RA "the_project/requestAssigner"
	"the_project/messageSync"
	"time"
)

const inactivityTimeout = 9 * time.Second
const obstructionTimeout = 5 * time.Second

func ElevatorServerMain(
	DataToFSM chan<- messageSync.ElevatorData_t,
	localID int,
	peersReceiver <-chan peers.PeerUpdate,
	getSystemData chan messageSync.GetSystemData_t,
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
				// Get system data from message sync
				systemDataRequest := messageSync.GetSystemData_t{Reply: messageSync.SystemData_t{}}
				getSystemData <- systemDataRequest
				systemData := systemDataRequest.Reply

				// Convert to RA_System_Data and get request assignments
				raSystemData := ra.Generating_RA_System_Data(systemData)
				raOutput := ra.Assign_Requests(raSystemData)

				// Convert RA_Output to motor direction and behavior
				pair := RAOutputToMotorDirectionPair(raOutput, localID, e_state)
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

//TODO: Helper function to convert ElevatorState_t to Elevator_Data_t for message sync
func BuildElevatorData(localID int, e_state ElevatorState_t) messageSync.ElevatorData_t {

}

//TODO: Helper function to convert RA_Output to MotorDirectionBehaviourPair using request assigner output
func RAOutputToMotorDirectionPair(raOutput RA.RA_Output, localID int, e_state ElevatorState_t) MotorDirectionBehaviourPair_t {

}
