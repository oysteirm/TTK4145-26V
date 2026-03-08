package elevatorServer

import (
	//"fmt"
	"theProject/elevator_IO"
	"theProject/fsm"
	"theProject/elevatorStateGuardian"
	"theProject/messageSync"
	"theProject/requestAssigner"
	"time"
)


func ElevatorServer(
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t,        //channel for sending data to messageSyncServer
    requestToMsgSync chan<- messageSync.RequestCyclicCounter_t,	//channel for sending done request CC to msg sync
	systemDataFromMsgSync <-chan messageSync.SystemData_t,				//channel for receiving confirmed system data
	localID int,													//local ID
){

    elevator_IO.Init("localhost:15657", elevator_IO.N_FLOORS)

    guardianCommands := make(chan elevatorStateGuardian.GuardianCommands_t)

	// Start elevator state server
	go elevatorStateGuardian.ElevatorStateGuardian(guardianCommands, elevatorDataToMsgSync, requestToMsgSync, localID)
    
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevator_IO.PollFloorSensor(drv_floors)
    go elevator_IO.PollObstructionSwitch(drv_obstr)
    go elevator_IO.PollStopButton(drv_stop)

    // Init FSM (handle between floors)
	fsm.OnInitBetweenFloors(guardianCommands)

	// Door timer
	doorTimerStart    := make(chan time.Duration)
	doorTimerStop     := make(chan struct{})
	doorTimerTimeout  := make(chan struct{})
	obstruction       := make(chan struct{})
	inactiveStart     := make(chan struct{})
	inactiveStop      := make(chan struct{})
	setFunctional     := make(chan bool)


	go elevator_IO.Timers(doorTimerStart, doorTimerStop, doorTimerTimeout, obstruction, inactiveStart, inactiveStop, setFunctional)
    
    for {
		select {

		// Recieved data from msg sync
		case btn := <-drv_buttons:
			//set lights based on confirmed data
			//Use the RA
			//Store requests, send this and the confirmed system data
			fsm.OnReceivedDataFromsMsgSync()

			elevator_IO.OnRequestButtonPress(guardianCommands, doorTimerStart, doorTimerStop, btn.Floor, btn.Button)

		// Floor arrival
		case floor := <-drv_floors:
			fsm.OnFloorArrival(guardianCommands, doorTimerStart, doorTimerStop, inactiveStart, inactiveStop)
			//TODO: add functionallity for updating IsFunctional

		// Door timeout
		case <-doorTimerTimeout:
			fsm.OnDoorTimeout(guardianCommands, doorTimerStart, doorTimerStop)
			//TODO: fix the double requests issue
			
		// Stop button
		case stop := <-drv_stop:
			if stop {
				elevator_IO.SetStopLamp(true)
			} else {
				elevator_IO.SetStopLamp(false)
			}

		// Obstruction
		case obstructed := <-drv_obstr:
			if obstructed {
				doorTimerStop <- struct{}{}
			} else {
				doorTimerStop <- struct{}{}
				doorTimerStart <- e_state.DoorOpenDuration //USE CONST!!!!
			}
		}
	}    
} 

