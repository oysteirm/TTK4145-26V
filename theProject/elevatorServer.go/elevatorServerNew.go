package elevatorServer

import (
	//"fmt"
	"os"
	"theProject/networkDriver/peers"
	"theProject/requestAssigner"
	"theProject/messageSync"
	"theProject/elevator"
	"time"
)


func ElevatorServer(){

    elevator.Init("localhost:15657", elevator.N_FLOORS)

    commands := make(chan elevator.Command_t)

	// Start elevator state server
	go elevator.ElevatorStateGuardian(commands)
    
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevator.PollFloorSensor(drv_floors)
    go elevator.PollObstructionSwitch(drv_obstr)
    go elevator.PollStopButton(drv_stop)

    // Init FSM (handle between floors)
	elevator.OnInitBetweenFloors(commands)

	// Door timer
	doorTimerStart    := make(chan time.Duration)
	doorTimerStop     := make(chan struct{})
	doorTimerTimeout  := make(chan struct{})
	obstruction       := make(chan struct{})
	inactiveStart     := make(chan struct{})
	inactiveStop      := make(chan struct{})
	setFunctional     := make(chan bool)


	go elevator.Timers(doorTimerStart, doorTimerStop, doorTimerTimeout, obstruction, inactiveStart, inactiveStop, setFunctional)
    
    for {
		select {

		// Recieved data from msg sync
		case btn := <-drv_buttons:
			elevator.OnRequestButtonPress(commands, doorTimerStart, doorTimerStop, btn.Floor, btn.Button)

		// Floor arrival
		case floor := <-drv_floors:
			elevator.OnFloorArrival(commands, doorTimerStart, doorTimerStop, inactiveStart, inactiveStop)
			//TODO: add functionallity for updating IsFunctional

		// Door timeout
		case <-doorTimerTimeout:
			elevator.OnDoorTimeout(commands, doorTimerStart, doorTimerStop)
			//TODO: fix the double requests issue
			
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