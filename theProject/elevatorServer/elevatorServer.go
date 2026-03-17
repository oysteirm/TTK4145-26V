package elevatorServer

import (
	"strconv"
	"theProject/elevatorStateGuardian"
	"theProject/elevator_IO"
	"theProject/fsm"
	"theProject/messageSync"
	"theProject/requestAssigner"
	"theProject/timer"
)

/*
-----------------------------------
Functionality: 
	- Initiaize the elevator IO through ioAddr
	- Reacts to the different events of elevator finite state machine (FSM)
	  (Floor arrival, New data, Obstruction, Door closing)
	- Recieves confirmed system data from message sync and uses this for requests assigning (RA)
	- Communicates new states of the FSM to the elevatorStateGuardian
-----------------------------------
*/

func ElevatorServer(
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t, 	//channel for sending data to messageSyncServer
	requestToMsgSync chan<- []elevator_IO.ButtonEvent_t, 		//channel for sending done request CC to msg sync
	systemDataFromMsgSync <-chan messageSync.SystemData_t, 		//channel for receiving confirmed system data
	localID int, ) { 											//ID of local elevator

	// Remebering the local state of the obstruction lever
	isObstructed := false
	
	// Channel used to set and get data stored in the elevatorStateGuardian routine
	guardianCommands := make(chan elevatorStateGuardian.GuardianCommands_t, 32)

	// Start elevator state guardian, used to safely store the systemData
	go elevatorStateGuardian.ElevatorStateGuardian(guardianCommands, elevatorDataToMsgSync, requestToMsgSync, localID)

	// Polling on the elevator hardware
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	go elevator_IO.PollFloorSensor(drv_floors)
	go elevator_IO.PollObstructionSwitch(drv_obstr)
	go elevator_IO.PollStopButton(drv_stop)

	// Init FSM (handling starting between floors)
	fsm.OnInitBetweenFloors(guardianCommands, drv_floors)

	// Timers for door and detecting physical functionallity failure
	doorTimerStart := make(chan struct{}, 1)
	doorTimerStop := make(chan struct{}, 1)
	doorTimerTimeout := make(chan struct{})
	isFunctionalStart := make(chan struct{}, 1)
	isFunctionalStop := make(chan struct{}, 1)
	isFunctionalTimeout := make(chan struct{})
	go timer.Timers(doorTimerStart, 
					doorTimerStop, 
					doorTimerTimeout, 
					isFunctionalStart, 
					isFunctionalStop, 
					isFunctionalTimeout)

	// Loop reacting on the FSM events
	for {
		select {

		// Recieved data from msg sync
		case newSystemData := <-systemDataFromMsgSync:

			// Not accepting unitialized floor value
			if newSystemData.ElevatorData[localID].Floor == -1 {break}

			// Storing the new systemData in elevatorStateGuardian and lighting the button
			guardianCommands <- elevatorStateGuardian.SetSystemData_t{SystemData: newSystemData}
			fsm.LightCabLights(newSystemData.ElevatorData[localID].CabRequests)
			fsm.LightHallLights(newSystemData.HallRequestData)

			// Calculating the assignedRequests and breaking if the is not any
			requestsMap := requestAssigner.AssignRequests(requestAssigner.Generating_RA_SystemData(newSystemData))
			if requestsMap == nil {break}

			// picking the local assigned requests based on ID and breaking if not any
			assignedRequests, exists := requestsMap[strconv.Itoa(localID)]
			if !exists {break}

			//Store requests, send this and the confirmed system data
			guardianCommands <- elevatorStateGuardian.SetAssignedRequest_t{AssignedRequests: assignedRequests}

			// Reacting to the new data
			fsm.OnReceivedDataFromMsgSync(	guardianCommands, 
											doorTimerStart, 
											doorTimerStop, 
											isFunctionalStart, 
											isFunctionalStop, 
											isObstructed)

		// Floor arrival
		case floor := <-drv_floors:
			fsm.OnFloorArrival(	guardianCommands, 
								doorTimerStart, 
								doorTimerStop, 
								isFunctionalStart, 
								isFunctionalStop, 
								floor, 
								isObstructed)

		// Door timeout
		case <-doorTimerTimeout:
			fsm.OnDoorTimeout(	guardianCommands, 
								doorTimerStart, 
								doorTimerStop, 
								isFunctionalStart, 
								isFunctionalStop, 
								isObstructed)

		// IsFunctional timeout and marking ourself as not funtional
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
				isObstructed = true
				guardianCommands <- elevatorStateGuardian.SetIsFunctional_t{IsFunctional: false}
			} else {
				isObstructed = false
				guardianCommands <- elevatorStateGuardian.SetIsFunctional_t{IsFunctional: true}
			}
		}
	}
}
