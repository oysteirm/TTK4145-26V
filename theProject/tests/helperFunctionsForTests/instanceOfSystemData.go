package testHelpers

import (
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/messageSync"
)

// making "confirmed" SystemData_t example
// floors (4)
// nElev (3)
func MakeFakeConfirmedSystemData(floors int, nElev int) messageSync.SystemData_t {
	var confirmed messageSync.SystemData_t

	// ---- Hall requests: [floors][2] ----
	confirmed.HallRequestData  = [config.N_FLOORS][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t{}

	setHall := func(floor int, btn int) {
		if floor < 0 || floor >= floors {
			return
		}
		if btn < 0 || btn >= 2 {
			return
		}
		confirmed.HallRequestData[floor][btn].Value = messageSync.CC_Confirmed
	}

	// Example hall calls
	setHall(1, 0) // hall up in "floor 2" (REMEMBER: 0-indexed)
	setHall(2, 1) // hall down in "floor 3"
	setHall(2, 0) // hall up in "floor 3"
	setHall(0, 0) // hall up in "floor 1"

	// Elevators
	if nElev < 0 {
		nElev = 0
	}
	confirmed.ElevatorData = [config.N_ELEVATORS]messageSync.ElevatorData_t{}


	setCab := func(eIdx int, floor int) {
		if eIdx < 0 || eIdx >= nElev {
			return
		}
		if floor < 0 || floor >= floors {
			return
		}
		confirmed.ElevatorData[eIdx].CabRequests = [config.N_FLOORS]messageSync.RequestCyclicCounter_t{}
	
		confirmed.ElevatorData[eIdx].CabRequests[floor].Value = messageSync.CC_Confirmed
	}

	// Elevator 1
	if nElev >= 1 {
		e := &confirmed.ElevatorData[0]
		e.ID = 1
		e.IsAlive = true
		e.IsFunctional = true
		e.Floor = 0
		e.ElevatorBehaviour = elevator_IO.EB_Idle
		e.MotorDirection = elevator_IO.MD_Stop
		e.CabRequests = [config.N_FLOORS]messageSync.RequestCyclicCounter_t{}

		// cab call in floor 3 (0-indexed => 2)
		setCab(0, 2)
	}

	// Elevator 2 (here set to not able -> filtered)
	if nElev >= 2 {
		e := &confirmed.ElevatorData[1]
		e.ID = 2
		e.IsAlive = true
		e.IsFunctional = false
		e.Floor = 2
		e.CabRequests = [config.N_FLOORS]messageSync.RequestCyclicCounter_t{}
	}

	// Elevator 3
	if nElev >= 3 {
		e := &confirmed.ElevatorData[2]
		e.ID = 3
		e.IsAlive = true
		e.IsFunctional = true
		e.Floor = 2
		e.ElevatorBehaviour = elevator_IO.EB_Moving
		e.MotorDirection = elevator_IO.MD_Down
		e.CabRequests = [config.N_FLOORS]messageSync.RequestCyclicCounter_t{}

	
		setCab(2, 0)
	}

	return confirmed
}