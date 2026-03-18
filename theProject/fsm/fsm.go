package fsm

// This code is inspired by provided code fetched from: "https://github.com/TTK4145/Project-resources/tree/master/elev_algo"

import (
	"fmt"
	"theProject/config"
	"theProject/converters"
	"theProject/elevatorStateGuardian"
	"theProject/elevator_IO"
	"theProject/messageSync"
	"theProject/requests"
)

/*
-----------------------------------
Functionality: 
	- Event based finite state machine (FSM) for local elevator
	- Changing states on events and setting it in elevatorStateGuardian
	- Using elevator_IO functions for controlling hardware (motor and lights)
-----------------------------------
*/

func LightCabLights(CabRequests [config.N_FLOORS]messageSync.RequestCyclicCounter_t) {

	for floor := 0; floor < config.N_FLOORS; floor++ {
		elevator_IO.SetButtonLamp(elevator_IO.BT_Cab, floor, converters.CC_ToBool(CabRequests[floor].Value))
	}
}

func LightHallLights(Hall_Requests [config.N_FLOORS][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		elevator_IO.SetButtonLamp(elevator_IO.BT_HallUp, floor, converters.CC_ToBool(Hall_Requests[floor][elevator_IO.BT_HallUp].Value))
		elevator_IO.SetButtonLamp(elevator_IO.BT_HallDown, floor, converters.CC_ToBool(Hall_Requests[floor][elevator_IO.BT_HallDown].Value))
	}
}

// Elevator moves down until hitting a floor
func OnInitBetweenFloors(guardianCommands chan elevatorStateGuardian.GuardianCommands_t, drv_floors chan int) {
	
	elevator_IO.SetDoorOpenLamp(false)
	elevator_IO.SetMotorDirection(elevator_IO.MD_Down)

	elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)

	for {
		floor := <-drv_floors
		if floor != -1 {
			elevator_IO.SetMotorDirection(elevator_IO.MD_Stop)
			elevatorState.Floor = floor
			break
		}
	}

	elevatorState.ElevatorBehaviour = elevator_IO.EB_Idle
	elevatorState.MotorDirection = elevator_IO.MD_Stop

	guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
}

func OnReceivedDataFromMsgSync(
	guardianCommands chan elevatorStateGuardian.GuardianCommands_t,
	doorTimerStart chan struct{},
	doorTimerStop chan struct{},
	isFunctionalStart chan struct{},
	isFunctionalStop chan struct{}) {

	elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
	assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)

	// To avoid breaking out of a door open sequence
	if elevatorState.ElevatorBehaviour != elevator_IO.EB_DoorOpen {

		var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests)

		elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
		elevatorState.MotorDirection = pair.MotorDirection
		
		guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
	}

	switch elevatorState.ElevatorBehaviour {
	case elevator_IO.EB_DoorOpen:
		isFunctionalStop <- struct{}{}
		doorTimerStart <- struct{}{}
		elevator_IO.SetDoorOpenLamp(true)

		requestsToClear := requests.RequestsClearOnNewData(elevatorState, assignedRequests)
		guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

	case elevator_IO.EB_Moving:
		isFunctionalStart <- struct{}{}
		elevator_IO.SetMotorDirection((elevatorState.MotorDirection))
		elevator_IO.SetDoorOpenLamp(false)

	case elevator_IO.EB_Idle:
		isFunctionalStop <- struct{}{}
		elevator_IO.SetMotorDirection((elevatorState.MotorDirection))
		elevator_IO.SetDoorOpenLamp(false)
	}
}

