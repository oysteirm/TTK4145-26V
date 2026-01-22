package fsm

import (
	"elevator"
	"fmt"
	elevator "project"
	"project/elevio/elevio"
)

//turns on all button lights
func set_all_lights(e_state elevator.ElevatorBehaviour){ 
	for floor := 0; floor < N_FLOORS; floor++{
		for btn := 0; btn < N_BUTTONS; btn++{
            elevio.SetButtonLamp(btn, floor, e_state.Requests[floor][btn])
		}
	}
}

//elevator moves down on init between floors
func OnInitBetweenFloors(commands chan elevator.Command){
	elevator_motor_direction(MD_Down)

	// e.dirn = D_Down //use server!
	// e.behavoiur = EB_Moving //use server!
    commands <- elevator.SetMotorDirection{MotorDirection: elevio.MD_Down}
    commands <- elevator.SetElevatorBehavior{ElevatorBehaviour: elevator.EB_Moving}
}


//what to do if there is a button press
func OnRequestButtonPress(commands chan elevator.Command, btn_floor int, btn_type elevio.ButtonType){
    e_state = elevator.GetState(commands)
	fmt.printf("\n\n%s(%d, %s)\n", __FUNCTION__, btn_floor, elevator_button_to_string(btn_type))
	elevator_print(e_state)

	switch(e_state.ElevatorBehaviour){
    case EB_DoorOpen:
        if(requests_should_clear_immediately(commands, btn_floor, btn_type)){
            timer_start(e_state.DoorOpenDuration);
        } else {
            // e.requests[btn_floor][btn_type] = 1; // use server!
            commands <- elevator.SetRequests{RequestValue: 1, Floor: btn_floor, Button: btn_type}
        }
        break;

    case EB_Moving:
        // e.requests[btn_floor][btn_type] = 1; //use server!
        commands <- elevator.SetRequests{RequestValue: 1, Floor: btn_floor, Button: btn_type}
        break;
        
    case EB_Idle:    
        //e.requests[btn_floor][btn_type] = 1; //use server
        commands <- elevator.SetRequests{RequestValue: 1, Floor: btn_floor, Button: btn_type}
        pair DirnBehaviourPair = requests_choose_direction(e);
        // e->dirn = pair.dirn; //use server!
        // e->behaviour = pair.behaviour; //use server!
        commands <- elevator.SetMotorDirection{MotorDiretion: pair.dirn}
        commands <- elevator.SetElevatorBehaviour{EleElevatorBehaviour: pair.behavoior}
        switch(pair.behaviour){
        case EB_Door_Open:
            elevio.SetDoorOpenLamp(true)
            timer_start(e.DoorOpenDuration);
            e_state = requests_clear_at_current_floor(e_state); //need to be changed
            break;

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
func OnFloorArrival(commands chan elevator.Command, newFloor int) {
    // Update floor
    commands <- elevator.SetFloor{Floor: newFloor}

    e_state := elevator.GetStateSafe(commands)

    elevio.SetFloorIndicator(newFloor)

    if e_state.ElevatorBehaviour == elevator.EB_Moving {
        if requests_should_stop(e_state) {

            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)

            newRequests := requests_clear_at_current_floor(e_state) //needs ti be changed
            commands <- elevator.SetRequests{Requests: newRequests} // this also

            commands <- elevator.SetElevatorBehavior{
                ElevatorBehaviour: elevator.EB_DoorOpen,
            }

            timer_start(state.DoorOpenDuration)
        }
    }
}


//what to do if the door timer runs out
func OnDoorTimeout(commands chan elevator.Command){
    e_state := elevator.GetState(commands)

    switch(e_state.ElevatorBehaviour){
    case elevator.EB_Door_Open:;
        pair DirnBehaviourPair= requests_choose_direction(e_state);
        // e.dirn = pair.dirn; //use server!
        // e.behaviour = pair.behaviour; //use server!
		commands <- elevator.SetMotorDirection{MotorDiretion: pair.dirn}
        commands <- elevator.SetElevatorBehaviour{EleElevatorBehaviour: pair.behavoior}
		
        switch(pair.behavoior){
        case elevator.EB_Door_Open:
            timer_start(state.DoorOpenDuration);
            e_state = requests_clear_at_current_floor(e_state);
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