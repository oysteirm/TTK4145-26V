package main

import (
	"project/elevator"
	//"fmt"
	"time"
)

func main(){

    //var N_ELEVATORS int = 1

    elevator.Init("localhost:15657", elevator.N_FLOORS)

    commands := make(chan elevator.Command_t)

	// Start elevator state server
	go elevator.Elevator_Server(commands)
    
    drv_buttons := make(chan elevator.ButtonEvent_t)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevator.PollButtons(drv_buttons)
    go elevator.PollFloorSensor(drv_floors)
    go elevator.PollObstructionSwitch(drv_obstr)
    go elevator.PollStopButton(drv_stop)

    // Init FSM (handle between floors)
	elevator.OnInitBetweenFloors(commands)

	// Door timer
	doorTimerStart := make(chan time.Duration)
	doorTimerStop  := make(chan struct{})
	doorTimeout    := make(chan struct{})

	go elevator.DoorTimer(doorTimerStart, doorTimerStop, doorTimeout)

    
    
    for {
		select {

		// Button pressed
		case btn := <-drv_buttons:
			elevator.OnRequestButtonPress(commands, doorTimerStart, doorTimerStop, btn.Floor, btn.Button)

		// Floor arrival
		case floor := <-drv_floors:
			elevator.OnFloorArrival(commands, doorTimerStart, doorTimerStop, floor)

		// Door timeout
		case <-doorTimeout:
			elevator.OnDoorTimeout(commands, doorTimerStart, doorTimerStop)

		// Stop button
		case stop := <-drv_stop:
			if stop {
				elevator.SetStopLamp(true)
			} else {
				elevator.SetStopLamp(false)
			}

		// Obstruction
		case obstructed := <-drv_obstr:
			if obstructed {
				doorTimerStop <- struct{}{}
			} else {
				var e_state elevator.ElevatorState_t = elevator.GetState(commands)
				doorTimerStop <- struct{}{}
				doorTimerStart <- e_state.DoorOpenDuration
			}
		}
	}    
} 