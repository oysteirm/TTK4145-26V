package elevatorServer

import (
	//"fmt"
	"os"
	"theProject/networkDriver/peers"
	ra "theProject/requestAssigner"
	"theProject/messageSync"
	e "theProject/elevator"
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
	commands := make(chan e.Command_t)

	floorSensor := make(chan int)
	obstructionSwitch := make(chan bool)
	doorTimerStart := make(chan time.Duration)
	doorTimerStop := make(chan struct{})
	doorTimerTimeout := make(chan struct{})

	go e.ElevatorServer(commands)

	go e.PollFloorSensor(floorSensor)
	go e.PollObstructionSwitch(obstructionSwitch)

	go e.DoorTimer(doorTimerStart, doorTimerStop, doorTimerTimeout)

	doorTimer := time.NewTimer(0)
	doorTimer.Stop()
	obstructionTimer := time.NewTimer(obstructionTimeout)
	obstructionTimer.Stop()
	inactivityTimer := time.NewTimer(inactivityTimeout)
	inactivityTimer.Stop()

	e.OnInitBetweenFloors(commands)

	isObstructed := false
	var activePeers []string

	e_state := e.GetState(commands)
	DataToFSM <- BuildElevatorData(localID, e_state)

	for {
		select {

		case newFloor := <-floorSensor:
			e.SetFloorIndicator(newFloor)
			commands <- e.SetFloor_t{Floor: newFloor}
			e_state := e.GetState(commands)

			if e_state.ElevatorBehaviour == e.EB_Moving {
				if e.RequestsShouldStop(e_state) {
					e.SetMotorDirection(e.MD_Stop)
					e.SetDoorOpenLamp(true)

					e_state = e.RequestsClearAtCurrentFloor(e_state)
					commands <- e.SetState_t{ElevatorState: e_state}
					commands <- e.SetMotorDirection_t{MotorDirection: e.MD_Stop}
					commands <- e.SetElevatorBehaviour_t{ElevatorBehaviour: e.EB_DoorOpen}
					doorTimerStop <- struct{}{}
					doorTimerStart <- e_state.DoorOpenDuration

					ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
				}
			}
			//TODO: UpdateFunctionalStatus(commands)

		case isObstructed = <-obstructionSwitch:
			e_state := e.GetState(commands)
			if e_state.ElevatorBehaviour == e.EB_DoorOpen {
				ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
			}
			//TODO: UpdateFunctionalStatus(commands)

		case <-doorTimerTimeout:
			e_state := e.GetState(commands)
			if e_state.ElevatorBehaviour == e.EB_DoorOpen && !isObstructed {
				// Get system data from message sync
				systemDataRequest := messageSync.GetSystemData_t{Reply: messageSync.SystemData_t{}}
				getSystemData <- systemDataRequest
				systemData := systemDataRequest.Reply

				// Convert to RA_System_Data and get request assignments
				raSystemData := ra.Generating_RA_SystemData(systemData)
				raOutput := ra.AssignRequests(raSystemData)

				// Convert RA_Output to motor direction and behavior
				pair := RAOutputToMotorDirectionPair(raOutput, localID, e_state)
				commands <- e.SetMotorDirection_t{MotorDirection: pair.MotorDirection}
				commands <- e.SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}

				e_state = e.GetState(commands)

				switch e_state.ElevatorBehaviour {
				case e.EB_DoorOpen:
					ResetTimers(isObstructed, obstructionTimer, doorTimer, inactivityTimer, e_state.DoorOpenDuration)
				case e.EB_Idle:
					e.SetDoorOpenLamp(false)
					e.SetMotorDirection(e_state.MotorDirection)
					doorTimer.Stop()
					inactivityTimer.Stop()
				case e.EB_Moving:
					inactivityTimer.Reset(inactivityTimeout)
					e.SetDoorOpenLamp(false)
					e.SetMotorDirection(e_state.MotorDirection)
					doorTimer.Stop()
				}
			}
			//TODO: UpdateFunctionalStatus(commands)

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
		case DataToFSM <- BuildElevatorData(localID, e.GetState(commands)):
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
func BuildElevatorData(localID int, e_state e.ElevatorState_t) messageSync.ElevatorData_t {

}

//TODO: Helper function to convert RA_Output to MotorDirectionBehaviourPair using request assigner output
func RAOutputToMotorDirectionPair(raOutput ra.RA_Output, localID int, e_state e.ElevatorState_t) e.MotorDirectionBehaviourPair_t {

}
