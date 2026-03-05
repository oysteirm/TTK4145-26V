package elevator

//import ()

// behavour-strukt, for retning og tilstand
type MotorDirectionBehaviourPair_t struct {
	MotorDirection    MotorDirection_t
	ElevatorBehaviour ElevatorBehaviour_t
}

// requests_choose_direction tilsvarer: requests_chooseDirection(ElevatorState e_state)


func requests_choose_direction(e_state ElevatorState_t) MotorDirectionBehaviourPair_t {
	switch e_state.MotorDirection {
	case MD_Up:
		if requests_above(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Up, EB_Moving}
		} else if requests_here(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Down, EB_DoorOpen}
		} else if requests_below(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Down, EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{MD_Stop, EB_Idle}
		}

	case MD_Down:
		if requests_below(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Down, EB_Moving}
		} else if requests_here(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Up, EB_DoorOpen}
		} else if requests_above(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Up, EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{MD_Stop, EB_Idle}
		}

	case MD_Stop:
		if requests_here(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Stop, EB_DoorOpen}
		} else if requests_above(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Up, EB_Moving}
		} else if requests_below(e_state) {
			return MotorDirectionBehaviourPair_t{MD_Down, EB_Moving}
		} else {
			return MotorDirectionBehaviourPair_t{MD_Stop, EB_Idle}
		}

	default:
		return MotorDirectionBehaviourPair_t{MD_Stop, EB_Idle}
	}
}

func requests_should_stop(e_state ElevatorState_t) bool {
	// Requests [][]bool:
	if e_state.Floor < 0 || e_state.Floor >= len(e_state.Requests) {
		return false
	}

	switch e_state.MotorDirection {
	case MD_Down:
		return e_state.Requests[e_state.Floor][BT_HallDown] ||
			e_state.Requests[e_state.Floor][BT_Cab] ||
			!requests_below(e_state)

	case MD_Up:
		return e_state.Requests[e_state.Floor][BT_HallUp] ||
			e_state.Requests[e_state.Floor][BT_Cab] ||
			!requests_above(e_state)

	case MD_Stop:
		fallthrough
	default:
		return true
	}
}

func requests_should_clear_immediately(e_state ElevatorState_t, btnFloor int, btnType ButtonType_t) bool {
	return e_state.Floor == btnFloor &&
		((e_state.MotorDirection == MD_Up && btnType == BT_HallUp) ||
			(e_state.MotorDirection == MD_Down && btnType == BT_HallDown) ||
			e_state.MotorDirection == MD_Stop ||
			btnType == BT_Cab)
}

//lager en SetState i elevator, så benytter ikke commands her!
func requests_clear_at_current_floor(e_state ElevatorState_t) ElevatorState_t {
	
	e_state.Requests[e_state.Floor][BT_Cab] = false

	switch e_state.MotorDirection {
	case MD_Up:
		if !requests_above(e_state) && !e_state.Requests[e_state.Floor][BT_HallUp] {
			e_state.Requests[e_state.Floor][BT_HallDown] = false
		}
		e_state.Requests[e_state.Floor][BT_HallUp] = false

	case MD_Down:
		if !requests_below(e_state) && !e_state.Requests[e_state.Floor][BT_HallDown] {
			e_state.Requests[e_state.Floor][BT_HallUp] = false
		}
		e_state.Requests[e_state.Floor][BT_HallDown] = false

	case MD_Stop:
		fallthrough
	default:
		e_state.Requests[e_state.Floor][BT_HallUp] = false
		e_state.Requests[e_state.Floor][BT_HallDown] = false
	}

	return e_state
}

// --- “static” helpers ---

func requests_above(e_state ElevatorState_t) bool {
	for f := e_state.Floor + 1; f < N_FLOORS; f++ {
		for btn := ButtonType_t(0); btn < N_BUTTONS; btn++ {
			if e_state.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e_state ElevatorState_t) bool {
	for f := 0; f < e_state.Floor; f++ {
		for btn := ButtonType_t(0); btn < N_BUTTONS; btn++ {
			if e_state.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e_state ElevatorState_t) bool {
	for btn := ButtonType_t(0); btn < N_BUTTONS; btn++ {
		if e_state.Requests[e_state.Floor][btn] {
			return true
		}
	}
	return false
}
