package elevator

import (
	"fmt"
)

//turns on all button lights
func set_all_lights(e_state ElevatorBehaviour_t){ 
	for floor := 0; floor < N_FLOORS; floor++{
		for btn := 0; btn < N_BUTTONS; btn++{
            elevio.SetButtonLamp(btn, floor, e_state.Requests[floor][btn])
		}
	}
}

//elevator moves down on init between floors
func OnInitBetweenFloors(commands chan Command_t){
	elevator_motor_direction(MD_Down)
    commands <- SetMotorDirection_t{MotorDirection: elevio.MD_Down}
    commands <- SetElevatorBehavior_t{ElevatorBehaviour: elevator.EB_Moving}
}


//what to do if there is a button press
func OnRequestButtonPress(commands chan Command_t, btn_floor int, btn_type ButtonType_t){
    e_state = elevator.GetState(commands)
	fmt.printf("\n\n%s(%d, %s)\n", __FUNCTION__, btn_floor, elevator_button_to_string(btn_type))
	elevator_print(e_state)

	switch(e_state.ElevatorBehaviour){
    case EB_DoorOpen:
        if(requests_should_clear_immediately(commands, btn_floor, btn_type)){
            timer_start(e_state.DoorOpenDuration);
        } else {
            commands <- SetRequest_t{RequestValue: 1, Floor: btn_floor, Button: btn_type}
        }
        break;

    case EB_Moving:
        commands <- SetRequest_t{RequestValue: 1, Floor: btn_floor, Button: btn_type}
        break;
        
    case EB_Idle:    
        commands <- SetRequest_t{RequestValue: 1, Floor: btn_floor, Button: btn_type}
        pair MotorDirectionBehaviourPair_t = requests_choose_direction(e_state);
        commands <- SetMotorDirection_t{MotorDiretion: pair.Dirn}
        commands <- SetElevatorBehaviour_t{ElevatorBehaviour: pair.ElevatorBehaviour}
        switch(pair.ElevatorBehaviour){
        case EB_Door_Open:
            elevio.SetDoorOpenLamp(true)
            timer_start(e_state.DoorOpenDuration);
            e_state = requests_clear_at_current_floor(e_state);
            commands <- SetState_t{ElevatorState: e_state}

        case EB_Moving:
            elevio.SetMotorDirection((e_state.MotorDirection))
            break;
            
        case EB_Idle:
            break;
        }
        break;
    }
    
    set_all_lights(e_state);
    
    fmt.printf("\nNew state:\n");
    elevator_print(e_state);
}


//what to do if we arrive at a floor
func OnFloorArrival(commands chan Command_t, newFloor int) {
    // Update floor
    commands <- SetFloor_t{Floor: newFloor}

    e_state := elevator.GetStateSafe(commands)

    elevio.SetFloorIndicator(newFloor)

    if e_state.ElevatorBehaviour == elevator.EB_Moving {
        if requests_should_stop(e_state) {

            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)

            e_state := requests_clear_at_current_floor(e_state) 
            commands <- SetState_t{ElevatorState: e_state}

            commands <- SetElevatorBehavior_t{ElevatorBehaviour: elevator.EB_DoorOpen}

            timer_start(state.DoorOpenDuration)
        }
    }
}


//what to do if the door timer runs out
func OnDoorTimeout(commands chan Command_t){
    e_state := elevator.GetState(commands)

    switch(e_state.ElevatorBehaviour){
    case elevator.EB_Door_Open:;
        pair DirnBehaviourPair= requests_choose_direction(e_state);
		commands <- SetMotorDirection_t{MotorDiretion: pair.dirn}
        commands <- SetElevatorBehaviour_t{EleElevatorBehaviour: pair.behavoior}
		
        switch(pair.behavoior){
        case elevator.EB_Door_Open:
            timer_start(state.DoorOpenDuration);
            e_state = requests_clear_at_current_floor(e_state);
            commands <- SetState_t{ElevatorState: e_state}
            set_all_lights(e_state);
            break;
        case elevator.EB_Moving:
        case elevator.EB_Idle:
            elevator_door_light(0);
            elevator_motor_direction(e_state.MotorDirection);
            break;
        }
        
        break;
    default:
        break;
    }
    
    fmt.printf("\nNew state:\n");
    elevator_print(e_state);
}