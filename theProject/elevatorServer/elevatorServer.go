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
	- Reacts to the different events of elevator finite state machine (FSM)
	  (Floor arrival, New data, Obstruction, Door closing)
	- Receives confirmed system data from message sync and uses this for requests assigning (RA)
	- Communicates new states of the FSM to the elevatorStateGuardian
-----------------------------------
*/

func ElevatorServer(
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t, 	
	requestToMsgSync chan<- []elevator_IO.ButtonEvent_t, 		
	systemDataFromMsgSync <-chan messageSync.SystemData_t, 		
	localID int, ) { 											

	isObstructed := false
	
	// Channel used to set and get data stored in the elevatorStateGuardian routine
	guardianCommands := make(chan elevatorStateGuardian.GuardianCommands_t, 32)

	go elevatorStateGuardian.ElevatorStateGuardian(guardianCommands, elevatorDataToMsgSync, requestToMsgSync, localID)

	// Polling on the elevator hardware
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	go elevator_IO.PollFloorSensor(drv_floors)
	go elevator_IO.PollObstructionSwitch(drv_obstr)
	go elevator_IO.PollStopButton(drv_stop)

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

		case newSystemData := <-systemDataFromMsgSync:

			// Not accepting unitialized floor value
			if newSystemData.ElevatorData[localID].Floor == -1 {
				break
			}

			guardianCommands <- elevatorStateGuardian.SetSystemData_t{SystemData: newSystemData}
			fsm.LightCabLights(newSystemData.ElevatorData[localID].CabRequests)
			fsm.LightHallLights(newSystemData.HallRequestData)

			requestsMap := requestAssigner.AssignRequests(requestAssigner.Generating_RA_SystemData(newSystemData))
			if requestsMap == nil {
				break
			}

			assignedRequests, exists := requestsMap[strconv.Itoa(localID)]
			if !exists {
				break
			}

			guardianCommands <- elevatorStateGuardian.SetAssignedRequest_t{AssignedRequests: assignedRequests}

			fsm.OnReceivedDataFromMsgSync(	guardianCommands, 
											doorTimerStart, 
											doorTimerStop, 
											isFunctionalStart, 
											isFunctionalStop)

		case floor := <-drv_floors:
			fsm.OnFloorArrival(	guardianCommands, 
								doorTimerStart, 
								doorTimerStop, 
								isFunctionalStart, 
								isFunctionalStop, 
								floor)

		case <-doorTimerTimeout:
			fsm.OnDoorTimeout(	guardianCommands, 
								doorTimerStart, 
								doorTimerStop, 
								isFunctionalStart, 
								isFunctionalStop, 
								isObstructed)

		case <-isFunctionalTimeout:
			guardianCommands <- elevatorStateGuardian.SetIsFunctional_t{IsFunctional: false}

		case stop := <-drv_stop:
			if stop {
				elevator_IO.SetStopLamp(true)
			} else {
				elevator_IO.SetStopLamp(false)
			}

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
