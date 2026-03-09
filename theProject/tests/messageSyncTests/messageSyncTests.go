package main

import (
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/messageSync"
	"theProject/fsm"
	"theProject/converters"
	"fmt"
)


func main(){
	//testUpdateHallRequests()
	//testUpdateElevatorDataAboutSelf()
	//testUpdateElevatorDataAboutOther()
	testonReceivedFreshData()
}

func testUpdateHallRequests(){
	localID := 0;
	oldHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	fullBarrier := []bool{true, true, true}
	messageSync.ActivePeers = [config.N_ELEVATORS]bool{true, true, true}

	newHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
	
			oldHallRequests[floor][btn].Value = messageSync.CC_No
			oldHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)

			newHallRequests[floor][btn].Value = messageSync.CC_Unconfirmed
			newHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)
		}
	}

	oldHallRequests = messageSync.UpdateHallRequestData(oldHallRequests, newHallRequests, localID)

	println("Active Peers:")
	for i := 0; i<3 ; i++{
		print(messageSync.ActivePeers[i], " ")
	} 
	println()
	
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			println(oldHallRequests[floor][btn].Value)
			for i := 0; i<3 ; i++{
				print(oldHallRequests[floor][btn].Barrier[i], " ")
			}
			println()
		}
	}
}

func testUpdateElevatorDataAboutSelf(){

	localID := 0
	fullBarrier := []bool{true, true, true}
	messageSync.ActivePeers = [config.N_ELEVATORS]bool{true, true, true}

	oldData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: false,
		IsFunctional: false,
		Floor: 2,
		MotorDirection: elevator_IO.MD_Stop,
		ElevatorBehaviour: elevator_IO.EB_Idle,
		ElevatorBarrier: []bool{true, false, true},
		CabRequests: make([]messageSync.RequestCyclicCounter_t, config.N_FLOORS),
	}

	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: []bool{true, false, false},
		CabRequests: make([]messageSync.RequestCyclicCounter_t, config.N_FLOORS),
	}

	for floor := 0; floor < config.N_FLOORS; floor++ {
	
			oldData.CabRequests[floor].Value = messageSync.CC_No
			oldData.CabRequests[floor].Barrier = messageSync.DeepCopyBarrier(fullBarrier)

			newData.CabRequests[floor].Value = messageSync.CC_Unconfirmed
			newData.CabRequests[floor].Barrier = messageSync.DeepCopyBarrier(fullBarrier)
	}

	assignedRequests := make([][]bool, elevator_IO.N_FLOORS)
    for i := range assignedRequests {
        assignedRequests[i] = make([]bool, elevator_IO.N_BUTTONS)
    }

	println("OldData:")
	fsm.ElevatorPrint(oldData, assignedRequests)
	oldData = messageSync.UpdateElevatorDataAboutSelf(oldData, newData, localID)
	println("UpdatedData:")
	fsm.ElevatorPrint(oldData, assignedRequests)

}

func testUpdateElevatorDataAboutOther(){
	localID := 0
	fullBarrier := []bool{true, true, true}
	messageSync.ActivePeers = [config.N_ELEVATORS]bool{true, true, true}

	oldData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: false,
		IsFunctional: false,
		Floor: 2,
		MotorDirection: elevator_IO.MD_Stop,
		ElevatorBehaviour: elevator_IO.EB_Idle,
		ElevatorBarrier: []bool{true, true, false},
		CabRequests: make([]messageSync.RequestCyclicCounter_t, config.N_FLOORS),
	}

	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: []bool{true, false, true},
		CabRequests: make([]messageSync.RequestCyclicCounter_t, config.N_FLOORS),
	}

	for floor := 0; floor < config.N_FLOORS; floor++ {
	
			oldData.CabRequests[floor].Value = messageSync.CC_No
			oldData.CabRequests[floor].Barrier = messageSync.DeepCopyBarrier(fullBarrier)

			newData.CabRequests[floor].Value = messageSync.CC_Unconfirmed
			newData.CabRequests[floor].Barrier = messageSync.DeepCopyBarrier(fullBarrier)
	}

	assignedRequests := make([][]bool, elevator_IO.N_FLOORS)
    for i := range assignedRequests {
        assignedRequests[i] = make([]bool, elevator_IO.N_BUTTONS)
    }

	println("OldData:")
	fsm.ElevatorPrint(oldData, assignedRequests)
	for i := 0; i<3 ; i++{
		print(oldData.ElevatorBarrier[i], " ")
	}
	oldData = messageSync.UpdateElevatorDataAboutOther(oldData, newData, localID)

	println("UpdatedData:")
	fsm.ElevatorPrint(oldData, assignedRequests)
	for i := 0; i<3 ; i++{
		print(oldData.ElevatorBarrier[i], " ")
	}

}

