package elevatorStateGuardian

import (
	"fmt"
	"theProject/messageSync"
    "theProject/elevatorIo"
)

type Command_t interface{}

//Get types
type GetElevatorData_t struct {
    Reply chan messageSync.ElevatorData_t
}
type GetAssignedRequests_t struct{
    Reply chan elevatorIo.AssignedRequests_t
}

//Set types
type SetSystemData_t struct {
    SystemData messageSync.SystemData_t
}
type SetElevatorData_t struct {
    ElevatorData messageSync.ElevatorData_t
}
type SetIsFunctional_t struct {
    IsFunctional bool
}
type SetFloor_t struct {
	Floor int
}
type SetMotorDirection_t struct {
	MotorDirection elevatorIo.MotorDirection_t
}
type SetElevatorBehaviour_t struct {
	ElevatorBehaviour elevatorIo.ElevatorBehaviour_t
}
type SetRequestsDone_t struct {
    RequestsToClear []elevatorIo.ButtonEvent_t
}
type SetCabRequestDone_t struct {
    Floor int
}
type SetHallRequestDone_t struct {
    Floor int
    Button elevatorIo.ButtonType_t
}
type SetAssignedRequest_t struct {
    AssignedRequests elevatorIo.AssignedRequests_t
}


//routine that owns the local elevator data
//responible for message passing with messageSync, FSM and RA
func ElevatorStateGuardian( 
    commands chan Command_t,                               //channel for using the locally stored system state
    elevatorDataToMsgSync chan<- messageSync.ElevatorData_t,        //channel for sending data to messageSyncServer
    hallRequestToMsgSync chan<- messageSync.RequestCyclicCounter_t, //channel for sending done request CC to msg sync
    localID int) {                                                  //ID of the local elevator 

	requests_temp := make([][]bool, elevatorIo.N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]bool, elevatorIo.N_BUTTONS)
    }

    //Initialize the system data 
    var systemData messageSync.SystemData_t
    var assignedRequests elevatorIo.AssignedRequests_t = requests_temp
	systemData, _ = messageSync.InitSystemData(localID)
    var elevatorDataChanged bool = false
    

	for cmd := range commands {
		switch c := cmd.(type) {

		case GetElevatorData_t:
			c.Reply <- systemData.ElevatorData[localID]
        
		case GetAssignedRequests_t:
			c.Reply <- assignedRequests

        //Used by msg sync to set the new confirmed system data
        case SetSystemData_t:
            systemData = messageSync.DeepCopySystemData(c.SystemData)
        
        //Used by the FSM to set the local elevator state
        case SetElevatorData_t:
            systemData.ElevatorData[localID].IsFunctional       = c.ElevatorData.IsFunctional
            systemData.ElevatorData[localID].Floor              = c.ElevatorData.Floor
            systemData.ElevatorData[localID].ElevatorBehaviour  = c.ElevatorData.ElevatorBehaviour
            systemData.ElevatorData[localID].MotorDirection     = c.ElevatorData.MotorDirection

            systemData.ElevatorData[localID].ElevatorBarrier            = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

        case SetIsFunctional_t:
            systemData.ElevatorData[localID].IsFunctional               = c.IsFunctional
            systemData.ElevatorData[localID].ElevatorBarrier            = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

		case SetFloor_t:
			systemData.ElevatorData[localID].Floor                      = c.Floor
            systemData.ElevatorData[localID].ElevatorBarrier            = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

		case SetMotorDirection_t:
			systemData.ElevatorData[localID].MotorDirection             = c.MotorDirection
            systemData.ElevatorData[localID].ElevatorBarrier            = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

		case SetElevatorBehaviour_t:
			systemData.ElevatorData[localID].ElevatorBehaviour          = c.ElevatorBehaviour
            systemData.ElevatorData[localID].ElevatorBarrier            = make([]bool, messageSync.N_ELEVATORS)
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

        //Settinq requests only need to take the request barrier into account
        case SetRequestsDone_t:
            for _, btnEvnt := range c.RequestsToClear{
                if btnEvnt.Button == elevatorIo.BT_Cab {
                    systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor].Value             = messageSync.CC_Done
                    systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor].Barrier           = make([]bool, messageSync.N_ELEVATORS)
                    systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor].Barrier[localID]  = true 

                    elevatorDataChanged = true 
                } else {
                    systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button].Value             = messageSync.CC_Done
                    systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button].Barrier           = make([]bool, messageSync.N_ELEVATORS)
                    systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button].Barrier[localID]  = true

                    hallRequestToMsgSync <- systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button]
                }
            }

        //The assigned reqests from RA
        case SetAssignedRequest_t:
            assignedRequests = c.AssignedRequests
		}

        //if data in the elevator state was changed by FSM, then we send it to msgSync
        if elevatorDataChanged {
            elevatorDataToMsgSync <- systemData.ElevatorData[localID]
        }
	}
}

