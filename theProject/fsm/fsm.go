package fsm

import (
	"fmt"
	"theProject/elevator_IO"
	"theProject/elevatorStateGuardian"
	"theProject/messageSync"
    "theProject/requests"
	"time"
)

//Light functions using cyclic counter values
func LightCabLights(CabRequests []messageSync.RequestCyclicCounter_t) {

	for floor := 0; floor < elevator_IO.N_FLOORS; floor++{
		elevator_IO.SetButtonLamp(elevator_IO.BT_Cab, floor, CC_ToBool(CabRequests[floor].Value))
	}
}
func LightHallLights(Hall_Requests [][2]messageSync.RequestCyclicCounter_t) {
	for floor := 0; floor < elevator_IO.N_FLOORS; floor++{
		elevator_IO.SetButtonLamp(elevator_IO.BT_HallUp, floor, CC_ToBool(Hall_Requests[floor][elevator_IO.BT_HallUp].Value))
		elevator_IO.SetButtonLamp(elevator_IO.BT_HallDown, floor, CC_ToBool(Hall_Requests[floor][elevator_IO.BT_HallDown].Value))
	}
}
//maybe move this?
func CC_ToBool(CC messageSync.CyclicCounter_t) bool {
	if (CC == messageSync.CC_Uninit || CC == messageSync.CC_No || CC == messageSync.CC_Unconfirmed) {
		return false
	}
	if CC == messageSync.CC_Confirmed || CC == messageSync.CC_Done {
		return true
	} else {
		print("wrong CC Value")
		return false
	}
}

//elevator moves down on init between floors
func OnInitBetweenFloors(guardianCommands chan elevatorStateGuardian.GuardianCommands_t){
	elevator_IO.SetMotorDirection(elevator_IO.MD_Down)

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    //Moving down
    elevatorState.ElevatorBehaviour = elevator_IO.EB_Moving
    elevatorState.MotorDirection    = elevator_IO.MD_Down
    //Save in guardian
    guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
}


//what to do if we recieve new data
func OnReceivedDataFromMsgSync(
    guardianCommands chan elevatorStateGuardian.GuardianCommands_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}){

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)

    //TODO: how we want to print? need to decide
	fmt.Printf("\n\n%s(%d, %s)\n", "OnReceivedDataFromMsgSync")
	ElevatorPrint(elevatorState)

    
    var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests);

    elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
    elevatorState.MotorDirection    = pair.MotorDirection
    //Save in guardian
    guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

    switch(pair.ElevatorBehaviour){
    case elevator_IO.EB_DoorOpen:
        //stop inactive timer
        elevator_IO.SetDoorOpenLamp(true)
        //TODO: use a const
        doorTimerStart <- elevatorState.DoorOpenDuration

        //change RequestsClearAtCurrentFloor return cleared request (in floor)
        requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests);
        //Set a cyclic counter to done channel...
        // mark if cab or hall
        guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

    case elevator_IO.EB_Moving:
        //start inactive timer
        elevator_IO.SetMotorDirection((pair.MotorDirection))
        break;

    case elevator_IO.EB_Idle:
        //stop inactive timer
        break;
    
    }
    //think about this
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(elevatorState);
}


//what to do if we arrive at a floor
func OnFloorArrival(
    guardianCommands chan elevatorStateGuardian.GuardianCommands_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}, 
    isFunctionalStart chan struct{}, 
    isFunctionalStop chan struct{}, 
    newFloor int) {

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)
    //update floor
    elevatorState.Floor = newFloor
    elevator_IO.SetFloorIndicator(newFloor)

    //change
    isFunctionalStop <- struct{}{}

    if elevatorState.ElevatorBehaviour == elevator_IO.EB_Moving {
        if requests.RequestsShouldStop(elevatorState, assignedRequests) {

            elevator_IO.SetMotorDirection(elevator_IO.MD_Stop)
            elevator_IO.SetDoorOpenLamp(true)

            elevatorState.MotorDirection = elevator_IO.MD_Stop
            elevatorState.ElevatorBehaviour = elevator_IO.EB_DoorOpen
            
            requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests) 
            guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

            guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

            //Reset doorTimer
            doorTimerStop <- struct{}{}
            doorTimerStart <- elevatorState.DoorOpenDuration //use CONST!
        }
    }
    //change also this to what we want
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(elevatorState);
}


//what to do if the door timer runs out
func OnDoorTimeout(
    guardianCommands chan elevatorStateGuardian.GuardianCommands_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}, 
    isFunctionalStart chan struct{}, 
    isFunctionalStop chan struct{}){

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)

    switch(elevatorState.ElevatorBehaviour){

    case elevator_IO.EB_DoorOpen:
        var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests);

        elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
        elevatorState.MotorDirection    = pair.MotorDirection
        //Save in guardian
        guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

        switch(elevatorState.ElevatorBehaviour){
        case elevator_IO.EB_DoorOpen:
            doorTimerStop <- struct{}{}
            doorTimerStart <- elevatorState.DoorOpenDuration //use CONST

            requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests) 
            guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

            break;

        //add functional timer to the cases under
        case elevator_IO.EB_Moving:
            //start functional timer
            elevator_IO.SetDoorOpenLamp(false)
            elevator_IO.SetMotorDirection(pair.MotorDirection);

        case elevator_IO.EB_Idle:
            //stop functional timer
            elevator_IO.SetDoorOpenLamp(false)
            elevator_IO.SetMotorDirection(pair.MotorDirection);
            break;
        }
        
        break;
    default:
        break;
    }
    //this also
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(elevatorState);
}