func OnFloorArrival(
	guardianCommands chan elevatorStateGuardian.GuardianCommands_t,
	doorTimerStart chan struct{},
	doorTimerStop chan struct{},
	isFunctionalStart chan struct{},
	isFunctionalStop chan struct{},
	newFloor int, 
	isObstructed bool) {

	elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
	assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)
	
	// Update floor and IsFunctional
	if !isObstructed {
		elevatorState.IsFunctional = true
	}

	elevatorState.Floor = newFloor
	elevator_IO.SetFloorIndicator(newFloor)

	if elevatorState.ElevatorBehaviour == elevator_IO.EB_Moving {

		// Not being able to go through the floor or the roof
		if (elevatorState.Floor == 0 && elevatorState.MotorDirection == elevator_IO.MD_Down) ||
	 		(elevatorState.Floor == (config.N_FLOORS-1) && elevatorState.MotorDirection == elevator_IO.MD_Up){
			elevator_IO.SetMotorDirection(elevator_IO.MD_Stop)
		}

		if requests.RequestsShouldStop(elevatorState, assignedRequests) {

			elevator_IO.SetMotorDirection(elevator_IO.MD_Stop)
			elevator_IO.SetDoorOpenLamp(true)

			elevatorState.ElevatorBehaviour = elevator_IO.EB_DoorOpen

			requestsToClear := requests.RequestsClearOnFloorArrival(elevatorState, assignedRequests)
			guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

			// Resetting door timer
			doorTimerStop <- struct{}{}
			doorTimerStart <- struct{}{}

			isFunctionalStop <- struct{}{}
		}

		guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

		if elevatorState.ElevatorBehaviour == elevator_IO.EB_Moving {
			// Resetting isFunctional timer
			isFunctionalStop <- struct{}{}
			isFunctionalStart <- struct{}{}
		} else {
			isFunctionalStop <- struct{}{}
		}
	}
	assignedRequests = elevatorStateGuardian.GetAssignedRequests(guardianCommands)
	fmt.Printf("\nNew state from FloorArrival:\n")
	ElevatorPrint(elevatorState, assignedRequests)
}

func OnDoorTimeout(
	guardianCommands chan elevatorStateGuardian.GuardianCommands_t,
	doorTimerStart chan struct{},
	doorTimerStop chan struct{},
	isFunctionalStart chan struct{},
	isFunctionalStop chan struct{},
	isObstructed bool) {

	elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
	assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)


	switch elevatorState.ElevatorBehaviour {
	case elevator_IO.EB_DoorOpen:

		if !isObstructed {
			var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests)

			elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
			elevatorState.MotorDirection = pair.MotorDirection
			
			guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
		}

		switch elevatorState.ElevatorBehaviour {
		case elevator_IO.EB_DoorOpen:
			// Resetting door timer
			doorTimerStop <- struct{}{}
			doorTimerStart <- struct{}{}

			requestsToClear := requests.RequestsClearOnDoorTimeout(elevatorState, assignedRequests)
			guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

		case elevator_IO.EB_Moving:

			isFunctionalStart <- struct{}{}
			elevator_IO.SetDoorOpenLamp(false)
			elevator_IO.SetMotorDirection(elevatorState.MotorDirection)

		case elevator_IO.EB_Idle:

			isFunctionalStop <- struct{}{}
			elevator_IO.SetDoorOpenLamp(false)
			elevator_IO.SetMotorDirection(elevatorState.MotorDirection)
		}
	default:
		break
	}
	assignedRequests = elevatorStateGuardian.GetAssignedRequests(guardianCommands)
	fmt.Printf("\nNew state from DoorTimeout:\n")
	ElevatorPrint(elevatorState, assignedRequests)
}

func ElevatorPrint(
	elevator messageSync.ElevatorData_t, 
	assignedRequests elevator_IO.AssignedRequests_t) {

	fmt.Printf("  +--------------------+\n")
	fmt.Printf(
		"  |IsAlive = %-9t |\n"+
			"  |IsFunctional = %-2t |\n"+
			"  |floor = %-11d |\n"+
			"  |dirn  = %-12s|\n"+
			"  |behav = %-12s|\n",
		elevator.IsAlive,
		elevator.IsFunctional,
		elevator.Floor,
		converters.ElevatorDirnToString(elevator.MotorDirection),
		converters.ElevatorBehaviourToString(elevator.ElevatorBehaviour),
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	for f := config.N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)

		for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.ButtonType_t(config.N_BUTTONS); btn++ {
			if (f == config.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
				(f == 0 && btn == elevator_IO.BT_HallDown) {

				fmt.Print("|     ")
			} else {
				if btn == elevator_IO.BT_Cab {
					if converters.CC_ToBool(elevator.CabRequests[f].Value) {
						fmt.Print("|  #  ")
					} else {
						fmt.Print("|  -  ")
					}
				} else {
					if assignedRequests[f][btn] {
						fmt.Print("|  #  ")
					} else {
						fmt.Print("|  -  ")
					}
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}