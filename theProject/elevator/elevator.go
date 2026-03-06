package elevator

import (
	"fmt"
    "time"
    "theProject/messageSync"
)

var N_FLOORS int = 4
var N_BUTTONS ButtonType_t = 3
const N_UP_DOWN int = 2

type ElevatorBehaviour_t int
type AssignedRequests_t [][]bool
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
    IsFunctional       bool
}


type GetState_t struct {
	Reply chan ElevatorState_t
}

type SetState_t struct {
    e_state ElevatorState_t
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

//New channel types 
type GetElevatorState_t struct {
    Reply messageSync.ElevatorData_t
}
type GetAssignedRequests_t struct{
    Reply AssignedRequests_t
}
type SetSystemData_t struct {
    SystemData messageSync.SystemData_t
}
type SetIsFunctnional_t struct {
    IsFunctional bool
}
type SetCabRequestDone_t struct {
    Floor int
}
type SetHallRequestDone_t struct {
    Floor int
    Button ButtonType_t
}
type SetAssignedRequest_t struct {
    AssignedRequests AssignedRequests_t
}



func ElevatorStateGuardian(commands chan Command_t, dataToMsgSync chan<- messageSync.SystemData_t, localID int) {
	requests_temp := make([][]bool, N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]bool, N_BUTTONS)
    }
    /* Old data
	systemData.ElevatorData[localID] := ElevatorState_t{
		Floor: -1,		
		MotorDirection: MD_Stop,
		Requests: requests_temp,
		ElevatorBehaviour: EB_Idle,
		DoorOpenDuration: 3 * time.Second,
	}*/
    //Initialize the system data 
    var systemData messageSync.SystemData_t
	systemData.ElevatorData[localID], _ = messageSync.InitSystemData(localID)
    var systemDataChanged bool = 0

    var assignedRequests Requests_t


	for cmd := range commands {
		switch c := cmd.(type) {

		case GetElevatorState_t:
			c.Reply <- systemData.ElevatorData[localID]
        
		case GetAssignedRequests_t:
			c.Reply <- assignedRequests

        //We don't care about elevator state about other elevators
        //Also ID and IsAlive can't change is the FMS, and USE the Set Hall/Cab Requests to set requests
        case SetSystemData_t:
            systemData.ElevatorData[localID].IsFunctional       = c.SystemData.ElevatorData[localID].IsFunctional
            systemData.ElevatorData[localID].Floor              = c.SystemData.ElevatorData[localID].Floor
            systemData.ElevatorData[localID].ElevatorBehaviour  = c.SystemData.ElevatorData[localID].ElevatorBehaviour
            systemData.ElevatorData[localID].MotorDirection     = c.SystemData.ElevatorData[localID].MotorDirection

            systemData.ElevatorData[localID].ElevatorBarrier = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID] = true
            systemDataChanged = true

            //something to think about: how to seperate when we recieve data from msgSync and we recieve from FSM
            //we dont need to mark as systemDataChanged when we recieve from msgSync and send it back to the msgSync
            //but maybe this is not a ploblem since the data should be the same? but now with empty barriers?
        
        case SetIsFunctnional_t:
            systemData.ElevatorData[localID].IsFunctional = c.IsFunctional
            systemData.ElevatorData[localID].ElevatorBarrier = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID] = true
            systemDataChanged = true

		case SetFloor_t:
			systemData.ElevatorData[localID].Floor = c.Floor
            systemData.ElevatorData[localID].ElevatorBarrier = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID] = true
            systemDataChanged = true

		case SetMotorDirection_t:
			systemData.ElevatorData[localID].MotorDirection = c.MotorDirection
            systemData.ElevatorData[localID].ElevatorBarrier = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID] = true
            systemDataChanged = true

		case SetElevatorBehaviour_t:
			systemData.ElevatorData[localID].ElevatorBehaviour = c.ElevatorBehaviour
            systemData.ElevatorData[localID].ElevatorBarrier = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID] = true
            systemDataChanged = true

        //Settinq requests only need to take the request barrier into account
        case SetCabRequestDone_t:
            systemData.ElevatorData[localID].CabRequests[c.Floor].Value = messageSync.CC_Done
            systemData.ElevatorData[localID].CabRequests[c.Floor].Barrier = make([]bool, messageSync.N_ELEVATORS)
            systemData.ElevatorData[localID].CabRequests[c.Floor].Barrier = true 
            systemDataChanged = true 

        case SetHallRequestDone_t:
            systemData.HallRequestData[c.Floor][c.Button].Value = messageSync.CC_Done
            systemData.HallRequestData[c.Floor][c.Button].Barrier = make([]bool, messageSync.N_ELEVATORS)
            systemData.HallRequestData[c.Floor][c.Button].Barrier = true
            systemDataChanged = true

        //The assigned reqests from RA
        case SetAssignedRequest_t:
            assignedRequests = c.AssignedRequests
		}
            
        if systemDataChanged {
            
        }
	}
}

func GetState(commands chan Command_t) ElevatorState_t {
    reply := make(chan ElevatorState_t)
    commands <- GetState_t{Reply: reply}
    return <-reply
}

func ElevatorBehaviourToString(eb ElevatorBehaviour_t) string {
    switch eb {
    case EB_Idle:
        return "idle"
    case EB_DoorOpen:
        return "doorOpen"
    case EB_Moving:
        return "moving"
    default:
        return "EB_UNDEFINED"
    }
}

func ElevatorDirnToString(d MotorDirection_t) string {
    switch d {
    case MD_Up:
        return "up"
    case MD_Down:
        return "down"
    case MD_Stop:
        return "stop"
    default:
        return "D_UNDEFINED"
    }
}

func ElevatorButtonToString(b ButtonType_t) string {
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


func ElevatorPrint(systemData.ElevatorData[localID] ElevatorState_t) {
    fmt.Println("  +--------------------+")
    fmt.Printf(
        "  |floor = %-2d          |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
        systemData.ElevatorData[localID].Floor,
        ElevatorDirnToString(systemData.ElevatorData[localID].MotorDirection),
        ElevatorBehaviourToString(systemData.ElevatorData[localID].ElevatorBehaviour),
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
                if systemData.ElevatorData[localID].Requests[f][btn] {
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
