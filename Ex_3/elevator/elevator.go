package elevator

import (
	"fmt"
)

N_FLOORS := 4
N_BUTTONS := 3

type ElevatorBehaviour int
type Requests [][]bool
type Command interface{}

const (
    EB_Idle ElevatorBehaviour = 0
    EB_DoorOpen               = 1  
    EB_Moving                 = 2
)

type ElevatorState struct {
    Floor              int
    MotorDirection     MotorDirection
    Requests           Requests
    ElevatorBehaviour  ElevatorBehaviour
    DoorOpenDuration   float64
}

type GetState struct {
	Reply chan ElevatorState
}

type SetState struct {
    ElevatorState ElevatorState
}

type SetFloor struct {
	Floor int
}

type SetMotorDirection struct {
	MotorDirection MotorDirection
}

type SetRequest struct {
	RequestValue bool //must be changed to a Request type later
    Floor int
    Button ButtonType
}

type SetElevatorBehavior struct {
	Behaviour ElevatorBehaviour
}

func Elevator_Server(commands chan Command) {
	requests_temp := make([][]bool, N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]bool, N_BUTTONS)
    }

	e_state := ElevatorState{
		floor: -1,		
		motor_direction: MD_Stop,
		requests: requests_temp,
		behaviour: EB_Idle,
		door_open_duration: 3,
	}

	for cmd := range commands {
		switch c := cmd.(type) {

		case GetState:
			c.Reply <- e_state
        case SetState:
            e_state.Floor = c.ElevatorState.Floor
            e_state.MotorDirection = c.ElevatorState.MotorDirection
            e_state.Requests = c.ElevatorState.Requests
            e_state.ElevatorBehaviour = c.ElevatorState.ElevatorBehaviour
            e_state.DoorOpenDuration = c.ElevatorState.DoorOpenDuration
		case SetFloor:
			e_state.floor = c.Floor
		case SetMotorDirection:
			e_state.motor_direction = c.MotorDirection
		case SetRequest:
			e_state.requests[c.Floor][c.Button] = c.RequestValue
		case SetElevatorBehavior:
			e_state.behaviour = c.Behaviour
		}
	}
}

func GetState(commands chan Command) ElevatorState {
    reply := make(chan ElevatorState)
    commands <- GetState{Reply: reply}
    return <-reply
}

func elevator_behaviour_to_string(eb ElevatorBehaviour) string {
    switch eb {
    case EB_Idle:
        return "EB_Idle"
    case EB_DoorOpen:
        return "EB_DoorOpen"
    case EB_Moving:
        return "EB_Moving"
    default:
        return "EB_UNDEFINED"
    }
}

func elevator_dirn_to_string(d Dirn) string {
    switch d {
    case D_Up:
        return "D_Up"
    case D_Down:
        return "D_Down"
    case D_Stop:
        return "D_Stop"
    default:
        return "D_UNDEFINED"
    }
}

func elevator_button_to_string(b Button) string {
    switch b {
    case B_HallUp:
        return "B_HallUp"
    case B_HallDown:
        return "B_HallDown"
    case B_Cab:
        return "B_Cab"
    default:
        return "B_UNDEFINED"
    }
}


func elevator_print(e_state ElevatorState) {
    fmt.Println("  +--------------------+")
    fmt.Printf(
        "  |floor = %-2d          |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
        e_state.Floor,
        elevator_dirn_to_string(e_state.MotorDirection),
        elevator_behaviour_to_string(e_state.ElevatorBehaviour),
    )
    fmt.Println("  +--------------------+")
    fmt.Println("  |  | up  | dn  | cab |")

    for f := N_FLOORS - 1; f >= 0; f-- {
        fmt.Printf("  | %d", f)

        for btn := 0; btn < N_BUTTONS; btn++ {
            if (f == N_FLOORS-1 && btn == B_HallUp) ||
                (f == 0 && btn == B_HallDown) {

                fmt.Print("|     ")
            } else {
                if e_state.Requests[f][btn] {
                    fmt.Print("|  #  ")
                } else {
                    fmt.Print("|  -  ")
                }
            }
        }
        fmt.Println("|")
    }

    fmt.Println("  +--------------------+")
}
