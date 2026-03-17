package main

import (
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/messageSync"
	"theProject/fsm"
	"theProject/converters"
	"fmt"
)

/*
-----------------------------------
Functionality: 
	- Run main with the test you want
	- All tests make an instance of old and new data and prints the return value
	- Remeber to either move into messageSync or make update functions public
-----------------------------------
*/

func main(){
	//testUpdateHallRequests()
	//testUpdateElevatorDataAboutSelf()
	//testUpdateElevatorDataAboutOther()
	//testonReceivedFreshData()
}

func testUpdateHallRequests(){
	localID := 0;
	oldHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	fullBarrier := [config.N_ELEVATORS]bool{}
	messageSync.ActivePeers = [config.N_ELEVATORS]bool{}

	newHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
	
			oldHallRequests[floor][btn].Value = messageSync.CC_No
			oldHallRequests[floor][btn].Barrier = [config.N_ELEVATORS]bool(fullBarrier)

			newHallRequests[floor][btn].Value = messageSync.CC_Unconfirmed
			newHallRequests[floor][btn].Barrier = [config.N_ELEVATORS]bool(fullBarrier)
		}
	}

	oldHallRequests = messageSync.updateHallRequestData(oldHallRequests, newHallRequests, localID)

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
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	for floor := 0; floor < config.N_FLOORS; floor++ {
	
			oldData.CabRequests[floor].Value = messageSync.CC_No
			oldData.CabRequests[floor].Barrier = [config.N_ELEVATORS]bool(fullBarrier)

			newData.CabRequests[floor].Value = messageSync.CC_Unconfirmed
			newData.CabRequests[floor].Barrier = [config.N_ELEVATORS]bool(fullBarrier)
	}

	assignedRequests := elevator_IO.AssignedRequests_t{}

	println("OldData:")
	fsm.ElevatorPrint(oldData, assignedRequests)
	oldData = messageSync.updateElevatorDataAboutSelf(oldData, newData, localID)
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
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	for floor := 0; floor < config.N_FLOORS; floor++ {
	
			oldData.CabRequests[floor].Value = messageSync.CC_No
			oldData.CabRequests[floor].Barrier = [config.N_ELEVATORS]bool(fullBarrier)

			newData.CabRequests[floor].Value = messageSync.CC_Unconfirmed
			newData.CabRequests[floor].Barrier = [config.N_ELEVATORS]bool(fullBarrier)
	}

	assignedRequests := elevator_IO.AssignedRequests_t{}

	println("OldData:")
	fsm.ElevatorPrint(oldData, assignedRequests)
	for i := 0; i<3 ; i++{
		print(oldData.ElevatorBarrier[i], " ")
	}
	oldData = messageSync.updateElevatorDataAboutOther(oldData, newData, localID)

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
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	conData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 2,
		MotorDirection: elevator_IO.MD_Down,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	newData := messageSync.ElevatorData_t{
		ID: 0,
		IsAlive: true,
		IsFunctional: true,
		Floor: 3,
		MotorDirection: elevator_IO.MD_Up,
		ElevatorBehaviour: elevator_IO.EB_Moving,
		ElevatorBarrier: [config.N_ELEVATORS]bool{},
		CabRequests: [config.N_FLOORS]messageSync.RequestCyclicCounter_t{},
	}

	oldHallRequests := [config.N_FLOORS][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t{}
	newHallRequests := [config.N_FLOORS][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t{}

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
	
			oldHallRequests[floor][btn].Value = messageSync.CC_No
			oldHallRequests[floor][btn].Barrier = [config.N_ELEVATORS]bool(fullBarrier)

			newHallRequests[floor][btn].Value = messageSync.CC_Confirmed
			newHallRequests[floor][btn].Barrier = [config.N_ELEVATORS]bool(fullBarrier)
		}
	}

	systemData := messageSync.SystemData_t{
		ID: 0,
		ElevatorData: [config.N_ELEVATORS]messageSync.ElevatorData_t{oldData, oldData, oldData},
		HallRequestData: oldHallRequests,
	}

	conFirmedSystemData := messageSync.SystemData_t{
		ID: 0,
		ElevatorData: [config.N_ELEVATORS]messageSync.ElevatorData_t{conData, conData, conData},
		HallRequestData: oldHallRequests,
	}

	freshData := messageSync.SystemData_t{
		ID: 1,
		ElevatorData: [config.N_ELEVATORS]messageSync.ElevatorData_t{newData, newData, newData},
		HallRequestData: newHallRequests,
	}

	systemData, conFirmedSystemData, isUpdated =  messageSync.onReceivedFreshData(systemData, conFirmedSystemData, freshData)

	print(isUpdated)
	println("SYSTEM DATA:")
	SystemPrintHorizotal(systemData)
	println("FRESH DATA:")
	SystemPrintHorizotal(freshData)
	println("CONFIRMED DATA:")
	SystemPrintHorizotal(conFirmedSystemData)
}


func SystemPrintHorizotal(systemData messageSync.SystemData_t) {

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Print("  +--------------------+")
    }
    fmt.Println()

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | Elevator: %-2d       |", i)
    }
    fmt.Println()

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | floor = %-2d         |", systemData.ElevatorData[i].Floor)
    }
    fmt.Println()

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | dirn  = %-10s |",
            converters.ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection))
    }
    fmt.Println()

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Printf("  | behav = %-10s |",
            converters.ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour))
    }
    fmt.Println()

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Print("  |  | up  | dn  | cab |")
    }
    fmt.Println()

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

    for i := 0; i < config.N_ELEVATORS; i++ {
        fmt.Print("  +--------------------+")
    }
    fmt.Println()
}
