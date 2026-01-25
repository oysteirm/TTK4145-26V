package elevator

import (
	"fmt"
    "time"
)

var N_FLOORS int = 4
var N_BUTTONS ButtonType_t = 3

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
    DoorOpenDuration   time.Duration
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

type SetElevatorBehaviour_t struct {
	ElevatorBehaviour ElevatorBehaviour_t
}

func Elevator_Server(commands chan Command_t) {
	requests_temp := make([][]bool, N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]bool, N_BUTTONS)
    }

	e_state := ElevatorState_t{
		Floor: -1,		
		MotorDirection: MD_Stop,
		Requests: requests_temp,
		ElevatorBehaviour: EB_Idle,
		DoorOpenDuration: 3 * time.Second,
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
			e_state.Floor = c.Floor
		case SetMotorDirection_t:
			e_state.MotorDirection = c.MotorDirection
		case SetRequest_t:
			e_state.Requests[c.Floor][c.Button] = c.RequestValue
		case SetElevatorBehaviour_t:
			e_state.ElevatorBehaviour = c.ElevatorBehaviour
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
    case MD_Up:
        return "D_Up"
    case MD_Down:
        return "D_Down"
    case MD_Stop:
        return "D_Stop"
    default:
        return "D_UNDEFINED"
    }
}

func elevator_button_to_string(b ButtonType_t) string {
    switch b {
    case BT_HallUp:
        return "B_HallUp"
    case BT_HallDown:
        return "B_HallDown"
    case BT_Cab:
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

        for btn := ButtonType_t(0) ; btn < N_BUTTONS; btn++ {
            if (f == N_FLOORS-1 && btn == BT_HallUp) ||
                (f == 0 && btn == BT_HallDown) {

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
