package messageSync

import (
	"fmt"
	"strconv"
	"theProject/networkDriver/peers"
	"theProject/elevator"
)

//Initalizing the the systemData and confirmedSystemData in Message_Sync_Server
//All values are initialized to 0, -1 (not in a floor, EB_Idle, CC_Uninit and empty barriers
func initSystemData(localID int) (SystemData_t, SystemData_t){

    var tmpCabRequests []RequestCyclicCounter_t = make([]RequestCyclicCounter_t, elevator.N_FLOORS)

    for floor := 0; floor < elevator.N_FLOORS; floor++ {
        tmpCabRequests[floor] = RequestCyclicCounter_t{
            Value:   CC_Uninit,
            Barrier: make([]bool, N_ELEVATORS),
        }
    }

    var tmpHallRequestData [][2]RequestCyclicCounter_t = make([][2]RequestCyclicCounter_t, elevator.N_FLOORS)

    for floor := 0; floor < elevator.N_FLOORS; floor++ {
        for btn := 0; btn < 2; btn++ {
            tmpHallRequestData[floor][btn] = RequestCyclicCounter_t{
                Value:   CC_Uninit,
                Barrier: make([]bool, N_ELEVATORS),
            }
        }
    }
	
	var tmpElevatorData []ElevatorData_t = make([]ElevatorData_t, N_ELEVATORS)

	for i := 0; i < N_ELEVATORS; i++ {
		tmpElevatorData[i] = ElevatorData_t{
			Id: i,
			//Msg_counter: 0,
			IsAlive: IsAliveData_t{
				Value:   false,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			IsFunctional: IsFunctionalData_t{
				Value:   false,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			Floor: FloorData_t{
				Value:   -1,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			ElevatorBehaviour: ElevatorBehaviourData_t{
				Value:   EB_Idle,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			MotorDirection: MotorDirectionData_t{
				Value:   MD_Stop,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			CabRequests: DeepCopyCabRequests(tmpCabRequests),
		}
	}

	var systemData SystemData_t = SystemData_t{ID: localID, ElevatorData: tmpElevatorData, HallRequestData: tmpHallRequestData}

	return systemData, deepCopySystemData(systemData)
}

//Processing the fresh data and undating systemData and confirmedSystemData accordingly
func onReceivedFreshData(systemData SystemData_t, 
							confirmedSystemData SystemData_t, 
							fresh_data SystemData_t) (SystemData_t, SystemData_t, bool){

	var updatedSystemData SystemData_t = deepCopySystemData(systemData)
	var updatedConfirmedSystemData SystemData_t = deepCopySystemData(confirmedSystemData)
	var isConfirmedDataUpdated bool = false

	for i := 0; i < N_ELEVATORS; i++{
		//if the new data have newer information about a elevator, we accept it

		if systemData.ElevatorData[i].ID == fresh_data.ID {
			updatedSystemData.ElevatorData[i] = updateElevatorDataAboutSelf(systemData.ElevatorData[i], fresh_data.ElevatorData[i], systemData.ID)
		} else {
			updatedSystemData.ElevatorData[i] = updateElevatorDataAboutOther(systemData.ElevatorData[i], fresh_data.ElevatorData[i], systemData.ID)
		}
	}
	//update hall requests with the cyclic counter 
	updatedSystemData.HallRequestData = updateHallRequestData(systemData.HallRequestData, fresh_data.HallRequestData, systemData.ID)

	//update the confirmed data that have recieved consensus
	updatedConfirmedSystemData, isConfirmedDataUpdated = updateConfirmedSystemData(systemData, confirmedSystemData)
	
	return updatedSystemData, updatedConfirmedSystemData, isConfirmedDataUpdated
}

//Functions for safely updating the system data
//-----------------------------------------------------------
func updateHallRequestData(	oldData [][2]RequestCyclicCounter_t, 
								newData [][2]RequestCyclicCounter_t, 
								ID int) [][2]RequestCyclicCounter_t {

	var updatedHallRequests [][2]RequestCyclicCounter_t = deepCopyHallRequests(oldData)
	
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			updatedHallRequests[floor][btn] = update_CC(oldData[floor][btn], newData[floor][btn], ID)
		}
	}
	return updatedHallRequests
}

//We trust info an elevator tells about itself. 
//If the data is the same: update barrier
//If the data is not the same: accept new data and sign the barrier
func updateElevatorDataAboutSelf(	oldData ElevatorData_t, 
										newData ElevatorData_t, 
										ID int) ElevatorData_t { 

	var updated_data ElevatorData_t = deepCopySingleElevatorData(oldData)

	if (oldData.IsAlive 			== newData.IsAlive && 
		oldData.IsFunctional 		== newData.IsFunctional &&
		oldData.Floor 				== newData.Floor &&
		oldData.ElevatorBehaviour	== newData.ElevatorBehaviour &&
		oldData.MotorDirection 		== newData.MotorDirection ){

		updated_data.ElevatorBarrier = boolUnion(oldData.ElevatorBarrier, newData.ElevatorBarrier)
	} else {
		updated_data.IsAlive 			= newData.IsAlive
		updated_data.IsFunctional 		= newData.IsFunctional
		updated_data.Floor 				= newData.Floor
		updated_data.ElevatorBehaviour	= newData.ElevatorBehaviour
		updated_data.MotorDirection 	= newData.MotorDirection

		updated_data.ElevatorBarrier = deepCopyBarrier(newData.ElevatorBarrier)
		updated_data.ElevatorBarrier[ID] = true
	}

	for i := 0; i < N_ELEVATORS; i++{
		updated_data.CabRequests[i] = update_CC(oldData.CabRequests[i], newData.CabRequests[i], ID)
	}

	return updated_data
}

//Only update cab requests CC and update barrier
func updateElevatorDataAboutOther(	oldData ElevatorData_t, 
										newData ElevatorData_t, 
										ID int) ElevatorData_t {

	var updated_data ElevatorData_t = deepCopySingleElevatorData(oldData)

	if (oldData.IsAlive 			== newData.IsAlive && 
		oldData.IsFunctional 		== newData.IsFunctional &&
		oldData.Floor 				== newData.Floor &&
		oldData.ElevatorBehaviour	== newData.ElevatorBehaviour &&
		oldData.MotorDirection 		== newData.MotorDirection ){

		updated_data.ElevatorBarrier = boolUnion(oldData.ElevatorBarrier, newData.ElevatorBarrier)
	}
	
	for i := 0; i < N_ELEVATORS; i++{
		updated_data.CabRequests[i] = update_CC(oldData.CabRequests[i], newData.CabRequests[i], ID)
	}

	return updated_data 
}

//Checking the Barrier 
func updateConfirmedSystemData(	unconfirmedData SystemData_t, 
									confirmedData SystemData_t) (SystemData_t, bool) {
	var updatedConfirmedData SystemData_t = confirmedData
	var isUpdated bool = false
	
	for i := 0; i < N_ELEVATORS; i++ {
		
		if checkBarrier(unconfirmedData.ElevatorData[i].ElevatorBarrier) {
			//If there is new data, we update
			if (unconfirmedData.ElevatorData[i].IsAlive 			!= confirmedData.ElevatorData[i].IsAlive ||
				unconfirmedData.ElevatorData[i].IsFunctional 		!= confirmedData.ElevatorData[i].IsFunctional ||
				unconfirmedData.ElevatorData[i].Floor 				!= confirmedData.ElevatorData[i].Floor ||
				unconfirmedData.ElevatorData[i].ElevatorBehaviour	!= confirmedData.ElevatorData[i].ElevatorBehaviour ||
				unconfirmedData.ElevatorData[i].MotorDirection 	!= confirmedData.ElevatorData[i].MotorDirection ){

					updatedConfirmedData.ElevatorData[i].IsAlive 			= unconfirmedData.ElevatorData[i].IsAlive
					updatedConfirmedData.ElevatorData[i].IsFunctional 		= unconfirmedData.ElevatorData[i].IsFunctional
					updatedConfirmedData.ElevatorData[i].Floor 				= unconfirmedData.ElevatorData[i].Floor
					updatedConfirmedData.ElevatorData[i].ElevatorBehaviour 	= unconfirmedData.ElevatorData[i].ElevatorBehaviour
					updatedConfirmedData.ElevatorData[i].MotorDirection 	= unconfirmedData.ElevatorData[i].MotorDirection
					isUpdated = true
			}
		}

		//Dont need Barrier check since update_CC() have Barrier checks 
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if unconfirmedData.ElevatorData[i].CabRequests[floor].Value != confirmedData.ElevatorData[i].CabRequests[floor].Value {
				updatedConfirmedData.ElevatorData[i].CabRequests[floor].Value = unconfirmedData.ElevatorData[i].CabRequests[floor].Value
				isUpdated = true
			}
		}
	}

	//Dont need Barrier check since update_CC() have Barrier checks 
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			if unconfirmedData.HallRequestData[floor][btn].Value != confirmedData.HallRequestData[floor][btn].Value {
				unconfirmedData.HallRequestData[floor][btn] = confirmedData.HallRequestData[floor][btn]
				isUpdated = true
			}
		}
	}

	return confirmedData, isUpdated
}

