package elevator

import (
	"fmt"
    "time"
)

//turns on all button lights
func SetAllLights(e_state ElevatorState_t){ 
	for floor := 0; floor < N_FLOORS; floor++{
		for btn := ButtonType_t(0); btn < N_BUTTONS; btn++{
            SetButtonLamp(btn, floor, e_state.Requests[floor][btn])
		}
	}
}

//elevator moves down on init between floors
func OnInitBetweenFloors(commands chan Command_t){
	SetMotorDirection(MD_Down)
    commands <- SetMotorDirection_t{MotorDirection: MD_Down}
    commands <- SetElevatorBehaviour_t{ElevatorBehaviour: EB_Moving}
}


//what to do if there is a button press
func OnRequestButtonPress(commands chan Command_t, doorTimerStart chan time.Duration, doorTimerStop chan struct{}, btnFloor int, btnType ButtonType_t){
    var e_state ElevatorState_t = GetState(commands)

	fmt.Printf("\n\n%s(%d, %s)\n", "OnRequestButtonPress",btnFloor, ElevatorButtonToString(btnType))
	ElevatorPrint(e_state)

	switch(e_state.ElevatorBehaviour){
    case EB_DoorOpen:
        if(RequestsShouldClearImmediately(e_state, btnFloor, btnType)){
            doorTimerStop <- struct{}{} 
            doorTimerStart <- e_state.DoorOpenDuration
            SetDoorOpenLamp(true)
        } else {
            commands <- SetRequest_t{RequestValue: true, Floor: btnFloor, Button: btnType}
        }
        break;

    case EB_Moving:
        commands <- SetRequest_t{RequestValue: true, Floor: btnFloor, Button: btnType}
        break;
        
    case EB_Idle:    
        commands <- SetRequest_t{RequestValue: true, Floor: btnFloor, Button: btnType}
        e_state = GetState(commands)
        var pair MotorDirectionBehaviourPair_t = RequestsChooseDirection(e_state);
        commands <- SetMotorDirection_t{MotorDirection: pair.MotorDirection}
        commands <- SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}
        switch(pair.ElevatorBehaviour){
        case EB_DoorOpen:
            SetDoorOpenLamp(true)
            
            doorTimerStart <- e_state.DoorOpenDuration
            e_state = RequestsClearAtCurrentFloor(e_state);
            commands <- SetState_t{ElevatorState: e_state}

        case EB_Moving:
            SetMotorDirection((pair.MotorDirection))
            break;
            
        case EB_Idle:
            break;
        }
        break;
    }
    e_state = GetState(commands)
    SetAllLights(e_state);
    
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(e_state);
}


//what to do if we arrive at a floor
func OnFloorArrival(commands chan Command_t, doorTimerStart chan time.Duration, doorTimerStop chan struct{}, resetInactive chan struct{}, newFloor int) {
    // Update floor
    commands <- SetFloor_t{Floor: newFloor}
    //TODO: lastFloorTime = time.Now()  // Update time when reaching floor

    var e_state ElevatorState_t = GetState(commands)

    SetFloorIndicator(newFloor)

    resetInactive <- struct{}{}

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
        }
    }
    e_state = GetState(commands)
    fmt.Printf("\nNew state:\n");
    ElevatorPrint(e_state);
}


//what to do if the door timer runs out
func OnDoorTimeout(commands chan Command_t, doorTimerStart chan time.Duration, doorTimerStop chan struct{}){
    var e_state ElevatorState_t = GetState(commands)

    switch(e_state.ElevatorBehaviour){
    case EB_DoorOpen:
        var pair MotorDirectionBehaviourPair_t = RequestsChooseDirection(e_state);
        commands <- SetMotorDirection_t{MotorDirection: pair.MotorDirection}
        commands <- SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}
        e_state = GetState(commands)
        switch(e_state.ElevatorBehaviour){
        case EB_DoorOpen:
            doorTimerStop <- struct{}{}
            doorTimerStart <- e_state.DoorOpenDuration
            e_state = RequestsClearAtCurrentFloor(e_state);
            commands <- SetState_t{ElevatorState: e_state}
            SetAllLights(e_state);
            break;
        case EB_Moving, EB_Idle:
            SetDoorOpenLamp(false)
            SetMotorDirection(pair.MotorDirection);
            SetAllLights(e_state);
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