func testonReceivedFreshData(){
	fullBarrier := []bool{true, true, true}
	messageSync.ActivePeers = [config.N_ELEVATORS]bool{true, true, true}
	isUpdated := false
	
	oldData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: false,
		IsFunctional: false,
		Floor: 2,
		MotorDirection: elevator_IO.MD_Stop,
		ElevatorBehaviour: elevator_IO.EB_Idle,
		ElevatorBarrier: []bool{true, false, true},
		CabRequests: make([]messageSync.RequestCyclicCounter_t, config.N_FLOORS),
	}

	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: []bool{true, false, false},
		CabRequests: make([]messageSync.RequestCyclicCounter_t, config.N_FLOORS),
	}

	oldHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)
	newHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
	
			oldHallRequests[floor][btn].Value = messageSync.CC_No
			oldHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)

			newHallRequests[floor][btn].Value = messageSync.CC_Unconfirmed
			newHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)
		}
	}

	systemData := messageSync.SystemData_t{
		ID: 0,
		ElevatorData: []messageSync.ElevatorData_t{oldData, oldData, oldData},
		HallRequestData: oldHallRequests,
	}

	conFirmedSystemData := messageSync.SystemData_t{
		ID: 0,
		ElevatorData: []messageSync.ElevatorData_t{oldData, oldData, oldData},
		HallRequestData: oldHallRequests,
	}

	freshData := messageSync.SystemData_t{
		ID: 0,
		ElevatorData: []messageSync.ElevatorData_t{newData, newData, newData},
		HallRequestData: newHallRequests,
	}

	systemData, conFirmedSystemData, isUpdated =  messageSync.OnReceivedFreshData(systemData, conFirmedSystemData, freshData)

	print(isUpdated)
	println("SYSTEM DATA:")
	ChatGPT_SystemPrint(systemData)
	println("FRESH DATA:")
	ChatGPT_SystemPrint(freshData)
	println("CONFIRMED DATA:")
	ChatGPT_SystemPrint(conFirmedSystemData)
}



//printing functoins
func SystemPrint(systemData messageSync.SystemData_t) {
    fmt.Println("  +--------------------------+")
    for i := 0; i < config.N_ELEVATORS; i++{
        
        fmt.Printf("  +----Elevator: %-2d \n", i)
        fmt.Printf(
            "  |floor = %-2d |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
            systemData.ElevatorData[i].Floor,
            converters.ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection),
            converters.ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour),
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
                        if converters.CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    } else {
                        if converters.CC_ToBool(systemData.HallRequestData[f][btn].Value) {
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
    fmt.Println("  +--------------------------+")    
}

//print from chatGPT 
func ChatGPT_SystemPrint(systemData messageSync.SystemData_t) {

    // Top line
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Print("  +--------------------+")
    }
    fmt.Println()

    // Elevator headers
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | Elevator: %-2d       |", i)
    }
    fmt.Println()

    // Floor
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | floor = %-2d         |", systemData.ElevatorData[i].Floor)
    }
    fmt.Println()

    // Direction
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | dirn  = %-10s |",
            converters.ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection))
    }
    fmt.Println()

    // Behaviour
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | behav = %-10s |",
            converters.ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour))
    }
    fmt.Println()

    // Button header
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Print("  |  | up  | dn  | cab |")
    }
    fmt.Println()

    // Floors
    for f := elevator_IO.N_FLOORS - 1; f >= 0; f-- {

        for i := 0; i < config.N_ELEVATORS; i++ {

            fmt.Printf("  | %d", f)

            for btn := elevator_IO.ButtonType_t(0); btn <elevator_IO.N_BUTTONS; btn++ {

                if (f == elevator_IO.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
                    (f == 0 && btn == elevator_IO.BT_HallDown) {

                    fmt.Print("|     ")

                } else {

                    if btn == elevator_IO.BT_Cab {
                        if converters.CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
                            fmt.Print("|  #  ")
                        } else {
                            fmt.Print("|  -  ")
                        }
                    } else {
                        if converters.CC_ToBool(systemData.HallRequestData[f][btn].Value) {
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
    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Print("  +--------------------+")
    }
    fmt.Println()
}
