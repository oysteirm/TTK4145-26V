package fsm

import "fmt"

//turns on all button lights
func set_all_lights(e Elevator){ 
	for floor := 0; floor < N_FLOORS; floor++{
		for btn := 0; btn < N_BUTTONS; btn++{
			elevator_request_button_light(floor, btn, e.requests[floor][btn])
		}
	}
}

//elevator moves down on init between floors
func fsm_on_init_between_floors(e *Elevator){
	elevator_motor_direction(D_Down)
	e.dirn = D_Down //use server!
	e.behavoiur = EB_Moving //use server!
}


//what to do if there is a button press
func fsm_on_request_button_press(e *Elevator, btn_floor int, btn_type Button){
	fmt.printf("\n\n%s(%d, %s)\n", __FUNCTION__, btn_floor, elevator_button_to_string(btn_type))
	elevator_print(*e)

	switch(e.behaviour){
    case EB_DoorOpen:
        if(requests_should_clear_immediately(*e, btn_floor, btn_type)){
            timer_start(e.config.doorOpenDuration_s);
        } else {
            e.requests[btn_floor][btn_type] = 1; // use server!
        }
        break;

    case EB_Moving:
        e.requests[btn_floor][btn_type] = 1; //use server!
        break;
        
    case EB_Idle:    
        e.requests[btn_floor][btn_type] = 1; //use server
        pair DirnBehaviourPair = requests_choose_direction(*e);
        e->dirn = pair.dirn; //use server!
        e->behaviour = pair.behaviour; //use server!
        switch(pair.behaviour){
        case EB_Door_Open:
            elevator_door_light(1);
            timer_start(e->config.doorOpenDuration_s);
            *e = requests_clear_at_current_floor(*e);
            break;

        case EB_Moving:
            elevator_motor_direction(e->dirn);
            break;
            
        case EB_Idle:
            break;
        }
        break;
    }
    
    set_all_lights(*e);
    
    fmt.printf("\nNew state:\n");
    elevator_print(*e);
}


//what to do if we arrive at a floor
func OnFloorArrival(commands chan elevator.Command, newFloor int) {
    state := elevator.GetStateSafe(commands)

    // Update floor
    commands <- elevator.SetFloor{Floor: newFloor}
    elevio.SetFloorIndicator(newFloor)

    if state.ElevatorBehaviour == elevator.EB_Moving {
        if requests_should_stop(state) {

            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)

            newRequests := requests_clear_at_current_floor(state)
            commands <- elevator.SetRequests{Requests: newRequests}

            commands <- elevator.SetElevatorBehavior{
                Behaviour: elevator.EB_DoorOpen,
            }

            timer_start(state.DoorOpenDuration)
        }
    }
}


//what to do if the door timer runs out
func OnDoorTimeout(commands chan elevator.Command){
    state := elevator.GetState(commands)

    switch(state.behaviour){
    case elevator.EB_Door_Open:;
        pair DirnBehaviourPair= requests_choose_direction(state);
        // e.dirn = pair.dirn; //use server!
        // e.behaviour = pair.behaviour; //use server!
		commands <- elevator.SetMotorDirection{MotorDiretion: pair.dirn}
        commands <- elevator.SetElevatorBehaviour{EleElevatorBehaviour: pair.behavoior}
		
        switch(state.behaviour){
        case elevator.EB_Door_Open:
            timer_start(state.DoorOpenDuration);
            *e = requests_clear_at_current_floor(*e);
            set_all_lights(state);
            break;
        case elevator.EB_Moving:
        case elevator.EB_Idle:
            elevator_door_light(0);
            elevator_motor_direction(state.dirn);
            break;
        }
        
        break;
    default:
        break;
    }
    
    // fmt.printf("\nNew state:\n");
    // elevator_print(*e);
}