package converters
//import something to use System_Data_t??
import (
	"theProject/messageSync"
	"theProject/elevator_IO"
)

func CC_ToBool(cc messageSync.CyclicCounter_t) bool {
	switch cc {
	case messageSync.CC_Confirmed, messageSync.CC_Done:
		return true
	default:
		return false
	}
}

func ElevatorBehaviourToString(eb elevator_IO.ElevatorBehaviour_t) string {
    switch eb {
    case elevator_IO.EB_Idle:
        return "idle"
    case elevator_IO.EB_DoorOpen:
        return "doorOpen"
    case elevator_IO.EB_Moving:
        return "moving"
    default:
        return "UNDEFINED"
    }
}

func ElevatorDirnToString(d elevator_IO.MotorDirection_t) string {
    switch d {
    case elevator_IO.MD_Up:
        return "up"
    case elevator_IO.MD_Down:
        return "down"
    case elevator_IO.MD_Stop:
        return "stop"
    default:
        return "UNDEFINED"
    }
}

func ElevatorButtonToString(b elevator_IO.ButtonType_t) string {
    switch b {
    case elevator_IO.BT_HallUp:
        return "B_HallUp"
    case elevator_IO.BT_HallDown:
        return "B_HallDown"
    case elevator_IO.BT_Cab:
        return "B_Cab"
    default:
        return "B_UNDEFINED"
    }
}