func update_CC(	old_CC RequestCyclicCounter_t, 
				new_CC RequestCyclicCounter_t, 
				ID int) RequestCyclicCounter_t {

	 var updated_CC RequestCyclicCounter_t = old_CC

	//update the CC based on rules
	if old_CC.Value == CC_Done && new_CC.Value == CC_No{
		//Accept new value
		updated_CC.Value 		= new_CC.Value
		updated_CC.Barrier 		= deepCopyBarrier(new_CC.Barrier)
		updated_CC.Barrier[ID] 	= true

	} else if old_CC.Value == CC_No && new_CC.Value == CC_Done{
		//Keep old value
		updated_CC.Value 	= old_CC.Value
		updated_CC.Barrier 	= deepCopyBarrier(old_CC.Barrier)

	} else if old_CC.Value == new_CC.Value{
		//They are the same, only update Barrier
		updated_CC.Barrier = boolUnion(old_CC.Barrier, new_CC.Barrier)

	} else if old_CC.Value < new_CC.Value {
		//Accept bigger value
		updated_CC.Value 		= new_CC.Value
		updated_CC.Barrier 		= deepCopyBarrier(new_CC.Barrier)
		updated_CC.Barrier[ID] 	= true
	}

	//update the CC if barriers are fulliled 
	if (updated_CC.Value == CC_Unconfirmed && checkBarrier(updated_CC.Barrier)){
		updated_CC.Value 		= CC_Confirmed
		updated_CC.Barrier 		= make([]bool, N_ELEVATORS)
		updated_CC.Barrier[ID] 	= true
	}
	if (updated_CC.Value == CC_Done && checkBarrier(updated_CC.Barrier)){
		updated_CC.Value 		= CC_No
		updated_CC.Barrier 		= make([]bool, N_ELEVATORS)
		updated_CC.Barrier[ID] 	= true
	}

	return updated_CC
}
//-----------------------------------------------------------

