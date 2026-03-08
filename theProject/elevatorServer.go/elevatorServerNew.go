package elevatorServer

import (
	"fmt"	
	"time"
	"strconv"
	"theProject/elevatorStateGuardian"
	"theProject/elevator_IO"
	"theProject/fsm"
	"theProject/messageSync"
	"theProject/requestAssigner"
	"theProject/timer"
	"theProject/config"
)


func ElevatorServer(
	elevatorDataToMsgSync chan<- messageSync.ElevatorData_t,    //channel for sending data to messageSyncServer
    requestToMsgSync chan<- messageSync.RequestCyclicCounter_t,	//channel for sending done request CC to msg sync
	systemDataFromMsgSync <-chan messageSync.SystemData_t,		//channel for receiving confirmed system data
	localID int,												//local ID
){
	//Connect
    elevator_IO.Init("localhost:15657", config.N_FLOORS)

    guardianCommands := make(chan elevatorStateGuardian.GuardianCommands_t)

	// Start elevator state server
	go elevatorStateGuardian.ElevatorStateGuardian(guardianCommands, elevatorDataToMsgSync, requestToMsgSync, localID)
    
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevator_IO.PollFloorSensor(drv_floors)
    go elevator_IO.PollObstructionSwitch(drv_obstr)
    go elevator_IO.PollStopButton(drv_stop)

    // Init FSM (handle between floors)
	fsm.OnInitBetweenFloors(guardianCommands)

	// Timers
	doorTimerStart    := make(chan time.Duration)
	doorTimerStop     := make(chan struct{})
	doorTimerTimeout  := make(chan struct{})
	obstruction       := make(chan struct{})
	inactiveStart     := make(chan struct{})
	inactiveStop      := make(chan struct{})
	setFunctional     := make(chan bool)

	//work in progress
	go timer.Timers(doorTimerStart, doorTimerStop, doorTimerTimeout, obstruction, inactiveStart, inactiveStop, setFunctional)
    
    for {
		select {

		// Recieved data from msg sync
		case newSystemData := <-systemDataFromMsgSync:

			//Use the RA
			requestsMap := requestAssigner.AssignRequests(requestAssigner.Generating_RA_SystemData(newSystemData))

			assignedRequests := requestsMap[strconv.Itoa(localID)]

			//Store requests and the confirmed system data in Guardian
			guardianCommands <- newSystemData
			guardianCommands <- assignedRequests

			fsm.LightCabLights(newSystemData.ElevatorData[localID].CabRequests)
			fsm.LightHallLights(newSystemData.HallRequestData)

			fsm.OnReceivedDataFromMsgSync(guardianCommands, doorTimerStart, doorTimerStop)

		// Floor arrival
		case floor := <-drv_floors:
			fsm.OnFloorArrival(guardianCommands, doorTimerStart, doorTimerStop, inactiveStart, inactiveStop, floor)
			//TODO: add functionallity for updating IsFunctional

		// Door timeout
		case <-doorTimerTimeout:
			fsm.OnDoorTimeout(guardianCommands, doorTimerStart, doorTimerStop, inactiveStart, inactiveStop)
			//TODO: fix the double requests issue
			
		// Stop button
		case stop := <-drv_stop:
			if stop {
				elevator_IO.SetStopLamp(true)
			} else {
				elevator_IO.SetStopLamp(false)
			}

		// Obstruction
		case obstructed := <-drv_obstr:
			if obstructed {
				doorTimerStop <- struct{}{}
			} else {
				doorTimerStop <- struct{}{}
				doorTimerStart <- config.DOOR_OPEN_DURATION //USE CONST!!!!
			}
		}
	}    
} 

//printing and converting functions
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

        for f := elevator_IO.N_FLOORS - 1; f >= 0; f-- {
            fmt.Printf("  | %d", f)

            for btn := elevator_IO.ButtonType_t(0) ; btn < elevator_IO.N_BUTTONS; btn++ {
                if (f == elevator_IO.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
                    (f == 0 && btn == elevator_IO.BT_HallDown) {

                    fmt.Print("|     ")
                } else {
                    if btn == elevator_IO.BT_Cab{
                        if fsm.CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    } else {
                        if fsm.CC_ToBool(systemData.HallRequestData[f][btn].Value) {
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
    for f := elevator_IO.N_FLOORS - 1; f >= 0; f-- {

        for i := 0; i < messageSync.N_ELEVATORS; i++ {

            fmt.Printf("  | %d", f)

            for btn := elevator_IO.ButtonType_t(0); btn <elevator_IO.N_BUTTONS; btn++ {

                if (f == elevator_IO.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
                    (f == 0 && btn == elevator_IO.BT_HallDown) {

                    fmt.Print("|     ")

                } else {

                    if btn == elevator_IO.BT_Cab {
                        if fsm.CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    } else {
                        if fsm.CC_ToBool(systemData.HallRequestData[f][btn].Value) {
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
