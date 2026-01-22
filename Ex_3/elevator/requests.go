package elevator

import (
	
)

// behavour-strukt, for retning og tilstand
type DirnBehaviourPair struct {
	Dirn              elevio.MotorDirection
	ElevatorBehaviour elevator.ElevatorBehaviour
}

// requests_choose_direction tilsvarer: requests_chooseDirection(ElevatorState e_state)



func requests_choose_direction(e_state elevator.ElevatorState) DirnBehaviourPair {
	switch e_state.Dirn {

	case elevio.MD_Up:
		if requests_above(e_state) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if requests_here(e_state) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if requests_below(e_state) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Down:
		if requests_below(e_state) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else if requests_here(e_state) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if requests_above(e_state) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if requests_here(e_state) {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if requests_above(e_state) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if requests_below(e_state) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
	}
}

func requests_should_stop(e_state elevator.ElevatorState) bool {
	// Requests [][]bool:
	if e_state.Floor < 0 || e_state.Floor >= len(e_state.Requests) {
		return false
	}

	switch e_state.Dirn {
	case elevio.MD_Down:
		return e_state.Requests[e_state.Floor][elevio.BT_HallDown] ||
			e_state.Requests[e_state.Floor][elevio.BT_Cab] ||
			!requests_below(e_state)

	case elevio.MD_Up:
		return e_state.Requests[e_state.Floor][elevio.BT_HallUp] ||
			e_state.Requests[e_state.Floor][elevio.BT_Cab] ||
			!requests_above(e_state)

	case elevio.MD_Stop:
		fallthrough
	default:
		return true
	}
}

func requests_should_clear_immediately(e_state elevator.ElevatorState, btnFloor int, btnType elevio.ButtonType) bool {
	return e_state.Floor == btnFloor &&
		((e_state.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp) ||
			(e_state.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown) ||
			e_state.Dirn == elevio.MD_Stop ||
			btnType == elevio.BT_Cab)
}

//lager en SetState i elevator, så benytter ikke commands her!
func requests_clear_at_current_floor(e_state elevator.ElevatorState) elevator.ElevatorState {
	if e_state.Floor < 0 || e_state.Floor >= len(e_state.Requests) {
		return e_state
	}

	e_state.Requests[e_state.Floor][elevio.BT_Cab] = false

	switch e_state.Dirn {
	case elevio.MD_Up:
		if !requests_above(e_state) && !e_state.Requests[e_state.Floor][elevio.BT_HallUp] {
			e_state.Requests[e_state.Floor][elevio.BT_HallDown] = false
		}
		e_state.Requests[e_state.Floor][elevio.BT_HallUp] = false

	case elevio.MD_Down:
		if !requests_below(e_state) && !e_state.Requests[e_state.Floor][elevio.BT_HallDown] {
			e_state.Requests[e_state.Floor][elevio.BT_HallUp] = false
		}
		e_state.Requests[e_state.Floor][elevio.BT_HallDown] = false

	case elevio.MD_Stop:
		fallthrough
	default:
		e_state.Requests[e_state.Floor][elevio.BT_HallUp] = false
		e_state.Requests[e_state.Floor][elevio.BT_HallDown] = false
	}

	return e_state
}

// --- “static” helpers ---

func requests_above(e_state elevator.ElevatorState) bool {
	if e_state.Floor < 0 || e_state.Floor >= len(e_state.Requests) {
		return false
	}

	for f := e_state.Floor + 1; f < len(e_state.Requests); f++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			if e_state.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e_state elevator.ElevatorState) bool {
	if e_state.Floor < 0 || e_state.Floor >= len(e_state.Requests) {
		return false
	}

	for f := 0; f < e_state.Floor; f++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			if e_state.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e_state elevator.ElevatorState) bool {
	if e_state.Floor < 0 || e_state.Floor >= len(e_state.Requests) {
		return false
	}

	for btn := elevio.ButtonType(0); btn < 3; btn++ {
		if e_state.Requests[e_state.Floor][btn] {
			return true
		}
	}
	return false
}
