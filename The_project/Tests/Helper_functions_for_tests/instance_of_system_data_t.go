package test_helpers



import (
	"TTK4145-26V/elevator"
	"TTK4145-26V/message_sync"
)

// making "confirmed" System_Data_t example
// floors (4)
// nElev (3)
func Make_Fake_Confirmed_System_Data_t(floors int, nElev int) message_sync.System_Data_t {
	var confirmed message_sync.System_Data_t

	// ---- Hall requests: [floors][2] ----
	confirmed.Hall_Request_Data = make([][2]message_sync.Request_Cyclic_Counter_t, floors)

	setHall := func(floor int, btn int) {
		if floor < 0 || floor >= floors {
			return
		}
		if btn < 0 || btn >= 2 {
			return
		}
		confirmed.Hall_Request_Data[floor][btn].Value = message_sync.CC_Confirmed
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
	confirmed.Elevator_Data = make([]message_sync.Elevator_Data_t, nElev)


	setCab := func(eIdx int, floor int) {
		if eIdx < 0 || eIdx >= nElev {
			return
		}
		if floor < 0 || floor >= floors {
			return
		}
		if confirmed.Elevator_Data[eIdx].Cab_Requests == nil {
			confirmed.Elevator_Data[eIdx].Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, floors)
		}
		confirmed.Elevator_Data[eIdx].Cab_Requests[floor].Value = message_sync.CC_Confirmed
	}

	// Elevator 1
	if nElev >= 1 {
		e := &confirmed.Elevator_Data[0]
		e.Id = 1
		e.Is_Alive.Value = true
		e.Is_Able.Value = true
		e.Floor.Value = 0
		e.Elevator_Behaviour.Value = elevator.EB_Idle
		e.Motor_Direction.Value = elevator.MD_Stop
		e.Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, floors)

		// cab call in floor 3 (0-indexed => 2)
		setCab(0, 2)
	}

	// Elevator 2 (here set to not able -> filtered)
	if nElev >= 2 {
		e := &confirmed.Elevator_Data[1]
		e.Id = 2
		e.Is_Alive.Value = true
		e.Is_Able.Value = false
		e.Floor.Value = 2
		e.Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, floors)
	}

	// Elevator 3
	if nElev >= 3 {
		e := &confirmed.Elevator_Data[2]
		e.Id = 3
		e.Is_Alive.Value = true
		e.Is_Able.Value = true
		e.Floor.Value = 2
		e.Elevator_Behaviour.Value = elevator.EB_Moving
		e.Motor_Direction.Value = elevator.MD_Down
		e.Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, floors)

	
		setCab(2, 0)
	}

	return confirmed
}