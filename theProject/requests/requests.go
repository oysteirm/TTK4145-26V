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

// requests_choose_direction tilsvarer: requests_chooseDirection(ElevatorState elevatorState)

//if troubles, look over what we was delivered
func RequestsChooseDirection(
	elevatorState messageSync.ElevatorData_t, 
	assignedRequests elevator_IO.AssignedRequests_t) MotorDirectionBehaviourPair_t {

	switch elevatorState.MotorDirection {
	case elevator_IO.MD_Up:
		if requestsAbove(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_Moving}
		} else if requestsHere(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_DoorOpen}
		} else if requestsBelow(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
		}

	case elevator_IO.MD_Down:
		if requestsBelow(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_Moving}
		} else if requestsHere(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_DoorOpen}
		} else if requestsAbove(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
		}

	case elevator_IO.MD_Stop:
		if requestsHere(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_DoorOpen}
		} else if requestsAbove(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Up, elevator_IO.EB_Moving}
		} else if requestsBelow(elevatorState, assignedRequests) {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Down, elevator_IO.EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
		}

	default:
		return MotorDirectionBehaviourPair_t{elevator_IO.MD_Stop, elevator_IO.EB_Idle}
	}
}

func RequestsShouldStop(
	elevatorState messageSync.ElevatorData_t, 
	assignedRequests elevator_IO.AssignedRequests_t) bool {


	if elevatorState.Floor < 0 || elevatorState.Floor >= len(assignedRequests) {
		return false
	}
	
	

	switch elevatorState.MotorDirection {
	case elevator_IO.MD_Down:
		return assignedRequests[elevatorState.Floor][elevator_IO.BT_HallDown] ||
			assignedRequests[elevatorState.Floor][elevator_IO.BT_Cab] ||
			!requestsBelow(elevatorState, assignedRequests)

	case elevator_IO.MD_Up:
		return assignedRequests[elevatorState.Floor][elevator_IO.BT_HallUp] ||
			assignedRequests[elevatorState.Floor][elevator_IO.BT_Cab] ||
			!requestsAbove(elevatorState, assignedRequests)

	case elevator_IO.MD_Stop:
		fallthrough

	default:
		return true
	}
}

// SHOULD THIS BE THINNER???
func RequestsClearAtCurrentFloor(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) []elevator_IO.ButtonEvent_t {
	
	var requestsToClear []elevator_IO.ButtonEvent_t
	requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_Cab, requestsToClear)
	
	switch elevatorState.MotorDirection {

	case elevator_IO.MD_Up:
		if !requestsAbove(elevatorState, assignedRequests) && !assignedRequests[elevatorState.Floor][elevator_IO.BT_HallUp] {
			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
		}
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

	case elevator_IO.MD_Down:
		if !requestsBelow(elevatorState, assignedRequests) && !assignedRequests[elevatorState.Floor][elevator_IO.BT_HallDown] {
			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)
		}
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)

	case elevator_IO.MD_Stop:
		fallthrough

	default:
		//CLEARING BOTH UP AND DOWN; BUT ONLY SUPPOSED TO CLEAR ONE?
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
	}

	return requestsToClear
}

// --- “static” helpers ---

func validFloor(floor int)bool{
	return floor>=0 && floor <elevator_IO.N_FLOORS
}

func requestsAbove(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) bool {

	
	for floor := elevatorState.Floor + 1; floor < elevator_IO.N_FLOORS; floor++ {
		for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {
			if assignedRequests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) bool {

	for floor := 0; floor < elevatorState.Floor; floor++ {
		for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {
			if assignedRequests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) bool {
	
	if !validFloor(elevatorState.Floor){
		return false
	}

	for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {
		if assignedRequests[elevatorState.Floor][btn] {
			return true
		}
	}
	return false
}

//GIVE BETTER NAME?
func appendRequestsToClearIfExisting(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t,
	button elevator_IO.ButtonType_t,
	requestsToClear []elevator_IO.ButtonEvent_t) []elevator_IO.ButtonEvent_t{
	
	if assignedRequests[elevatorState.Floor][button]{
		requestsToClear = append(requestsToClear, elevator_IO.ButtonEvent_t{Floor: elevatorState.Floor, Button: button})
	}
	
	return requestsToClear

}

//in progess below

func RequestsClearOnFloorArrival(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) []elevator_IO.ButtonEvent_t {
	
	var requestsToClear []elevator_IO.ButtonEvent_t
	requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_Cab, requestsToClear)
	
	switch elevatorState.MotorDirection {

	case elevator_IO.MD_Up:
		
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

	case elevator_IO.MD_Down:
		
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)

	case elevator_IO.MD_Stop:
		fallthrough
	default:
	}

	return requestsToClear
}

func RequestsClearOnDoorTimeout(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) []elevator_IO.ButtonEvent_t {
	
	var requestsToClear []elevator_IO.ButtonEvent_t
	requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_Cab, requestsToClear)
	
	switch elevatorState.MotorDirection {

	case elevator_IO.MD_Up:
		if !requestsAbove(elevatorState, assignedRequests) && !assignedRequests[elevatorState.Floor][elevator_IO.BT_HallUp] {
			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
		}
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

	case elevator_IO.MD_Down:
		if !requestsBelow(elevatorState, assignedRequests) && !assignedRequests[elevatorState.Floor][elevator_IO.BT_HallDown] {
			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)
		}
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)

	case elevator_IO.MD_Stop:
		fallthrough
	default:
	}

	return requestsToClear
}

func RequestsClearOnNewData(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t) []elevator_IO.ButtonEvent_t {
	
	var requestsToClear []elevator_IO.ButtonEvent_t
	requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_Cab, requestsToClear)
	
	switch elevatorState.MotorDirection {

	case elevator_IO.MD_Up:
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

	case elevator_IO.MD_Down:
		requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)

	case elevator_IO.MD_Stop:
		fallthrough

	default:
		if assignedRequests[elevatorState.Floor][elevator_IO.BT_HallUp]{

			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

		} else {
			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
		}
	}

	return requestsToClear
}