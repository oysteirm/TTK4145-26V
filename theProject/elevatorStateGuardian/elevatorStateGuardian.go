package elevatorStateGuardian

import (
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/messageSync"
    "fmt"
    "theProject/converters"
)

type GuardianCommands_t interface{}

//Get types
type GetElevatorData_t struct {
    Reply chan messageSync.ElevatorData_t
}
type GetAssignedRequests_t struct{
    Reply chan elevator_IO.AssignedRequests_t
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
	MotorDirection elevator_IO.MotorDirection_t
}
type SetElevatorBehaviour_t struct {
	ElevatorBehaviour elevator_IO.ElevatorBehaviour_t
}
type SetRequestsDone_t struct {
    RequestsToClear []elevator_IO.ButtonEvent_t
}
type SetCabRequestDone_t struct {
    Floor int
}
type SetHallRequestDone_t struct {
    Floor int
    Button elevator_IO.ButtonType_t
}
type SetAssignedRequest_t struct {
    AssignedRequests elevator_IO.AssignedRequests_t
}


//routine that owns the local elevator data
//responible for message passing with messageSync, FSM and RA
func ElevatorStateGuardian( 
    guardianCommands chan GuardianCommands_t,                               //channel for using the locally stored system state
    elevatorDataToMsgSync chan<- messageSync.ElevatorData_t,        //channel for sending data to messageSyncServer
    requestsToMsgSync chan<- []elevator_IO.ButtonEvent_t, //channel for sending done request CC to msg sync
    localID int) {                                                  //ID of the local elevator 

	requests_temp := make([][]bool, elevator_IO.N_FLOORS)
    for i := range requests_temp {
        requests_temp[i] = make([]bool, elevator_IO.N_BUTTONS)
    }

    //Initialize the system data 
    var systemData messageSync.SystemData_t
    var assignedRequests elevator_IO.AssignedRequests_t = requests_temp
	systemData, _ = messageSync.InitSystemData(localID)
    var elevatorDataChanged bool = false
    

	for cmd := range guardianCommands {
		switch c := cmd.(type) {

		case GetElevatorData_t:
			c.Reply <- systemData.ElevatorData[localID]
        
		case GetAssignedRequests_t:
			c.Reply <- assignedRequests

        //Used by msg sync to set the new confirmed system data
        case SetSystemData_t:
            systemData = c.SystemData
        
        //Used by the FSM to set the local elevator state
        case SetElevatorData_t:
            systemData.ElevatorData[localID].IsFunctional       = c.ElevatorData.IsFunctional
            systemData.ElevatorData[localID].Floor              = c.ElevatorData.Floor
            systemData.ElevatorData[localID].ElevatorBehaviour  = c.ElevatorData.ElevatorBehaviour
            systemData.ElevatorData[localID].MotorDirection     = c.ElevatorData.MotorDirection

            systemData.ElevatorData[localID].ElevatorBarrier            = [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

        case SetIsFunctional_t:
            systemData.ElevatorData[localID].IsFunctional               = c.IsFunctional
            systemData.ElevatorData[localID].ElevatorBarrier            = [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

		case SetFloor_t:
			systemData.ElevatorData[localID].Floor                      = c.Floor
            systemData.ElevatorData[localID].ElevatorBarrier            = [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

		case SetMotorDirection_t:
			systemData.ElevatorData[localID].MotorDirection             = c.MotorDirection
            systemData.ElevatorData[localID].ElevatorBarrier            = [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

		case SetElevatorBehaviour_t:
			systemData.ElevatorData[localID].ElevatorBehaviour          = c.ElevatorBehaviour
            systemData.ElevatorData[localID].ElevatorBarrier            = [config.N_ELEVATORS]bool{}
			systemData.ElevatorData[localID].ElevatorBarrier[localID]   = true
            elevatorDataChanged = true

        //Settinq requests only need to take the request barrier into account
        case SetRequestsDone_t:
            requestsToMsgSync <- c.RequestsToClear
            // for _, btnEvnt := range c.RequestsToClear{
            //     if btnEvnt.Button == elevator_IO.BT_Cab {
            //         systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor].Value             = messageSync.CC_Done
            //         systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor].Barrier           = make([]bool, messageSync.N_ELEVATORS)
            //         systemData.ElevatorData[localID].CabRequests[btnEvnt.Floor].Barrier[localID]  = true 

            //         elevatorDataChanged = true 
            //     } else {
            //         systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button].Value             = messageSync.CC_Done
            //         systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button].Barrier           = make([]bool, messageSync.N_ELEVATORS)
            //         systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button].Barrier[localID]  = true

            //         requestToMsgSync <- systemData.HallRequestData[btnEvnt.Floor][btnEvnt.Button]
            //     }
            // }

        //The assigned reqests from RA
        case SetAssignedRequest_t:
            assignedRequests = c.AssignedRequests
		}

        //if data in the elevator state was changed by FSM, then we send it to msgSync
        if elevatorDataChanged {
            elevatorDataToMsgSync <- systemData.ElevatorData[localID]
            elevatorDataChanged = false
            // println("Sending data to msg sync from FSM")
            // ElevatorPrint(systemData.ElevatorData[localID], assignedRequests)
        }
	}
}

//using the get functionallity
func GetElevatorData(guardianCommands chan GuardianCommands_t) messageSync.ElevatorData_t {
    reply := make(chan messageSync.ElevatorData_t)
    guardianCommands <- GetElevatorData_t{Reply: reply}
    return <-reply
}

func GetAssignedRequests(guardianCommands chan GuardianCommands_t) elevator_IO.AssignedRequests_t {
    reply := make(chan elevator_IO.AssignedRequests_t)
    guardianCommands <- GetAssignedRequests_t{Reply: reply}
    return <-reply
}


//printing functoins
func ElevatorPrint(elevator messageSync.ElevatorData_t, assignedRequests elevator_IO.AssignedRequests_t) {

    fmt.Printf("  +--------------------+\n")
    fmt.Printf(
        "  |IsAlive = %-9t |\n"+
        "  |IsFunctional = %-2t |\n"+
        "  |floor = %-11d |\n"+
        "  |dirn  = %-12s|\n"+
        "  |behav = %-12s|\n",
        elevator.IsAlive,
        elevator.IsFunctional,
        elevator.Floor,
        converters.ElevatorDirnToString(elevator.MotorDirection),
        converters.ElevatorBehaviourToString(elevator.ElevatorBehaviour),
    )
    fmt.Println("  +--------------------+")
    fmt.Println("  |  | up  | dn  | cab |")

    for f := elevator_IO.N_FLOORS - 1; f >= 0; f-- {
        fmt.Printf("  | %d", f)

        for btn := elevator_IO.ButtonType_t(0) ; btn < elevator_IO.N_BUTTONS; btn++ {
            if (f == elevator_IO.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
                (f == 0 && btn == elevator_IO.BT_HallDown) {

                fmt.Print("|     ")
            } else {
                if btn == elevator_IO.BT_Cab{
                    if converters.CC_ToBool(elevator.CabRequests[f].Value) {
                        fmt.Print("|  #  ")
                    } else {
                        fmt.Print("|  -  ")
                    }
                } else {
                    if assignedRequests[f][btn] {
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


