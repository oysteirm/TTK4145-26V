package main

import "elevator"
import "fmt"

func main(){

    num_floors := 4
    num_elevators := 1

    elevator.Init("localhost:15657", numFloors)

    commands := make(chan elevator.Command)

	// Start elevator state server
	go elevator.Elevator_Server(commands)
    
    drv_buttons := make(chan elevator.ButtonEvent)
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
	doorTimer := time.NewTimer(0)
	doorTimer.Stop()
    
    
    for {
		select {

		// Button pressed
		case btn := <-drvButtons:
			elevator.OnRequestButtonPress(commands, btn.Floor, btn.Button)

		// Floor arrival
		case floor := <-drvFloors:
			elevator.OnFloorArrival(commands, floor)

		// Door timeout
		case <-doorTimer.C:
			elevator.OnDoorTimeout(commands)

		// Stop button
		case stop := <-drvStop:
			if stop {
				elevator.SetStopLamp(true)
			} else {
				elevator.SetStopLamp(false)
			}

		// Obstruction
		case obstructed := <-drvObstr:
			if obstructed {
				doorTimer.Stop()
			}
		}
	}    
} 