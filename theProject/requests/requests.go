package requests

import (
	"theProject/messageSync"
	"theProject/elevator_IO"
)

// behavour-strukt, for retning og tilstand
type MotorDirectionBehaviourPair_t struct {
	MotorDirection    elevator_IO.MotorDirection_t
	ElevatorBehaviour elevator_IO.ElevatorBehaviour_t
}

// requests_choose_direction tilsvarer: requests_chooseDirection(ElevatorState e_state)

//if troubles, look over what we was delivered
func RequestsChooseDirection(
	e_state messageSync.ElevatorData_t, 
	assignedRequests elevator_IO.AssignedRequests_t) MotorDirectionBehaviourPair_t {

	switch e_state.MotorDirection {
	case elevator_IO.MD_Up:
		if requestsAbove(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_Moving}
		} else if requestsHere(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_DoorOpen}
		} else if requestsBelow(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
		}

	case elevator_IO.MD_Down:
		if requestsBelow(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_Moving}
		} else if requestsHere(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_DoorOpen}
		} else if requestsAbove(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
		}

	case elevator_IO.MD_Stop:
		if requestsHere(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_DoorOpen}
		} else if requestsAbove(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_Moving}
		} else if requestsBelow(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
		}

	default:
		return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
	}
}

func RequestsShouldStop(
	e_state messageSync.ElevatorData_t, 
	assignedRequests elevator_IO.AssignedRequests_t) bool {

	// Requests [][]bool:
	if e_state.Floor < 0 || e_state.Floor >= len(assignedRequests) {
		return false
	}

	switch e_state.MotorDirection {
	case elevator_IO.MD_Down:
		return assignedRequests[e_state.Floor][elevator_IO.BT_HallDown] ||
			assignedRequests[e_state.Floor][elevator_IO.BT_Cab] ||
			!requestsBelow(e_state, assignedRequests)

	case elevator_IO.MD_Up:
		return assignedRequests[e_state.Floor][elevator_IO.BT_HallUp] ||
			assignedRequests[e_state.Floor][elevator_IO.BT_Cab] ||
			!requestsAbove(e_state, assignedRequests)

	case elevator_IO.MD_Stop:
		fallthrough

	default:
		return true
	}
}

// SHOULD THIS BE THINNER???
func RequestsClearAtCurrentFloor(
	e_state messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) []elevator_IO.ButtonEvent_t {
	
	var requestsToClear []elevator_IO.ButtonEvent_t
	
	requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_Cab, requestsToClear)
	
	switch e_state.MotorDirection {

	case elevator_IO.MD_Up:
		if !requestsAbove(e_state, assignedRequests) && !assignedRequests[e_state.Floor][elevator_IO.BT_HallUp] {
			requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
		}
		requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

	case elevator_IO.MD_Down:
		if !requestsBelow(e_state, assignedRequests) && !assignedRequests[e_state.Floor][elevator_IO.BT_HallDown] {
			requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)
		}
		requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)

	case elevator_IO.MD_Stop:
		fallthrough

	default:
		//CLEARING BOTH UP AND DOWN; BUT ONLY SUPPOSED TO CLEAR ONE?
		requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)
		requestsToClear = appendRequestsToClearIfExisting(e_state, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
	}

	return requestsToClear
}

// --- “static” helpers ---

func requestsAbove(
	e_state messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) bool {
	for f := e_state.Floor + 1; f < elevator_IO.N_FLOORS; f++ {
		for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {
			if assignedRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(
	e_state messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) bool {

	for f := 0; f < e_state.Floor; f++ {
		for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {
			if assignedRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(
	e_state messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) bool {

	for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {
		if assignedRequests[e_state.Floor][btn] {
			return true
		}
	}
	return false
}

//GIVE BETTER NAME?
func appendRequestsToClearIfExisting(
	e_state messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t,
	button elevator_IO.ButtonType_t,
	requestsToClear []elevator_IO.ButtonEvent_t) []elevator_IO.ButtonEvent_t{
	
	if assignedRequests[e_state.Floor][button]{
		requestsToClear = append(requestsToClear, elevator_IO.ButtonEvent_t{Floor: e_state.Floor, Button: button})
	}
	
	return requestsToClear

}