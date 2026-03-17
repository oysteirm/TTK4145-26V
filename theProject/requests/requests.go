package requests

// This code is inspired by provided code, fetched from https://github.com/TTK4145/Project-resources/blob/master/elev_algo/requests.c

import (
	"theProject/messageSync"
	"theProject/elevator_IO"
)

/*
-----------------------------------
Functionallity: 
	- Contains logic for request decision for a local elevator
	- Contains helper functions for checking valid floor and for requests below, above and at current floor
	- Decides movement direction and behaviour based on assigned requests
	- Decides when the elevator should stop
	- Decides which requests that should be cleared
	
Design:
	- The requests module does not own state itself
	- It uses assigned requests and current elevator state
	- The fsm calls these functions to decide what to do next
-----------------------------------
*/

// Pair used for deciding next motor direction and elevator behaviour
type MotorDirectionBehaviourPair_t struct {
	MotorDirection    elevator_IO.MotorDirection_t
	ElevatorBehaviour elevator_IO.ElevatorBehaviour_t
}


// Chooses the next direction and behaviour for the elevator
func RequestsChooseDirection(
	elevatorState messageSync.ElevatorData_t, 
	assignedRequests elevator_IO.AssignedRequests_t) MotorDirectionBehaviourPair_t {
	
	/* 
	Logic in switch-case:
		- continuing in the current direction if requests remain
		- otherwise serve requests at the current floor
		- If needed, reverse direction to serve remaining requests
		- If no requests exist, go idle
	*/	
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

	//Guard against invalid floor before indexing assigned requests
	if elevatorState.Floor < 0 || elevatorState.Floor >= len(assignedRequests) {
		return false
	}

	/* 
	Logic in switch-case:
		Stop if:
			- matching hall request in current direction at current floor
			- cab request at current floor
			- no more requests in current direction
			- direction already is stop
	*/
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

/*
-----------------------------------
Three functions for clearing requests at current floor due to different situations in FSM:
	- RequestsClearOnFloorArrival	
	- RequestsClearOnDoorTimeout
	- RequestsClearOnNewData


They all return a slice of button events that should be cleared when the elevator is serving the current floor,
but they use different clearing policies depending on when the decision is made.

General idea:
	Note that requests only are appended if they actually exist.
	- Cab request at current floor always appended 
	- Hall requests may be appended depending on motor direction and event type
	- Some policies are conservative, while others allow broader clearing
-----------------------------------
*/

// Conservative clearing policy for floor arrival, clearing hall request matching current travel direction and clearing cab requests 
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

// Broader policy for door timeout, to also append the opposite hall request if no further requests remain in the current direction
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


// Clearing policy for new assigned data received while the elevator is already at the floor
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
		// Choose hall up as priority, since more have to take the stairs in this case:)
		if assignedRequests[elevatorState.Floor][elevator_IO.BT_HallUp]{

			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallUp, requestsToClear)

		} else {
			requestsToClear = appendRequestsToClearIfExisting(elevatorState, assignedRequests, elevator_IO.BT_HallDown, requestsToClear)
		}
	}
	return requestsToClear
}


// Used inside alle RequestsClear... fuctions to check that request exists and to avoid repeated code when building the slice 
func appendRequestsToClearIfExisting(
	elevatorState messageSync.ElevatorData_t,
	assignedRequests elevator_IO.AssignedRequests_t,
	button elevator_IO.ButtonType_t,
	requestsToClear []elevator_IO.ButtonEvent_t) []elevator_IO.ButtonEvent_t{
	
	if !validFloor(elevatorState.Floor){
		return  requestsToClear
	} 
	
	if assignedRequests[elevatorState.Floor][button]{
		requestsToClear = append(requestsToClear, elevator_IO.ButtonEvent_t{Floor: elevatorState.Floor, Button: button})
	}
	
	return requestsToClear
}

/*
-----------------------------------
Internal helper functions for checking valid floor and for requests below, above and at current floor
*/

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




