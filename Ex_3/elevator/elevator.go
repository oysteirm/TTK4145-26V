package elevator

import (
	"fmt"
)

N_FLOORS := 4
N_BUTTONS := 3

type ElevatorBehaviour_t int
type Requests_t [][]bool
type Command_t interface{}

const (
    EB_Idle ElevatorBehaviour_t = 0
    EB_DoorOpen               = 1  
    EB_Moving                 = 2
)

type ElevatorState_t struct {
    Floor              int
    MotorDirection     MotorDirection_t
    Requests           Requests_t
    ElevatorBehaviour  ElevatorBehaviour_t
    DoorOpenDuration   float64
}

type GetState_t struct {
	Reply chan ElevatorState_t
}

type SetState_t struct {
    ElevatorState ElevatorState_t
}

type SetFloor_t struct {
	Floor int
}

type SetMotorDirection_t struct {
	MotorDirection MotorDirection_t
}

type SetRequest_t struct {
	RequestValue bool //must be changed to a Request type later
    Floor int
    Button ButtonType_t
}

type SetElevatorBehavior_t struct {
	Behaviour ElevatorBehaviour_t
}

func Elevator_Server(commands chan Command_t) {
	requests_temp := make([][]bool, N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]bool, N_BUTTONS)
    }

	e_state := ElevatorState_t{
		floor: -1,		
		motor_direction: MD_Stop,
		requests: requests_temp,
		behaviour: EB_Idle,
		door_open_duration: 3,
	}

	for cmd := range commands {
		switch c := cmd.(type) {

		case GetState_t:
			c.Reply <- e_state
        case SetState_t:
            e_state.Floor = c.ElevatorState.Floor
            e_state.MotorDirection = c.ElevatorState.MotorDirection
            e_state.Requests = c.ElevatorState.Requests
            e_state.ElevatorBehaviour = c.ElevatorState.ElevatorBehaviour
            e_state.DoorOpenDuration = c.ElevatorState.DoorOpenDuration
		case SetFloor_t:
			e_state.floor = c.Floor
		case SetMotorDirection_t:
			e_state.motor_direction = c.MotorDirection
		case SetRequest_t:
			e_state.requests[c.Floor][c.Button] = c.RequestValue
		case SetElevatorBehavior_t:
			e_state.behaviour = c.Behaviour
		}
	}
}

func GetState(commands chan Command_t) ElevatorState_t {
    reply := make(chan ElevatorState_t)
    commands <- GetState_t{Reply: reply}
    return <-reply
}

func elevator_behaviour_to_string(eb ElevatorBehaviour_t) string {
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

func elevator_dirn_to_string(d MotorDirection_t) string {
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

func elevator_button_to_string(b ButtonType_t) string {
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


func elevator_print(e_state ElevatorState_t) {
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
