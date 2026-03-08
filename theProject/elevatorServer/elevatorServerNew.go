package elevatorServer

import (
	"strconv"
	"theProject/config"
	"theProject/elevatorStateGuardian"
	"theProject/elevator_IO"
	"theProject/fsm"
	"theProject/messageSync"
	"theProject/requestAssigner"
	"theProject/timer"
)


func ElevatorServer(
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t,    //channel for sending data to messageSyncServer
    requestToMsgSync chan<- messageSync.RequestCyclicCounter_t,	//channel for sending done request CC to msg sync
	systemDataFromMsgSync <-chan messageSync.SystemData_t,		//channel for receiving confirmed system data
	localID int,												//local ID
){

    elevator_IO.Init("localhost:15657", config.N_FLOORS)

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

	// Timers
	doorTimerStart       := make(chan struct{})
	doorTimerStop        := make(chan struct{})
	doorTimerTimeout     := make(chan struct{})
	isFunctionalStart    := make(chan struct{})
	isFunctionalStop     := make(chan struct{})
	isFunctionalTimeout  := make(chan struct{})

	//work in progress
	go timer.Timers(doorTimerStart, doorTimerStop, doorTimerTimeout, isFunctionalStart, isFunctionalStop, isFunctionalTimeout)
    
    for {
		select {

		// Recieved data from msg sync
		case newSystemData := <-systemDataFromMsgSync:

			//Use the RA
			//Use the RA
			requestsMap := requestAssigner.AssignRequests(requestAssigner.Generating_RA_SystemData(newSystemData))

			assignedRequests := requestsMap[strconv.Itoa(localID)]

			//Store requests, send this and the confirmed system data
			guardianCommands <- newSystemData
			guardianCommands <- assignedRequests

			fsm.LightCabLights(newSystemData.ElevatorData[localID].CabRequests)
			fsm.LightHallLights(newSystemData.HallRequestData)

			fsm.OnReceivedDataFromMsgSync(guardianCommands, doorTimerStart, doorTimerStop, isFunctionalStart, isFunctionalStop)

		// Floor arrival
		case floor := <-drv_floors:
			fsm.OnFloorArrival(guardianCommands, doorTimerStart, doorTimerStop, isFunctionalStart, isFunctionalStop, floor)
			//TODO: add functionallity for updating IsFunctional

		// Door timeout
		case <-doorTimerTimeout:
			fsm.OnDoorTimeout(guardianCommands, doorTimerStart, doorTimerStop, isFunctionalStart, isFunctionalStop)
			//TODO: fix the double requests issue

		//isFunctional timeout
		case <-isFunctionalTimeout:
			guardianCommands <- elevatorStateGuardian.SetIsFunctional_t{IsFunctional: false}

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
				if elevatorStateGuardian.GetElevatorData(guardianCommands).ElevatorBehaviour == elevator_IO.EB_DoorOpen{
					guardianCommands <- elevatorStateGuardian.SetIsFunctional_t{IsFunctional: false}
				}
			} else {
				doorTimerStop <- struct{}{}
				doorTimerStart <- struct{}{}
			}
		}
	}    
} 

