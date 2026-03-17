package messageSync

import (
	"fmt"
	"strconv"
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/networkDriver/peers"
)


// Initalizing systemData_t types
func InitSystemData(localID int) (SystemData_t, SystemData_t) {

	var tmpCabRequests [config.N_FLOORS]RequestCyclicCounter_t

	for floor := 0; floor < config.N_FLOORS; floor++ {
		tmpCabRequests[floor] = RequestCyclicCounter_t{
			Value:   CC_Uninit,
			Barrier: [config.N_ELEVATORS]bool{},
		}
	}

	var tmpHallRequestData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			tmpHallRequestData[floor][btn] = RequestCyclicCounter_t{
				Value:   CC_Uninit,
				Barrier: [config.N_ELEVATORS]bool{},
			}
		}
	}

	var tmpElevatorData [config.N_ELEVATORS]ElevatorData_t

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		tmpElevatorData[elevatorID] = ElevatorData_t{
			ID:                elevatorID,
			IsAlive:           true,
			IsFunctional:      true,
			Floor:             -1,
			ElevatorBehaviour: elevator_IO.EB_Idle,
			MotorDirection:    elevator_IO.MD_Stop,
			CabRequests:       tmpCabRequests,
		}
	}

	var systemData SystemData_t = SystemData_t{ID: localID, ElevatorData: tmpElevatorData, HallRequestData: tmpHallRequestData}

	return systemData, systemData
}

// Processing the freshSystemData and updating systemData and confirmedSystemData accordingly
func OnReceivedFreshData(
	systemData SystemData_t,
	confirmedSystemData SystemData_t,
	freshSystemData SystemData_t) (SystemData_t, SystemData_t, bool) {

	//Starting with no changes
	var updatedSystemData SystemData_t = systemData
	var updatedConfirmedSystemData SystemData_t = confirmedSystemData
	var isConfirmedDataUpdated bool = false

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		// Difference between an elevator giving information about itself and about another elevator.
		if systemData.ElevatorData[elevatorID].ID == freshSystemData.ID {
			updatedSystemData.ElevatorData[elevatorID] = UpdateElevatorDataAboutSelf(	systemData.ElevatorData[elevatorID], 
																						freshSystemData.ElevatorData[elevatorID], 
																						systemData.ID)

		} else {
			updatedSystemData.ElevatorData[elevatorID] = UpdateElevatorDataAboutOther(	systemData.ElevatorData[elevatorID], 
																						freshSystemData.ElevatorData[elevatorID], 
																						systemData.ID)
		}
	}
	
	//TODO: evaluate if this belongs here
	// Updating all the cab requests using the Cyclic Counter logic
	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		for floor := 0; floor < config.N_FLOORS; floor++ {
			updatedSystemData.ElevatorData[elevatorID].CabRequests[floor] = update_CC(
				systemData.ElevatorData[elevatorID].CabRequests[floor],
				freshSystemData.ElevatorData[elevatorID].CabRequests[floor],
				systemData.ID)
		}
	}

	// Update hall requests with the cyclic counter
	updatedSystemData.HallRequestData = UpdateHallRequestData(	systemData.HallRequestData, 
																freshSystemData.HallRequestData, 
																systemData.ID)

	// Update the confirmed data with the data that have recieved consensus, and remebering we updated something
	updatedConfirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(updatedSystemData, confirmedSystemData)

	return updatedSystemData, updatedConfirmedSystemData, isConfirmedDataUpdated
}


