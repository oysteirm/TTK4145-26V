package requests

import (
	"theProject/messageSync"
	"theProject/elevatorIo"
)

// behavour-strukt, for retning og tilstand
type MotorDirectionBehaviourPair_t struct {
	MotorDirection    elevatorIo.MotorDirection_t
	ElevatorBehaviour elevatorIo.ElevatorBehaviour_t
}

// requests_choose_direction tilsvarer: requests_chooseDirection(ElevatorState e_state)

//if troubles, look over what we was delivered
func RequestsChooseDirection(
	e_state messageSync.ElevatorData_t, 
	assignedRequests elevatorIo.AssignedRequests_t) MotorDirectionBehaviourPair_t {

	switch e_state.MotorDirection {
	case elevatorIo.MD_Up:
		if requestsAbove(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Up, elevatorIo.EB_Moving}
		} else if requestsHere(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Down, elevatorIo.EB_DoorOpen}
		} else if requestsBelow(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Down, elevatorIo.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Stop, elevatorIo.EB_Idle}
		}

	case elevatorIo.MD_Down:
		if requestsBelow(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Down, elevatorIo.EB_Moving}
		} else if requestsHere(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Up, elevatorIo.EB_DoorOpen}
		} else if requestsAbove(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Up, elevatorIo.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Stop, elevatorIo.EB_Idle}
		}

	case elevatorIo.MD_Stop:
		if requestsHere(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Stop, elevatorIo.EB_DoorOpen}
		} else if requestsAbove(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Up, elevatorIo.EB_Moving}
		} else if requestsBelow(e_state, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Down, elevatorIo.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevatorIo.MD_Stop, elevatorIo.EB_Idle}
		}

	default:
		return MotorDirectionBehaviourPair_t{elevatorIo.MD_Stop, elevatorIo.EB_Idle}
	}
}

func RequestsShouldStop(
	e_state messageSync.ElevatorData_t, 
	assignedRequests elevatorIo.AssignedRequests_t) bool {

	// Requests [][]bool:
	if e_state.Floor < 0 || e_state.Floor >= len(assignedRequests) {
		return false
	}

	switch e_state.MotorDirection {
	case elevatorIo.MD_Down:
		return assignedRequests[e_state.Floor][elevatorIo.BT_HallDown] ||
			assignedRequests[e_state.Floor][elevatorIo.BT_Cab] ||
			!requestsBelow(e_state, assignedRequests)

	case elevatorIo.MD_Up:
		return assignedRequests[e_state.Floor][elevatorIo.BT_HallUp] ||
			assignedRequests[e_state.Floor][elevatorIo.BT_Cab] ||
			!requestsAbove(e_state, assignedRequests)

	case elevatorIo.MD_Stop:
		fallthrough
	default:
		return true
	}
}

// there is a lot of ugly if statements here
func RequestsClearAtCurrentFloor(e_state messageSync.ElevatorData_t, assignedRequests elevatorIo.AssignedRequests_t) []elevatorIo.ButtonEvent_t {
	var requestsToClear []elevatorIo.ButtonEvent_t

	if assignedRequests[e_state.Floor][elevatorIo.BT_Cab]{
		requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_Cab})
	}

	switch e_state.MotorDirection {
	case elevatorIo.MD_Up:
		if !requestsAbove(e_state, assignedRequests) && !assignedRequests[e_state.Floor][elevatorIo.BT_HallUp] {
			if assignedRequests[e_state.Floor][elevatorIo.BT_HallDown]{
				requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_HallDown})
			}
		}
		if assignedRequests[e_state.Floor][elevatorIo.BT_HallUp]{
			requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_HallUp})
		}
	case elevatorIo.MD_Down:
		if !requestsBelow(e_state, assignedRequests) && !assignedRequests[e_state.Floor][elevatorIo.BT_HallDown] {
			if assignedRequests[e_state.Floor][elevatorIo.BT_HallUp]{
			requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_HallUp})
			}
		}
		if assignedRequests[e_state.Floor][elevatorIo.BT_HallDown]{
				requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_HallDown})
		}

	case elevatorIo.MD_Stop:
		fallthrough
	default:
		//CLEARING BOTH UP AND DOWN; BUT ONLY SUPPOSED TO CLEAR ONE?
		if assignedRequests[e_state.Floor][elevatorIo.BT_HallUp]{
			requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_HallUp})
		}
		if assignedRequests[e_state.Floor][elevatorIo.BT_HallDown]{
				requestsToClear = append(requestsToClear, elevatorIo.ButtonEvent_t{Floor: e_state.Floor, Button: elevatorIo.BT_HallDown})
		}
	}

	return requestsToClear
}

// --- “static” helpers ---

func requestsAbove(e_state messageSync.ElevatorData_t, assignedRequests elevatorIo.AssignedRequests_t) bool {
	for f := e_state.Floor + 1; f < elevatorIo.N_FLOORS; f++ {
		for btn := elevatorIo.ButtonType_t(0); btn < elevatorIo.N_BUTTONS; btn++ {
			if assignedRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e_state messageSync.ElevatorData_t, assignedRequests elevatorIo.AssignedRequests_t) bool {
	for f := 0; f < e_state.Floor; f++ {
		for btn := elevatorIo.ButtonType_t(0); btn < elevatorIo.N_BUTTONS; btn++ {
			if assignedRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e_state messageSync.ElevatorData_t, assignedRequests elevatorIo.AssignedRequests_t) bool {
	for btn := elevatorIo.ButtonType_t(0); btn < elevatorIo.N_BUTTONS; btn++ {
		if assignedRequests[e_state.Floor][btn] {
			return true
		}
	}
	return false
}
