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
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t, //channel for sending data to messageSyncServer
	requestToMsgSync chan<- []elevator_IO.ButtonEvent_t, //channel for sending done request CC to msg sync
	systemDataFromMsgSync <-chan messageSync.SystemData_t, //channel for receiving confirmed system data
	ioAddr string,
	localID int, //local ID
) {

	elevator_IO.Init(ioAddr, config.N_FLOORS)
	isObstructed := false

	guardianCommands := make(chan elevatorStateGuardian.GuardianCommands_t,32)

	// Start elevator state server
	go elevatorStateGuardian.ElevatorStateGuardian(guardianCommands, elevatorDataToMsgSync, requestToMsgSync, localID)

	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevator_IO.PollFloorSensor(drv_floors)
	go elevator_IO.PollObstructionSwitch(drv_obstr)
	go elevator_IO.PollStopButton(drv_stop)

	// Init FSM (handle between floors)
	fsm.OnInitBetweenFloors(guardianCommands, drv_floors)

	// Timers
	doorTimerStart := make(chan struct{})
	doorTimerStop := make(chan struct{})
	doorTimerTimeout := make(chan struct{})
	isFunctionalStart := make(chan struct{})
	isFunctionalStop := make(chan struct{})
	isFunctionalTimeout := make(chan struct{})

	//work in progress
	go timer.Timers(doorTimerStart, doorTimerStop, doorTimerTimeout, isFunctionalStart, isFunctionalStop, isFunctionalTimeout)

	for {
		select {

		// Recieved data from msg sync
		case newSystemData := <-systemDataFromMsgSync:

			if newSystemData.ElevatorData[localID].Floor == -1 {
				break
			}
			requestsMap := requestAssigner.AssignRequests(requestAssigner.Generating_RA_SystemData(newSystemData))
			if requestsMap == nil {
				break
			}

			assignedRequests, exists := requestsMap[strconv.Itoa(localID)]
			if !exists {
				break
			}

			//Store requests, send this and the confirmed system data
			guardianCommands <- elevatorStateGuardian.SetSystemData_t{SystemData: newSystemData}
			guardianCommands <- elevatorStateGuardian.SetAssignedRequest_t{AssignedRequests: assignedRequests}

			fsm.LightCabLights(newSystemData.ElevatorData[localID].CabRequests)
			fsm.LightHallLights(newSystemData.HallRequestData)

			fsm.OnReceivedDataFromMsgSync(guardianCommands, doorTimerStart, doorTimerStop, isFunctionalStart, isFunctionalStop, isObstructed)

		// Floor arrival
		case floor := <-drv_floors:
			fsm.OnFloorArrival(guardianCommands, doorTimerStart, doorTimerStop, isFunctionalStart, isFunctionalStop, floor, isObstructed)

		// Door timeout
		case <-doorTimerTimeout:
			fsm.OnDoorTimeout(guardianCommands, doorTimerStart, doorTimerStop, isFunctionalStart, isFunctionalStop, isObstructed)
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

		// Obstruction, this does not work
		//maybe fix: add a state, and have this functionallity outside the select case'
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