// Functions for updating confirmedSystemData, elevatorData and requests
// -----------------------------------------------------------
// Udating hall requests
func UpdateHallRequestData(
	oldData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t,
	freshData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t,
	ID int) [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t {

	//Starting with no changes
	var updatedHallRequests [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t = oldData

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_UP_DOWN; btn++ {
			updatedHallRequests[floor][btn] = update_CC(oldData[floor][btn], freshData[floor][btn], ID)
		}
	}
	return updatedHallRequests
}

// We trust info an elevator tells about itself.
// If the data is the same: update barrier
// If the data is not the same: accept new data and sign the barrier
func UpdateElevatorDataAboutSelf(
	oldData ElevatorData_t,
	freshData ElevatorData_t,
	ID int) ElevatorData_t {

	//Starting with no changes
	var updatedData ElevatorData_t = oldData

	if  oldData.IsAlive 			== freshData.IsAlive &&
		oldData.IsFunctional 		== freshData.IsFunctional &&
		oldData.Floor 				== freshData.Floor &&
		oldData.ElevatorBehaviour 	== freshData.ElevatorBehaviour &&
		oldData.MotorDirection 		== freshData.MotorDirection {

		updatedData.ElevatorBarrier = boolUnion(oldData.ElevatorBarrier, freshData.ElevatorBarrier)
	} else {
		updatedData.IsAlive 			= freshData.IsAlive
		updatedData.IsFunctional 		= freshData.IsFunctional
		updatedData.Floor 				= freshData.Floor
		updatedData.ElevatorBehaviour 	= freshData.ElevatorBehaviour
		updatedData.MotorDirection 		= freshData.MotorDirection
		updatedData.ElevatorBarrier 	= freshData.ElevatorBarrier
		updatedData.ElevatorBarrier[ID] = true
	}

	return updatedData
}

// We don't trust info an elevator tells about others
// Update barrier if data is the same
func UpdateElevatorDataAboutOther(
	oldData ElevatorData_t,
	freshData ElevatorData_t,
	ID int) ElevatorData_t {

	//Starting with no changes
	var updatedData ElevatorData_t = oldData

	if oldData.IsAlive == freshData.IsAlive &&
		oldData.IsFunctional == freshData.IsFunctional &&
		oldData.Floor == freshData.Floor &&
		oldData.ElevatorBehaviour == freshData.ElevatorBehaviour &&
		oldData.MotorDirection == freshData.MotorDirection {

		updatedData.ElevatorBarrier = boolUnion(oldData.ElevatorBarrier, freshData.ElevatorBarrier)
	}

	return updatedData
}

// Updating confirmed data and detecting if updated
func updateConfirmedSystemData(
	unconfirmedData SystemData_t,
	confirmedData SystemData_t) (SystemData_t, bool) {

	//Starting with no changes
	var updatedConfirmedData SystemData_t = confirmedData
	var isUpdated bool = false

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		// Check for consensus
		if checkBarrier(unconfirmedData.ElevatorData[elevatorID].ElevatorBarrier) {
			// If there is new data, we update
			if  unconfirmedData.ElevatorData[elevatorID].IsAlive 			!= confirmedData.ElevatorData[elevatorID].IsAlive ||
				unconfirmedData.ElevatorData[elevatorID].IsFunctional 		!= confirmedData.ElevatorData[elevatorID].IsFunctional ||
				unconfirmedData.ElevatorData[elevatorID].Floor 				!= confirmedData.ElevatorData[elevatorID].Floor ||
				unconfirmedData.ElevatorData[elevatorID].ElevatorBehaviour 	!= confirmedData.ElevatorData[elevatorID].ElevatorBehaviour ||
				unconfirmedData.ElevatorData[elevatorID].MotorDirection 	!= confirmedData.ElevatorData[elevatorID].MotorDirection {

				updatedConfirmedData.ElevatorData[elevatorID].IsAlive 			= unconfirmedData.ElevatorData[elevatorID].IsAlive
				updatedConfirmedData.ElevatorData[elevatorID].IsFunctional 		= unconfirmedData.ElevatorData[elevatorID].IsFunctional
				updatedConfirmedData.ElevatorData[elevatorID].Floor 			= unconfirmedData.ElevatorData[elevatorID].Floor
				updatedConfirmedData.ElevatorData[elevatorID].ElevatorBehaviour = unconfirmedData.ElevatorData[elevatorID].ElevatorBehaviour
				updatedConfirmedData.ElevatorData[elevatorID].MotorDirection 	= unconfirmedData.ElevatorData[elevatorID].MotorDirection
				isUpdated = true
			}
		}

		// Dont need Barrier check since update_CC() have Barrier checks
		for floor := 0; floor < elevator_IO.N_FLOORS; floor++ {

			if  unconfirmedData.ElevatorData[elevatorID].CabRequests[floor].Value != 
				confirmedData.ElevatorData[elevatorID].CabRequests[floor].Value {

				updatedConfirmedData.ElevatorData[elevatorID].CabRequests[floor] = unconfirmedData.ElevatorData[elevatorID].CabRequests[floor]
				isUpdated = true
			}
		}
	}

	// Dont need Barrier check since update_CC() have Barrier checks
	for floor := 0; floor < elevator_IO.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {

			if  unconfirmedData.HallRequestData[floor][btn].Value != 
				confirmedData.HallRequestData[floor][btn].Value {

				updatedConfirmedData.HallRequestData[floor][btn] = unconfirmedData.HallRequestData[floor][btn]
				isUpdated = true
			}
		}
	}

	return updatedConfirmedData, isUpdated
}

