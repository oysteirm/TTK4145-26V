package main

import (
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/messageSync"
	"theProject/fsm"
)


func main(){
	testUpdateElevatorDataAboutSelf()
	
}

func testUpdateHallRequests(){
	localID := 0;
	oldHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	fullBarrier := []bool{true, true, true}
	messageSync.ActivePeers = [config.N_ELEVATORS]bool{true, true, true}

	newHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
	
			oldHallRequests[floor][btn].Value = messageSync.CC_No
			oldHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)

			newHallRequests[floor][btn].Value = messageSync.CC_Unconfirmed
			newHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)
		}
	}

	oldHallRequests = messageSync.UpdateHallRequestData(oldHallRequests, newHallRequests, localID)

	println("Active Peers:")
	for i := 0; i<3 ; i++{
		print(messageSync.ActivePeers[i], " ")
	} 
	println()
	
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			println(oldHallRequests[floor][btn].Value)
			for i := 0; i<3 ; i++{
				print(oldHallRequests[floor][btn].Barrier[i], " ")
			}
			println()
		}
	}
}

func testUpdateElevatorDataAboutSelf(){

	localID := 0

	oldData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 2,
		MotorDirection: elevator_IO.MD_Stop,
		ElevatorBehaviour: elevator_IO.EB_Idle,
		ElevatorBarrier: []bool{true, true, true},
	}
	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: []bool{true, false, false},
		CabRequests: make([]RequestCyclicCounter_t{},
	}

	assignedRequests := make([][]bool, elevator_IO.N_FLOORS)
    for i := range assignedRequests {
        assignedRequests[i] = make([]bool, elevator_IO.N_BUTTONS)
    }

	oldData = messageSync.UpdateElevatorDataAboutSelf(oldData, newData, localID)

	fsm.ElevatorPrint(oldData, assignedRequests)

}

