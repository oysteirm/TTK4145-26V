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
// merge into right case
func OnInitBetweenFloors(commands chan Command_t){
	elevatorIo.SetMotorDirection(elevatorIo.MD_Down)
    commands <- SetMotorDirection_t{MotorDirection: elevatorIo.MD_Down}
    commands <- SetElevatorBehaviour_t{ElevatorBehaviour: elevatorIo.EB_Moving}
}


//what to do if there is a button press
// name change (onRecievdDataFromsMsgSync)
func OnReviecedDataFromMsgSync(
    commands chan Command_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}, 
    btnFloor int, 
    btnType elevatorIo.ButtonType_t){

    var e_state ElevatorState_t = GetState(commands)

	fmt.Printf("\n\n%s(%d, %s)\n", "OnRequestButtonPress",btnFloor, ElevatorButtonToString(btnType))
	ElevatorPrint(e_state)

    //add assigned requests to RequestsCD function
    var pair MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(e_state );
    //put into one (under)
    commands <- SetMotorDirection_t{MotorDirection: pair.MotorDirection}
    commands <- SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}
    switch(pair.ElevatorBehaviour){
    case elevatorIp.EB_DoorOpen:
        elevatorIo.SetDoorOpenLamp(true)
        
        doorTimerStart <- e_state.DoorOpenDuration
        //change RequestsClearAtCurrentFloor return cleared request (in floor)
        e_state = requests.RequestsClearAtCurrentFloor(e_state);
        //Set a cyclic counter to done channel...
        // mark if cab or hall
        commands <- SetState_t{ElevatorState: e_state}

    case elevatorIo.EB_Moving:
        elevatorIo.SetMotorDirection((pair.MotorDirection))
        break;

    case elevatorIo.EB_Idle:
        break;
    
    }

    fmt.Printf("\nNew state:\n");
    ElevatorPrint(e_state);
}


//what to do if we arrive at a floor
func OnFloorArrival(
    commands chan Command_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}, 
    inactiveStart chan struct{}, 
    inactiveStop chan struct{}, 
    newFloor int) {

    // Update floor
    commands <- SetFloor_t{Floor: newFloor}
    //TODO: lastFloorTime = time.Now()  // Update time when reaching floor

    var e_state ElevatorState_t = GetState(commands)

    elevatorIo.SetFloorIndicator(newFloor)

    inactiveStop <- struct{}{}

    if e_state.ElevatorBehaviour == elevatorIo.EB_Moving {
        if requests.RequestsShouldStop(e_state) {

            elevatorIo.SetMotorDirection(elevatorIo.MD_Stop)
            elevatorIo.SetDoorOpenLamp(true)
//change RequestsClearAtCurrentFloor return cleared request (in floor)
            e_state = requests.RequestsClearAtCurrentFloor(e_state) 
            //Set a cyclic counter to done channel...
              // mark if cab or hall
            commands <- SetState_t{ElevatorState: e_state}
            //Use only one
            commands <- SetMotorDirection_t{MotorDirection: elevatorIo.MD_Stop}
            commands <- SetElevatorBehaviour_t{ElevatorBehaviour: elevatorIo.EB_DoorOpen}

            //Reset doorTimer
            doorTimerStop <- struct{}{}
            doorTimerStart <- e_state.DoorOpenDuration
        }
    }
    e_state = GetState(commands)
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(e_state);
}


//what to do if the door timer runs out
func OnDoorTimeout(
    commands chan Command_t, 
    doorTimerStart chan time.Duration, 
    doorTimerStop chan struct{}){

    var e_state ElevatorState_t = GetState(commands)

    switch(e_state.ElevatorBehaviour){

    case elevatorIo.EB_DoorOpen:
        var pair MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(e_state);
        //Use only one
        commands <- SetMotorDirection_t{MotorDirection: pair.MotorDirection}
        commands <- SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}
        e_state = GetState(commands)

        switch(e_state.ElevatorBehaviour){
        case elevatorIo.EB_DoorOpen:
            doorTimerStop <- struct{}{}
            doorTimerStart <- e_state.DoorOpenDuration
            //change RequestsClearAtCurrentFloor return cleared request (in floor)
            e_state = requests.RequestsClearAtCurrentFloor(e_state);
            // mark if cab or hall
            commands <- SetState_t{ElevatorState: e_state}

            break;

        case elevatorIo.EB_Moving, elevatorIo.EB_Idle:
            elevatorIo.SetDoorOpenLamp(false)
            elevatorIo.SetMotorDirection(pair.MotorDirection);
            break;
        }
        
        break;
    default:
        break;
    }
    e_state = GetState(commands)
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(e_state);
}