//TODO: these do not belong here
func lightCabLights(CabRequests []RequestCyclicCounter_t) {

	for floor := 0; floor < elevator.N_FLOORS; floor++{
		elevator.SetButtonLamp(elevator.BT_Cab, floor, CC_ToBool(CabRequests[floor].Value))
	}
}
func lightHallLights(Hall_Requests [][2]RequestCyclicCounter_t) {
	for floor := 0; floor < elevator.N_FLOORS; floor++{
		elevator.SetButtonLamp(elevator.BT_HallUp, floor, CC_ToBool(Hall_Requests[floor][elevator.BT_HallUp].Value))
		elevator.SetButtonLamp(elevator.BT_HallDown, floor, CC_ToBool(Hall_Requests[floor][elevator.BT_HallDown].Value))
	}
}
func CC_ToBool(CC CyclicCounter_t) bool {
	if (CC == CC_Uninit || CC == CC_No || CC == CC_Unconfirmed) {
		return false
	}
	if CC == CC_Confirmed || CC == CC_Done {
		return true
	} else {
		print("wrong CC Value")
		return false
	}
}

//Helper functions
//-----------------------------------------------------------
func checkBarrier(Barrier []bool) bool {
	for i := 0; i < N_ELEVATORS; i++{
		if Barrier[i] != activePeers[i]{
			return false
		}
	}
	return true
}

func boolUnion(a []bool, b []bool) []bool {
    n := len(a)
    if len(b) > n {
        n = len(b)
    }
    result := make([]bool, n)
    for i := 0; i < n; i++ {
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

func fromPeersUpdateToActivePeers(peersUpdate peers.PeerUpdate) []bool { 
	activePeers:= make([]bool, N_ELEVATORS)
	
	for _, peer := range peersUpdate.Peers {
		idx := peerStrToInt(peer)
		activePeers[idx] = true
	}
	return activePeers
}

func peerStrToInt(peerStr string) int {
	num, err := strconv.Atoi(peerStr)
	if err != nil {
		fmt.Println("Invalid number:", err)
		return -1
	}
	return num
}
//-----------------------------------------------------------


//Deep copy funtions for msg_sync_types
//-----------------------------------------------------------
func deepCopySystemData(src SystemData_t)SystemData_t{
	dst := src 
	dst.ElevatorData = deepCopyElevatordata(src.ElevatorData)
	dst.HallRequestData = deepCopyHallRequests(src.HallRequestData)
	return dst
}

func deepCopyElevatordata(src []ElevatorData_t) []ElevatorData_t {
	dst := make([]ElevatorData_t, len(src))

	for i := range src {
		dst[i] = deepCopySingleElevatorData(src[i])
	}
	return dst
}

func deepCopySingleElevatorData(src ElevatorData_t) ElevatorData_t {
	dst := src
	dst.ElevatorBarrier = deepCopyBarrier(src.ElevatorBarrier)
	dst.CabRequests = deepCopyCabRequests(src.CabRequests)

	return dst
}

func deepCopyHallRequests(src [][2]RequestCyclicCounter_t) [][2]RequestCyclicCounter_t {
	dst := make([][2]RequestCyclicCounter_t, len(src))
	
    for floor := range src {
		for btn := 0; btn < 2; btn++ {
			dst[floor][btn] = src[floor][btn]
			
			barrierCopy := make([]bool, len(src[floor][btn].Barrier))
			copy(barrierCopy, src[floor][btn].Barrier)
			dst[floor][btn].Barrier = barrierCopy
        }
    }
    return dst
}

func deepCopyCabRequests(src []RequestCyclicCounter_t) []RequestCyclicCounter_t {
	dst := make([]RequestCyclicCounter_t, len(src))
	for i := range src {
		dst[i] = src[i]
		if src[i].Barrier != nil {
			barrierCopy := make([]bool, len(src[i].Barrier))
			copy(barrierCopy, src[i].Barrier)
			dst[i].Barrier = barrierCopy
		}
	}
	return dst
}

func deepCopyBarrier(src []bool) []bool {
	dst := make([]bool, len(src))
	copy(dst, src)
	return dst
}
//-----------------------------------------------------------

