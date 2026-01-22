package elevator

import (
	"fmt"
	"project/fsm"
)

type ElevatorBehaviour int
type Requests [][]int
type Command interface{}

const (
    EB_Idle ElevatorBehaviour = iota
    EB_DoorOpen
    EB_Moving
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

type SetFloor struct {
	Floor int
}

type SetMotorDirection struct {
	MotorDirection MotorDirection
}

type SetRequests struct {
	Requests Requests
}

type SetElevatorBehavior struct {
	Behaviour ElevatorBehaviour
}

func Elevator_Server(commands chan Command) {
	requests_temp := make([][]int, N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]int, N_BUTTONS)
    }

	elevator_state := ElevatorState{
		floor: -1,		
		motor_direction: MD_Stop,
		requests: requests_temp,
		behaviour: EB_Idle,
		door_open_duration: 3,
	}

	for cmd := range commands {
		switch c := cmd.(type) {

		case GetState:
			c.Reply <- elevator_state
		case SetFloor:
			elevator_state.floor = c.Floor
		case SetMotorDirection:
			elevator_state.motor_direction = c.MotorDirection
		case SetRequests:
			elevator_state.requests = c.Requests
		case SetElevatorBehavior:
			elevator_state.behaviour = c.Behaviour
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


func elevator_print(es Elevator) {
    fmt.Println("  +--------------------+")
    fmt.Printf(
        "  |floor = %-2d          |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
        es.Floor,
        elevator_dirn_to_string(es.Dirn),
        elevator_behaviour_to_string(es.Behaviour),
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
                if es.Requests[f][btn] {
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

//intialze a elevator
func elevator_uninitialized(void) Elevator {
    elevator_hardware_init();
    return (Elevator){ //must be rewritten
        .floor = -1,
        .dirn = D_Stop,
        .behaviour = EB_Idle,
        .config = {
        .doorOpenDuration_s = 3.0,
        },
    };
}


