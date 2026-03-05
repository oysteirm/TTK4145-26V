package testHelpers



import (
	"TTK4145-26V/elevator"
	"TTK4145-26V/messageSync"
)

// making "confirmed" SystemData_t example
// floors (4)
// nElev (3)
func MakeFakeConfirmedSystemData(floors int, nElev int) messageSync.SystemData_t {
	var confirmed messageSync.SystemData_t

	// ---- Hall requests: [floors][2] ----
	confirmed.HallRequestData = make([][2]messageSync.RequestCyclicCounter_t, floors)

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
	confirmed.ElevatorData = make([]messageSync.ElevatorData_t, nElev)


	setCab := func(eIdx int, floor int) {
		if eIdx < 0 || eIdx >= nElev {
			return
		}
		if floor < 0 || floor >= floors {
			return
		}
		if confirmed.ElevatorData[eIdx].CabRequests == nil {
			confirmed.ElevatorData[eIdx].CabRequests = make([]messageSync.RequestCyclicCounter_t, floors)
		}
		confirmed.ElevatorData[eIdx].CabRequests[floor].Value = messageSync.CC_Confirmed
	}

	// Elevator 1
	if nElev >= 1 {
		e := &confirmed.ElevatorData[0]
		e.Id = 1
		e.IsAlive = true
		e.IsFunctional = true
		e.Floor = 0
		e.ElevatorBehaviour = elevator.EB_Idle
		e.MotorDirection = elevator.MD_Stop
		e.CabRequests = make([]messageSync.RequestCyclicCounter_t, floors)

		// cab call in floor 3 (0-indexed => 2)
		setCab(0, 2)
	}

	// Elevator 2 (here set to not able -> filtered)
	if nElev >= 2 {
		e := &confirmed.ElevatorData[1]
		e.Id = 2
		e.IsAlive = true
		e.IsFunctonal = false
		e.Floor = 2
		e.CabRequests = make([]messageSync.RequestCyclicCounter_t, floors)
	}

	// Elevator 3
	if nElev >= 3 {
		e := &confirmed.ElevatorData[2]
		e.Id = 3
		e.IsAlive = true
		e.IsFunctional = true
		e.Floor = 2
		e.ElevatorBehaviour = elevator.EB_Moving
		e.MotorDirection = elevator.MD_Down
		e.CabRequests = make([]messageSync.RequestCyclicCounter_t, floors)

	
		setCab(2, 0)
	}

	return confirmed
}