//using the get functionallity
func GetElevatorData(commands chan Command_t) messageSync.ElevatorData_t {
    reply := make(chan messageSync.ElevatorData_t)
    commands <- GetElevatorData_t{Reply: reply}
    return <-reply
}

func GetAssignedRequests(commands chan Command_t) elevatorIo.AssignedRequests_t {
    reply := make(chan elevatorIo.AssignedRequests_t)
    commands <- GetAssignedRequests_t{Reply: reply}
    return <-reply
}

//printing and converting functions
func ElevatorBehaviourToString(eb elevatorIo.ElevatorBehaviour_t) string {
    switch eb {
    case elevatorIo.EB_Idle:
        return "idle"
    case elevatorIo.EB_DoorOpen:
        return "doorOpen"
    case elevatorIo.EB_Moving:
        return "moving"
    default:
        return "UNDEFINED"
    }
}

func ElevatorDirnToString(d elevatorIo.MotorDirection_t) string {
    switch d {
    case elevatorIo.MD_Up:
        return "up"
    case elevatorIo.MD_Down:
        return "down"
    case elevatorIo.MD_Stop:
        return "stop"
    default:
        return "UNDEFINED"
    }
}

func ElevatorButtonToString(b elevatorIo.ButtonType_t) string {
    switch b {
    case elevatorIo.BT_HallUp:
        return "B_HallUp"
    case elevatorIo.BT_HallDown:
        return "B_HallDown"
    case elevatorIo.BT_Cab:
        return "B_Cab"
    default:
        return "B_UNDEFINED"
    }
}


func SystemPrint(systemData messageSync.SystemData_t) {
    fmt.Print("  +--------------------------+")
    for i := 0; i < messageSync.N_ELEVATORS; i++{
        
        fmt.Printf("  +----Elevator: %-2d ----+", i)
        fmt.Printf(
            "  |floor = %-2d |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
            systemData.ElevatorData[i].Floor,
            ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection),
            ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour),
        )
        fmt.Println("  +--------------------+")
        fmt.Println("  |  | up  | dn  | cab |")

        for f := elevatorIo.N_FLOORS - 1; f >= 0; f-- {
            fmt.Printf("  | %d", f)

            for btn := elevatorIo.ButtonType_t(0) ; btn < elevatorIo.N_BUTTONS; btn++ {
                if (f == elevatorIo.N_FLOORS-1 && btn == elevatorIo.BT_HallUp) ||
                    (f == 0 && btn == elevatorIo.BT_HallDown) {

                    fmt.Print("|     ")
                } else {
                    if btn == elevatorIo.BT_Cab{
                        if messageSync.CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    } else {
                        if messageSync.CC_ToBool(systemData.HallRequestData[f][btn].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    }
                }
            }
            fmt.Println("|")
        }
        fmt.Println("  +--------------------+")
    }
    fmt.Print("  +--------------------------+")    
}

//print from chatGPT 
func ChatGPT_SystemPrint(systemData messageSync.SystemData_t) {

    // Top line
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Print("  +--------------------+")
    }
    fmt.Println()

    // Elevator headers
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Printf("  | Elevator: %-2d       |", i)
    }
    fmt.Println()

    // Floor
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Printf("  | floor = %-2d         |", systemData.ElevatorData[i].Floor)
    }
    fmt.Println()

    // Direction
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Printf("  | dirn  = %-10s |",
            ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection))
    }
    fmt.Println()

    // Behaviour
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Printf("  | behav = %-10s |",
            ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour))
    }
    fmt.Println()

    // Button header
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Print("  |  | up  | dn  | cab |")
    }
    fmt.Println()

    // Floors
    for f := elevatorIo.N_FLOORS - 1; f >= 0; f-- {

        for i := 0; i < messageSync.N_ELEVATORS; i++ {

            fmt.Printf("  | %d", f)

            for btn := elevatorIo.ButtonType_t(0); btn <elevatorIo.N_BUTTONS; btn++ {

                if (f == elevatorIo.N_FLOORS-1 && btn == elevatorIo.BT_HallUp) ||
                    (f == 0 && btn == elevatorIo.BT_HallDown) {

                    fmt.Print("|     ")

                } else {

                    if btn == elevatorIo.BT_Cab {
                        if messageSync.CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    } else {
                        if messageSync.CC_ToBool(systemData.HallRequestData[f][btn].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    }

                }
            }

            fmt.Print("|")
        }

        fmt.Println()
    }

    // Bottom line
    for i := 0; i < messageSync.N_ELEVATORS; i++ {
        fmt.Print("  +--------------------+")
    }
    fmt.Println()
}