// Updates cyclic counter
func update_CC(
	old_CC RequestCyclicCounter_t,
	new_CC RequestCyclicCounter_t,
	ID int) RequestCyclicCounter_t {

		// Starting with no changes
	var updated_CC RequestCyclicCounter_t = old_CC

	// Cycle back to CC_No
	if old_CC.Value == CC_Done && new_CC.Value == CC_No {
		updated_CC.Value 		= new_CC.Value
		updated_CC.Barrier 		= new_CC.Barrier
		updated_CC.Barrier[ID] 	= true

	// Can't go from 0 -> max value
	} else if old_CC.Value == CC_No && new_CC.Value == CC_Done {
		updated_CC.Value 	= old_CC.Value
		updated_CC.Barrier 	= old_CC.Barrier

	// They are the same, only update Barrier
	} else if old_CC.Value == new_CC.Value {
		updated_CC.Barrier = boolUnion(old_CC.Barrier, new_CC.Barrier)

	// Accept bigger value
	} else if old_CC.Value < new_CC.Value {
		updated_CC.Value 		= new_CC.Value
		updated_CC.Barrier 		= new_CC.Barrier
		updated_CC.Barrier[ID] 	= true
	}

	// Transition the CC if consesus is reached
	if updated_CC.Value == CC_Unconfirmed && checkBarrier(updated_CC.Barrier) {
		updated_CC.Value 		= CC_Confirmed
		updated_CC.Barrier 		= [config.N_ELEVATORS]bool{}
		updated_CC.Barrier[ID] 	= true
	}
	if updated_CC.Value == CC_Done && checkBarrier(updated_CC.Barrier) {
		updated_CC.Value 		= CC_No
		updated_CC.Barrier 		= [config.N_ELEVATORS]bool{}
		updated_CC.Barrier[ID] 	= true
	}

	return updated_CC
}

// Checks if request cyclic counters can transition to new state based on ActivePeers
func update_CC_ForCurrentPeers(systemData SystemData_t, localID int) SystemData_t {

	//Starting with no changes
	updatedSystemData := systemData

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		for floor := 0; floor < config.N_FLOORS; floor++ {
		request_CC := updatedSystemData.ElevatorData[elevatorID].CabRequests[floor]
			updatedSystemData.ElevatorData[elevatorID].CabRequests[floor] = update_CC(request_CC, request_CC, localID)
		}
	}

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_UP_DOWN; btn++ {
			request_CC := updatedSystemData.HallRequestData[floor][btn]
			updatedSystemData.HallRequestData[floor][btn] = update_CC(request_CC, request_CC, localID)
		}
	}
	return updatedSystemData
}

//-----------------------------------------------------------

// Helper functions
// -----------------------------------------------------------
// Compares barrier with ActivePeers
func checkBarrier(barrier [config.N_ELEVATORS]bool) bool {

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		if ActivePeers[elevatorID] && !barrier[elevatorID] {
			return false
		}
	}
	return true
}

// Returns union of two bool lists
func boolUnion(a [config.N_ELEVATORS]bool, b [config.N_ELEVATORS]bool) [config.N_ELEVATORS]bool {

	var result [config.N_ELEVATORS]bool

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		var valA, valB bool
		if elevatorID < len(a) {
			valA = a[elevatorID]
		}
		if elevatorID < len(b) {
			valB = b[elevatorID]
		}
		result[elevatorID] = valA || valB
	}
	return result
}

// Type coverting 
func fromPeersUpdateToActivePeers(peersUpdate peers.PeerUpdate) [config.N_ELEVATORS]bool {

	var ActivePeers [config.N_ELEVATORS]bool

	for _, peer := range peersUpdate.Peers {
		idx, _ := strconv.Atoi(peer)
		ActivePeers[idx] = true
	}
	return ActivePeers
}

// -----------------------------------------------------------

// Prints the whole system
func SystemPrintHorizontal(systemData SystemData_t) {

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Print("  +--------------------+")
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | Elevator: %-2d       |", i)
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  |IsAlive = %-9t |", systemData.ElevatorData[i].IsAlive)
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  |IsFunctional = %-2t |", systemData.ElevatorData[i].IsFunctional)
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | floor = %-2d         |", systemData.ElevatorData[i].Floor)
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | dirn  = %-10s |",
			ElevatorDirnToString(systemData.ElevatorData[i].MotorDirection))
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Printf("  | behav = %-10s |",
			ElevatorBehaviourToString(systemData.ElevatorData[i].ElevatorBehaviour))
	}
	fmt.Println()

	for i := 0; i < config.N_ELEVATORS; i++ {
		fmt.Print("  |  | up  | dn  | cab |")
	}
	fmt.Println()

	for f := elevator_IO.N_FLOORS - 1; f >= 0; f-- {

		for i := 0; i < config.N_ELEVATORS; i++ {

			fmt.Printf("  | %d", f)

			for btn := elevator_IO.ButtonType_t(0); btn < elevator_IO.N_BUTTONS; btn++ {

				if (f == elevator_IO.N_FLOORS-1 && btn == elevator_IO.BT_HallUp) ||
					(f == 0 && btn == elevator_IO.BT_HallDown) {

					fmt.Print("|     ")

				} else {

					if btn == elevator_IO.BT_Cab {
						if CC_ToBool(systemData.ElevatorData[i].CabRequests[f].Value) {
							fmt.Print("|  #  ")
						} else {
							fmt.Print("|  -  ")
						}
					} else {
						if CC_ToBool(systemData.HallRequestData[f][btn].Value) {
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

//The functions below is also implemented other places, but is duplicated due to cycle import problems
func CC_ToBool(cc CyclicCounter_t) bool {
	switch cc {
	case CC_Confirmed, CC_Done:
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
