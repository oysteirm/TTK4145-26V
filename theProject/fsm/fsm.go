package fsm

import (
	"fmt"
	"theProject/elevatorIo"
	"theProject/elevatorStateGuardian"
	"theProject/messageSync"
    "theProject/requests"
	"time"
)

//Light functions using cyclic counter values
func lightCabLights(CabRequests []messageSync.RequestCyclicCounter_t) {

	for floor := 0; floor < elevatorIo.N_FLOORS; floor++{
		elevatorIo.SetButtonLamp(elevatorIo.BT_Cab, floor, CC_ToBool(CabRequests[floor].Value))
	}
}
func lightHallLights(Hall_Requests [][2]messageSync.RequestCyclicCounter_t) {
	for floor := 0; floor < elevatorIo.N_FLOORS; floor++{
		elevatorIo.SetButtonLamp(elevatorIo.BT_HallUp, floor, CC_ToBool(Hall_Requests[floor][elevatorIo.BT_HallUp].Value))
		elevatorIo.SetButtonLamp(elevatorIo.BT_HallDown, floor, CC_ToBool(Hall_Requests[floor][elevatorIo.BT_HallDown].Value))
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
func OnInitBetweenFloors(commands chan elevatorStateGuardian.Command_t){
	elevatorIo.SetMotorDirection(elevatorIo.MD_Down)

    elevatorState := elevatorStateGuardian.GetElevatorData(commands)
    //Moving down
    elevatorState.ElevatorBehaviour = elevatorIo.EB_Moving
    elevatorState.MotorDirection    = elevatorIo.MD_Down
    //Save in guardian
    commands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
}


//what to do if we recieve new data
func OnReceivedDataFromMsgSync(
    commands chan elevatorStateGuardian.Command_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}, 
    btnFloor int, 
    btnType elevatorIo.ButtonType_t){

    elevatorState := elevatorStateGuardian.GetElevatorData(commands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(commands)

    //TODO: how we want to print? need to decide
	fmt.Printf("\n\n%s(%d, %s)\n", "OnRequestButtonPress",btnFloor, elevatorStateGuardian.ElevatorButtonToString(btnType))
	ElevatorPrint(elevatorState)

    
    var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests);

    elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
    elevatorState.MotorDirection    = pair.MotorDirection
    //Save in guardian
    commands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

    switch(pair.ElevatorBehaviour){
    case elevatorIo.EB_DoorOpen:
        //stop inactive timer
        elevatorIo.SetDoorOpenLamp(true)
        //TODO: use a const
        doorTimerStart <- elevatorState.DoorOpenDuration

        //change RequestsClearAtCurrentFloor return cleared request (in floor)
        requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests);
        //Set a cyclic counter to done channel...
        // mark if cab or hall
        commands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

    case elevatorIo.EB_Moving:
        //start inactive timer
        elevatorIo.SetMotorDirection((pair.MotorDirection))
        break;

    case elevatorIo.EB_Idle:
        //stop inactive timer
        break;
    
    }
    //think about this
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(elevatorState);
}


//what to do if we arrive at a floor
func OnFloorArrival(
    commands chan elevatorStateGuardian.Command_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}, 
    inactiveStart chan struct{}, 
    inactiveStop chan struct{}, 
    newFloor int) {

    elevatorState := elevatorStateGuardian.GetElevatorData(commands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(commands)
    //update floor
    elevatorState.Floor = newFloor
    elevatorIo.SetFloorIndicator(newFloor)

    //change
    inactiveStop <- struct{}{}

    if elevatorState.ElevatorBehaviour == elevatorIo.EB_Moving {
        if requests.RequestsShouldStop(elevatorState, assignedRequests) {

            elevatorIo.SetMotorDirection(elevatorIo.MD_Stop)
            elevatorIo.SetDoorOpenLamp(true)

            elevatorState.MotorDirection = elevatorIo.MD_Stop
            elevatorState.ElevatorBehaviour = elevatorIo.EB_DoorOpen
            
            requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests) 
            commands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

            commands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

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
    commands chan elevatorStateGuardian.Command_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}){

    elevatorState := elevatorStateGuardian.GetElevatorData(commands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(commands)

    switch(elevatorState.ElevatorBehaviour){

    case elevatorIo.EB_DoorOpen:
        var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests);

        elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
        elevatorState.MotorDirection    = pair.MotorDirection
        //Save in guardian
        commands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}

        switch(elevatorState.ElevatorBehaviour){
        case elevatorIo.EB_DoorOpen:
            doorTimerStop <- struct{}{}
            doorTimerStart <- elevatorState.DoorOpenDuration //use CONST

            requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests) 
            commands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

            break;

        //add functional timer to the cases under
        case elevatorIo.EB_Moving:
            //start functional timer
            elevatorIo.SetDoorOpenLamp(false)
            elevatorIo.SetMotorDirection(pair.MotorDirection);

        case elevatorIo.EB_Idle:
            //stop functional timer
            elevatorIo.SetDoorOpenLamp(false)
            elevatorIo.SetMotorDirection(pair.MotorDirection);
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