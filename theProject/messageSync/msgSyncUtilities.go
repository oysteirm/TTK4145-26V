package messageSync

import (
	"fmt"
	"strconv"
	"theProject/config"
	"theProject/elevator_IO"
	"theProject/networkDriver/peers"
)

// Initalizing the the systemData and confirmedSystemData in Message_Sync_Server
// All values are initialized to 0, -1 (not in a floor, EB_Idle, CC_Uninit and empty barriers
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

	for i := 0; i < config.N_ELEVATORS; i++ {
		tmpElevatorData[i] = ElevatorData_t{
			ID:                i,
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

// Processing the fresh data and undating systemData and confirmedSystemData accordingly
func OnReceivedFreshData(
	systemData SystemData_t,
	confirmedSystemData SystemData_t,
	freshSystemData SystemData_t) (SystemData_t, SystemData_t, bool) {

	//Starting with no changes in the data
	var updatedSystemData SystemData_t = systemData
	var updatedConfirmedSystemData SystemData_t = confirmedSystemData
	var isConfirmedDataUpdated bool = false

	for i := 0; i < config.N_ELEVATORS; i++ {

		if systemData.ElevatorData[i].ID == freshSystemData.ID {
			updatedSystemData.ElevatorData[i] = UpdateElevatorDataAboutSelf(systemData.ElevatorData[i], freshSystemData.ElevatorData[i], systemData.ID)

		} else {
			updatedSystemData.ElevatorData[i] = UpdateElevatorDataAboutOther(systemData.ElevatorData[i], freshSystemData.ElevatorData[i], systemData.ID)
		}
	}

	for elevatorID := 0; elevatorID < config.N_ELEVATORS; elevatorID++ {
		for floor := 0; floor < config.N_FLOORS; floor++ {
			updatedSystemData.ElevatorData[elevatorID].CabRequests[floor] = update_CC(
				systemData.ElevatorData[elevatorID].CabRequests[floor],
				freshSystemData.ElevatorData[elevatorID].CabRequests[floor],
				systemData.ID)
		}
	}

	//update hall requests with the cyclic counter
	updatedSystemData.HallRequestData = UpdateHallRequestData(systemData.HallRequestData, freshSystemData.HallRequestData, systemData.ID)

	//update the confirmed data that have recieved consensus
	updatedConfirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(updatedSystemData, confirmedSystemData)

	return updatedSystemData, updatedConfirmedSystemData, isConfirmedDataUpdated
}

// Functions for safely updating the system data
// -----------------------------------------------------------
func UpdateHallRequestData(oldData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t,
	freshData [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t,
	ID int) [config.N_FLOORS][config.N_UP_DOWN]RequestCyclicCounter_t {

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

	var updatedData ElevatorData_t = oldData

	if oldData.IsAlive == freshData.IsAlive &&
		oldData.IsFunctional == freshData.IsFunctional &&
		oldData.Floor == freshData.Floor &&
		oldData.ElevatorBehaviour == freshData.ElevatorBehaviour &&
		oldData.MotorDirection == freshData.MotorDirection {

		updatedData.ElevatorBarrier = boolUnion(oldData.ElevatorBarrier, freshData.ElevatorBarrier)
	} else {
		updatedData.IsAlive = freshData.IsAlive
		updatedData.IsFunctional = freshData.IsFunctional
		updatedData.Floor = freshData.Floor
		updatedData.ElevatorBehaviour = freshData.ElevatorBehaviour
		updatedData.MotorDirection = freshData.MotorDirection
		updatedData.ElevatorBarrier = freshData.ElevatorBarrier
		updatedData.ElevatorBarrier[ID] = true
	}

	// for floor := 0; floor < config.N_FLOORS; floor++ {

	// 	if oldData.CabRequests[floor].Value == freshData.CabRequests[floor].Value {
	// 		updatedData.CabRequests[floor].Barrier = boolUnion(oldData.CabRequests[floor].Barrier, freshData.CabRequests[floor].Barrier)

	// 	} else if freshData.CabRequests[floor].Value != CC_Uninit {
	// 		updatedData.CabRequests[floor] = freshData.CabRequests[floor]
	// 		updatedData.CabRequests[floor].Barrier[ID] = true
	// 	}
	// }

	return updatedData
}

// Only update cab requests CC and update barrier
func UpdateElevatorDataAboutOther(oldData ElevatorData_t,
	freshData ElevatorData_t,
	localID int) ElevatorData_t {

	var updatedData ElevatorData_t = oldData

	if oldData.IsAlive == freshData.IsAlive &&
		oldData.IsFunctional == freshData.IsFunctional &&
		oldData.Floor == freshData.Floor &&
		oldData.ElevatorBehaviour == freshData.ElevatorBehaviour &&
		oldData.MotorDirection == freshData.MotorDirection {

		updatedData.ElevatorBarrier = boolUnion(oldData.ElevatorBarrier, freshData.ElevatorBarrier)
	}

	// //Updating cab requests
	// for floor := 0; floor < config.N_FLOORS; floor++ {

	// 	if oldData.CabRequests[floor].Value == CC_Uninit {
	// 		updatedData.CabRequests[floor] = update_CC(oldData.CabRequests[floor], freshData.CabRequests[floor], localID)

	// 	} else if oldData.CabRequests[floor].Value == freshData.CabRequests[floor].Value {
	// 		updatedData.CabRequests[floor].Barrier = boolUnion(oldData.CabRequests[floor].Barrier, freshData.CabRequests[floor].Barrier)
	// 	}
	// }

	return updatedData
}

// Checking the Barrier
func updateConfirmedSystemData(unconfirmedData SystemData_t,
	confirmedData SystemData_t) (SystemData_t, bool) {

	var updatedConfirmedData SystemData_t = confirmedData
	var isUpdated bool = false

	for i := 0; i < config.N_ELEVATORS; i++ {

		if checkBarrier(unconfirmedData.ElevatorData[i].ElevatorBarrier) {
			//If there is new data, we update
			if unconfirmedData.ElevatorData[i].IsAlive != confirmedData.ElevatorData[i].IsAlive ||
				unconfirmedData.ElevatorData[i].IsFunctional != confirmedData.ElevatorData[i].IsFunctional ||
				unconfirmedData.ElevatorData[i].Floor != confirmedData.ElevatorData[i].Floor ||
				unconfirmedData.ElevatorData[i].ElevatorBehaviour != confirmedData.ElevatorData[i].ElevatorBehaviour ||
				unconfirmedData.ElevatorData[i].MotorDirection != confirmedData.ElevatorData[i].MotorDirection {

				updatedConfirmedData.ElevatorData[i].IsAlive = unconfirmedData.ElevatorData[i].IsAlive
				updatedConfirmedData.ElevatorData[i].IsFunctional = unconfirmedData.ElevatorData[i].IsFunctional
				updatedConfirmedData.ElevatorData[i].Floor = unconfirmedData.ElevatorData[i].Floor
				updatedConfirmedData.ElevatorData[i].ElevatorBehaviour = unconfirmedData.ElevatorData[i].ElevatorBehaviour
				updatedConfirmedData.ElevatorData[i].MotorDirection = unconfirmedData.ElevatorData[i].MotorDirection
				isUpdated = true
			}
		}

		//Dont need Barrier check since update_CC() have Barrier checks
		for floor := 0; floor < elevator_IO.N_FLOORS; floor++ {
			if unconfirmedData.ElevatorData[i].CabRequests[floor].Value != confirmedData.ElevatorData[i].CabRequests[floor].Value {
				updatedConfirmedData.ElevatorData[i].CabRequests[floor] = unconfirmedData.ElevatorData[i].CabRequests[floor]
				isUpdated = true
			}
		}
	}

	//Dont need Barrier check since update_CC() have Barrier checks
	for floor := 0; floor < elevator_IO.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if unconfirmedData.HallRequestData[floor][btn].Value != confirmedData.HallRequestData[floor][btn].Value {
				updatedConfirmedData.HallRequestData[floor][btn] = unconfirmedData.HallRequestData[floor][btn]
				isUpdated = true
			}
		}
	}

	return updatedConfirmedData, isUpdated
}

func update_CC(old_CC RequestCyclicCounter_t,
	new_CC RequestCyclicCounter_t,
	ID int) RequestCyclicCounter_t {

	var updated_CC RequestCyclicCounter_t = old_CC

	//update the CC based on rules
	if old_CC.Value == CC_Done && new_CC.Value == CC_No {
		//Accept new value
		updated_CC.Value = new_CC.Value
		updated_CC.Barrier = new_CC.Barrier
		updated_CC.Barrier[ID] = true

	} else if old_CC.Value == CC_No && new_CC.Value == CC_Done {
		//Keep old value
		updated_CC.Value = old_CC.Value
		updated_CC.Barrier = old_CC.Barrier

	} else if old_CC.Value == new_CC.Value {
		//They are the same, only update Barrier
		updated_CC.Barrier = boolUnion(old_CC.Barrier, new_CC.Barrier)

	} else if old_CC.Value < new_CC.Value {
		//Accept bigger value
		updated_CC.Value = new_CC.Value
		updated_CC.Barrier = new_CC.Barrier
		updated_CC.Barrier[ID] = true
	}

	//update the CC if barriers are fulliled
	if updated_CC.Value == CC_Unconfirmed && checkBarrier(updated_CC.Barrier) {
		updated_CC.Value = CC_Confirmed
		updated_CC.Barrier = [config.N_ELEVATORS]bool{}
		updated_CC.Barrier[ID] = true
	}
	if updated_CC.Value == CC_Done && checkBarrier(updated_CC.Barrier) {
		updated_CC.Value = CC_No
		updated_CC.Barrier = [config.N_ELEVATORS]bool{}
		updated_CC.Barrier[ID] = true
	}

	return updated_CC
}

//-----------------------------------------------------------

// Helper functions
// -----------------------------------------------------------
func checkBarrier(Barrier [config.N_ELEVATORS]bool) bool {
	for i := 0; i < config.N_ELEVATORS; i++ {
		if ActivePeers[i] && !Barrier[i] {
			return false
		}
	}
	return true
}

func boolUnion(a [config.N_ELEVATORS]bool, b [config.N_ELEVATORS]bool) [config.N_ELEVATORS]bool {

	var result [config.N_ELEVATORS]bool
	for i := 0; i < config.N_ELEVATORS; i++ {
		var valA, valB bool
		if i < len(a) {
			valA = a[i]
		}
		if i < len(b) {
			valB = b[i]
		}
		result[i] = valA || valB
	}
	return result
}

func fromPeersUpdateToActivePeers(peersUpdate peers.PeerUpdate) [config.N_ELEVATORS]bool {
	var ActivePeers [config.N_ELEVATORS]bool

	for _, peer := range peersUpdate.Peers {
		idx := peerStrToInt(peer)
		ActivePeers[idx] = true
	}
	return ActivePeers
}

func peerStrToInt(peerStr string) int {
	num, err := strconv.Atoi(peerStr)
	if err != nil {
		fmt.Println("Invalid number:", err)
		return -1
	}
	return num
}

func update_CC_ForCurrentPeers(systemData SystemData_t, localID int) SystemData_t {
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

// Deep copy funtions for msg_sync_types
// -----------------------------------------------------------
// func DeepCopySystemData(src SystemData_t) SystemData_t {
// 	dst                := src
// 	dst.ElevatorData    = deepCopyElevatordata(src.ElevatorData)
// 	dst.HallRequestData = DeepCopyHallRequests(src.HallRequestData)
// 	return dst
// }

// func deepCopyElevatordata(src []ElevatorData_t) []ElevatorData_t {
// 	dst := make([]ElevatorData_t, len(src))

// 	for i := range src {
// 		dst[i] = deepCopySingleElevatorData(src[i])
// 	}
// 	return dst
// }

// func deepCopySingleElevatorData(src ElevatorData_t) ElevatorData_t {
// 	dst                := src
// 	dst.ElevatorBarrier = DeepCopyBarrier(src.ElevatorBarrier)
// 	dst.CabRequests     = deepCopyCabRequests(src.CabRequests)

// 	return dst
// }

// func DeepCopyHallRequests(src [][2]RequestCyclicCounter_t) [][2]RequestCyclicCounter_t {
// 	dst := make([][2]RequestCyclicCounter_t, len(src))

// 	for floor := range src {
// 		for btn := 0; btn < 2; btn++ {
// 			dst[floor][btn] = src[floor][btn]

// 			barrierCopy := make([]bool, len(src[floor][btn].Barrier))
// 			copy(barrierCopy, src[floor][btn].Barrier)
// 			dst[floor][btn].Barrier = barrierCopy
// 		}
// 	}
// 	return dst
// }

// func deepCopyCabRequests(src []RequestCyclicCounter_t) []RequestCyclicCounter_t {
// 	dst := make([]RequestCyclicCounter_t, len(src))
// 	for i := range src {
// 		dst[i] = src[i]
// 		if src[i].Barrier != nil {
// 			barrierCopy   := make([]bool, len(src[i].Barrier))
// 			copy(barrierCopy, src[i].Barrier)
// 			dst[i].Barrier = barrierCopy
// 		}
// 	}
// 	return dst
// }

// func DeepCopyBarrier(src []bool) []bool {
// 	dst := make([]bool, len(src))
// 	copy(dst, src)
// 	return dst
// }

// //-----------------------------------------------------------
