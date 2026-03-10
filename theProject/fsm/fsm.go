package fsm

import (
	"fmt"
	"theProject/config"
	"theProject/elevatorStateGuardian"
	"theProject/elevator_IO"
	"theProject/messageSync"
	"theProject/requests"
    "theProject/converters"
	
)

//Light functions using cyclic counter values
func LightCabLights(CabRequests [config.N_FLOORS]messageSync.RequestCyclicCounter_t) {

	for floor := 0; floor < elevator_IO.N_FLOORS; floor++{
		elevator_IO.SetButtonLamp(elevator_IO.BT_Cab, floor, converters.CC_ToBool(CabRequests[floor].Value))
	}
}
func LightHallLights(Hall_Requests [config.N_FLOORS][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t) {
	for floor := 0; floor < elevator_IO.N_FLOORS; floor++{
		elevator_IO.SetButtonLamp(elevator_IO.BT_HallUp, floor, converters.CC_ToBool(Hall_Requests[floor][elevator_IO.BT_HallUp].Value))
		elevator_IO.SetButtonLamp(elevator_IO.BT_HallDown, floor, converters.CC_ToBool(Hall_Requests[floor][elevator_IO.BT_HallDown].Value))
	}
}


//elevator moves down on init between floors
func OnInitBetweenFloors(guardianCommands chan elevatorStateGuardian.GuardianCommands_t, drv_floors chan int){
	elevator_IO.SetMotorDirection(elevator_IO.MD_Down)

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)

    for { 
		floor := <- drv_floors
		if floor != -1 {
            elevator_IO.SetMotorDirection(elevator_IO.MD_Stop)
            elevatorState.Floor = floor
			break
		}
	}

    elevatorState.ElevatorBehaviour = elevator_IO.EB_Idle
    elevatorState.MotorDirection    = elevator_IO.MD_Stop
    //Save in guardian
    guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
}


//what to do if we recieve new data
func OnReceivedDataFromMsgSync(
    guardianCommands chan elevatorStateGuardian.GuardianCommands_t, 
    doorTimerStart chan struct{}, 
    doorTimerStop chan struct{},
    isFunctionalStart chan struct{},
    isFunctionalStop chan struct{},
    isObstructed bool){

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)
    
    //obstruction is not affecting the elevator since door not open
    if  elevatorState.ElevatorBehaviour != elevator_IO.EB_DoorOpen{

        var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests);
        
        elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
        elevatorState.MotorDirection    = pair.MotorDirection
        //Save in guardian
        guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
    } 

    switch(elevatorState.ElevatorBehaviour){
    case elevator_IO.EB_DoorOpen:
        isFunctionalStop <- struct{}{}
        elevator_IO.SetDoorOpenLamp(true)

        //OBS!!
        doorTimerStart <- struct{}{}

        //change RequestsClearAtCurrentFloor return cleared request (in floor)
        requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests);
        guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}

    case elevator_IO.EB_Moving:
        isFunctionalStart <- struct{}{}
        elevator_IO.SetMotorDirection((elevatorState.MotorDirection))

    case elevator_IO.EB_Idle:
        isFunctionalStop <- struct{}{}
        elevator_IO.SetMotorDirection((elevatorState.MotorDirection))
    
    }

    // assignedRequests = elevatorStateGuardian.GetAssignedRequests(guardianCommands)
    // fmt.Printf("\nNew state from new data:\n");
    // ElevatorPrint(elevatorState, assignedRequests);
}


//what to do if we arrive at a floor
func OnFloorArrival(
    guardianCommands chan elevatorStateGuardian.GuardianCommands_t, 
    doorTimerStart chan struct{}, 
    doorTimerStop chan struct{}, 
    isFunctionalStart chan struct{}, 
    isFunctionalStop chan struct{}, 
    newFloor int,
    isObstructed bool) {

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)

    //update floor and IsFunctional
    if !isObstructed{
        elevatorState.IsFunctional = true
    }
    elevatorState.Floor = newFloor
    elevator_IO.SetFloorIndicator(newFloor)

    
    if elevatorState.ElevatorBehaviour == elevator_IO.EB_Moving {
        if requests.RequestsShouldStop(elevatorState, assignedRequests) {

            elevator_IO.SetMotorDirection(elevator_IO.MD_Stop)
            elevator_IO.SetDoorOpenLamp(true)
            
            elevatorState.ElevatorBehaviour = elevator_IO.EB_DoorOpen
            
            //removing this for keeping the previous direction, to avoid clearing both up and down in single floor when not supposed to
            //elevatorState.MotorDirection = elevator_IO.MD_Stop
            
            
            requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests) 
            guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}
            
            //RESET doorTimer
            doorTimerStop <- struct{}{}
            doorTimerStart <- struct{}{}
            
            //STOP isFunctional timer
            isFunctionalStop <- struct{}{}
        }
        //update the data
        guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
        //RESET isFunctionsl timer
        isFunctionalStop <- struct{}{}
        isFunctionalStart <- struct{}{}
    }

    assignedRequests = elevatorStateGuardian.GetAssignedRequests(guardianCommands)
    fmt.Printf("\nNew state from FloorArrival:\n");
    ElevatorPrint(elevatorState, assignedRequests);
}


//what to do if the door timer runs out
func OnDoorTimeout(
    guardianCommands chan elevatorStateGuardian.GuardianCommands_t, 
    doorTimerStart chan struct{}, 
    doorTimerStop chan struct{}, 
    isFunctionalStart chan struct{}, 
    isFunctionalStop chan struct{},
    isObstructed bool){

    elevatorState := elevatorStateGuardian.GetElevatorData(guardianCommands)
    assignedRequests := elevatorStateGuardian.GetAssignedRequests(guardianCommands)

    switch(elevatorState.ElevatorBehaviour){

    case elevator_IO.EB_DoorOpen:
        if !isObstructed{
            var pair requests.MotorDirectionBehaviourPair_t = requests.RequestsChooseDirection(elevatorState, assignedRequests);
            
            elevatorState.ElevatorBehaviour = pair.ElevatorBehaviour
            elevatorState.MotorDirection    = pair.MotorDirection
            //Save in guardian
            guardianCommands <- elevatorStateGuardian.SetElevatorData_t{ElevatorData: elevatorState}
        }

        switch(elevatorState.ElevatorBehaviour){
        case elevator_IO.EB_DoorOpen:
            doorTimerStop <- struct{}{}
            doorTimerStart <- struct{}{}

            requestsToClear := requests.RequestsClearAtCurrentFloor(elevatorState, assignedRequests) 
            guardianCommands <- elevatorStateGuardian.SetRequestsDone_t{RequestsToClear: requestsToClear}
            
        case elevator_IO.EB_Moving:

            isFunctionalStart <- struct{}{}
            elevator_IO.SetDoorOpenLamp(false)
            elevator_IO.SetMotorDirection(elevatorState.MotorDirection);

        case elevator_IO.EB_Idle:

            isFunctionalStop <- struct{}{}
            elevator_IO.SetDoorOpenLamp(false)
            elevator_IO.SetMotorDirection(elevatorState.MotorDirection);
        }
    default:
        break;
    }

    assignedRequests = elevatorStateGuardian.GetAssignedRequests(guardianCommands)
    fmt.Printf("\nNew state from DoorTimeout:\n");
    ElevatorPrint(elevatorState, assignedRequests);
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

