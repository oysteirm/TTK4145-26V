package main

import "Driver-go/elevio"
import "fmt"

func main(){

    num_floors := 4
    num_elevators := 1

    elevio.Init("localhost:15657", numFloors)

    commands := make(chan elevator.Command)

	// Start elevator state server
	go elevator.Elevator_Server(commands)
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    // Init FSM (handle between floors)
	fsm.OnInitBetweenFloors(commands)

	// Door timer
	doorTimer := time.NewTimer(0)
	doorTimer.Stop()
    
    
    for {
		select {

		// Button pressed
		case btn := <-drvButtons:
			fsm.OnRequestButtonPress(commands, btn.Floor, btn.Button)

		// Floor arrival
		case floor := <-drvFloors:
			fsm.OnFloorArrival(commands, floor)

		// Door timeout
		case <-doorTimer.C:
			fsm.OnDoorTimeout(commands)

		// Stop button
		case stop := <-drvStop:
			if stop {
				elevio.SetStopLamp(true)
			} else {
				elevio.SetStopLamp(false)
			}

		// Obstruction
		case obstructed := <-drvObstr:
			if obstructed {
				doorTimer.Stop()
			}
		}
	}    
} 