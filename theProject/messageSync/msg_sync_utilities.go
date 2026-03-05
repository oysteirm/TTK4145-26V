package messageSync

import (
	"the_project/elevator"
)

func InitSystemData(localID int) (SystemData_t, SystemData_t) {

	var tmpCabRequests []RequestCyclicCounter_t = make([]RequestCyclicCounter_t, elevator.N_FLOORS)

	for floor := 0; floor < N_FLOORS; floor++ {
		tmpCabRequests[floor] = RequestCyclicCounter_t{
			Value:   CC_Uninit,
			Barrier: make(ElevList_t, N_ELEVATORS),
		}
	}

	var tmpHallRequestData [][2]RequestCyclicCounter_t = make([][2]RequestCyclicCounter_t, N_FLOORS)

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			tmpHallRequestData[floor][btn] = RequestCyclicCounter_t{
				Value:   CC_Uninit,
				Barrier: make(ElevList_t, N_ELEVATORS),
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

	var systemData SystemData_t = SystemData_t{Id: localID, ElevatorData: tmpElevatorData, HallRequestData: tmpHallRequestData}

	return systemData, DeepCopySystemData(systemData)
}

func OnReceivedFreshData(systemData SystemData_t, confirmedSystemData SystemData_t, freshData SystemData_t) (SystemData_t, SystemData_t, bool) {
	var updatedSystemData SystemData_t = DeepCopySystemData(systemData)
	var updatedConfirmedSystemData SystemData_t = DeepCopySystemData(confirmedSystemData)
	var isConfirmedDataUpdated bool = false

	for i := 0; i < N_ELEVATORS; i++ {
		//if the new data have newer information about a elevator, we accept it

		if systemData.ElevatorData[i].Id == freshData.Id {
			updatedSystemData.ElevatorData[i] = UpdateElevatorDataAboutSelf(systemData.ElevatorData[i], freshData.ElevatorData[i], systemData.Id)
		} else {
			updatedSystemData.ElevatorData[i] = UpdateElevatorDataAboutOther(systemData.ElevatorData[i], freshData.ElevatorData[i], systemData.Id)
		}
	}
	//update hall requests with the cyclic counter
	updatedSystemData.HallRequestData = UpdateHallRequestData(systemData.HallRequestData, freshData.HallRequestData, systemData.Id)

	//update the confirmed data that have recieved consensus
	updatedConfirmedSystemData, isConfirmedDataUpdated = UpdateConfirmedSystemData(systemData, confirmedSystemData)

	return updatedSystemData, updatedConfirmedSystemData, isConfirmedDataUpdated
}

func UpdateHallRequestData(oldData [][2]RequestCyclicCounter_t, newData [][2]RequestCyclicCounter_t, id int) [][2]RequestCyclicCounter_t {
	var updatedHallRequests [][2]RequestCyclicCounter_t = DeepCopyHallRequests(oldData)

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			updatedHallRequests[floor][btn] = UpdateCC(oldData[floor][btn], newData[floor][btn], id)
		}
	}
	return updatedHallRequests
}

// We trust info an elevator tells about itself.
func UpdateElevatorDataAboutSelf(oldData ElevatorData_t, newData ElevatorData_t, id int) ElevatorData_t {
	var updatedData ElevatorData_t = DeepCopySingleElevatorData(oldData)

	if oldData.IsAlive.Value == newData.IsAlive.Value {
		updatedData.IsAlive.Barrier = BoolUnion(oldData.IsAlive.Barrier, newData.IsAlive.Barrier)
	} else {
		updatedData.IsAlive = newData.IsAlive
		updatedData.IsAlive.Barrier[id] = true
	}

	if oldData.IsFunctional.Value == newData.IsFunctional.Value {
		updatedData.IsFunctional.Barrier = BoolUnion(oldData.IsFunctional.Barrier, newData.IsFunctional.Barrier)
	} else {
		updatedData.IsFunctional = newData.IsFunctional
		updatedData.IsFunctional.Barrier[id] = true
	}

	if oldData.Floor.Value == newData.Floor.Value {
		updatedData.Floor.Barrier = BoolUnion(oldData.Floor.Barrier, newData.Floor.Barrier)
	} else {
		updatedData.Floor = newData.Floor
		updatedData.Floor.Barrier[id] = true
	}

	if oldData.ElevatorBehaviour.Value == newData.ElevatorBehaviour.Value {
		updatedData.ElevatorBehaviour.Barrier = BoolUnion(oldData.ElevatorBehaviour.Barrier, newData.ElevatorBehaviour.Barrier)
	} else {
		updatedData.ElevatorBehaviour = newData.ElevatorBehaviour
		updatedData.ElevatorBehaviour.Barrier[id] = true
	}

	if oldData.MotorDirection.Value == newData.MotorDirection.Value {
		updatedData.MotorDirection.Barrier = BoolUnion(oldData.MotorDirection.Barrier, newData.MotorDirection.Barrier)
	} else {
		updatedData.MotorDirection = newData.MotorDirection
		updatedData.MotorDirection.Barrier[id] = true
	}

	for i := 0; i < N_ELEVATORS; i++ {
		updatedData.CabRequests[i] = UpdateCC(oldData.CabRequests[i], newData.CabRequests[i], id)
	}

	return updatedData
}

// Only update cabcalls CC or update Barriers
func UpdateElevatorDataAboutOther(oldData ElevatorData_t, newData ElevatorData_t, id int) ElevatorData_t {
	var updatedData ElevatorData_t = DeepCopySingleElevatorData(oldData)

	if oldData.IsAlive.Value == newData.IsAlive.Value {
		updatedData.IsAlive.Barrier = BoolUnion(oldData.IsAlive.Barrier, newData.IsAlive.Barrier)
	}

	if oldData.IsFunctional.Value == newData.IsFunctional.Value {
		updatedData.IsFunctional.Barrier = BoolUnion(oldData.IsFunctional.Barrier, newData.IsFunctional.Barrier)
	}

	if oldData.Floor.Value == newData.Floor.Value {
		updatedData.Floor.Barrier = BoolUnion(oldData.Floor.Barrier, newData.Floor.Barrier)
	}

	if oldData.ElevatorBehaviour.Value == newData.ElevatorBehaviour.Value {
		updatedData.ElevatorBehaviour.Barrier = BoolUnion(oldData.ElevatorBehaviour.Barrier, newData.ElevatorBehaviour.Barrier)
	}

	if oldData.MotorDirection.Value == newData.MotorDirection.Value {
		updatedData.MotorDirection.Barrier = BoolUnion(oldData.MotorDirection.Barrier, newData.MotorDirection.Barrier)
	}

	for i := 0; i < N_ELEVATORS; i++ {
		updatedData.CabRequests[i] = UpdateCC(oldData.CabRequests[i], newData.CabRequests[i], id)
	}

	return updatedData
}

func UpdateConfirmedSystemData(unconfinedData SystemData_t, confirmedData SystemData_t) (SystemData_t, bool) {
	var isUpdated bool = false

	for i := 0; i < N_ELEVATORS; i++ {
		if CheckBarrier(unconfinedData.ElevatorData[i].IsAlive.Barrier, elevator_network_list) {
			confirmedData.ElevatorData[i].IsAlive.Value = unconfinedData.ElevatorData[i].IsAlive.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.ElevatorData[i].IsFunctional.Barrier, elevator_network_list) {
			confirmedData.ElevatorData[i].IsFunctional.Value = unconfinedData.ElevatorData[i].IsFunctional.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.ElevatorData[i].Floor.Barrier, elevator_network_list) {
			confirmedData.ElevatorData[i].Floor.Value = unconfinedData.ElevatorData[i].Floor.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.ElevatorData[i].ElevatorBehaviour.Barrier, elevator_network_list) {
			confirmedData.ElevatorData[i].ElevatorBehaviour.Value = unconfinedData.ElevatorData[i].ElevatorBehaviour.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.ElevatorData[i].MotorDirection.Barrier, elevator_network_list) {
			confirmedData.ElevatorData[i].MotorDirection.Value = unconfinedData.ElevatorData[i].MotorDirection.Value
			isUpdated = true
		}
		for floor := 0; floor < N_FLOORS; floor++ {
			if CheckBarrier(unconfinedData.ElevatorData[i].CabRequests[floor].Barrier, elevator_network_list) {
				confirmedData.ElevatorData[i].CabRequests[floor].Value = unconfinedData.ElevatorData[i].CabRequests[floor].Value
				isUpdated = true
			}
		}
	}
	//Dont need Barrier check since the UpdateCC() have Barrier checks
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if unconfinedData.HallRequestData[floor][btn].Value != confirmedData.HallRequestData[floor][btn].Value {
				unconfinedData.HallRequestData[floor][btn] = confirmedData.HallRequestData[floor][btn]
				isUpdated = true
			}
		}
	}

	return confirmedData, isUpdated
}

func UpdateCC(oldCC RequestCyclicCounter_t, newCC RequestCyclicCounter_t, id int) RequestCyclicCounter_t {
	var updatedCC RequestCyclicCounter_t = oldCC

	if oldCC.Value == CC_Done && newCC.Value == CC_No {
		//Accept new value
		updatedCC.Value = newCC.Value
		updatedCC.Barrier = DeepCopyBarrier(newCC.Barrier)
		updatedCC.Barrier[id] = true
	} else if oldCC.Value == CC_No && newCC.Value == CC_Done {
		//Keep old value
		updatedCC.Value = oldCC.Value
		updatedCC.Barrier = DeepCopyBarrier(oldCC.Barrier)
	} else if oldCC.Value == newCC.Value {
		//They are the same, only update Barrier
		updatedCC.Barrier = BoolUnion(oldCC.Barrier, newCC.Barrier)
	} else if oldCC.Value < newCC.Value {
		//Accept new value
		updatedCC.Value = newCC.Value
		updatedCC.Barrier = DeepCopyBarrier(newCC.Barrier)
		updatedCC.Barrier[id] = true
	}
	return updatedCC
}

func LightCabLights(CabRequests []RequestCyclicCounter_t) {

	for floor := 0; floor < N_FLOORS; floor++ {
		SetButtonLamp(BT_Cab, floor, CCToBool(CabRequests[floor].Value))
	}
}

func LightHallLights(HallRequests [][2]RequestCyclicCounter_t) {
	for floor := 0; floor < N_FLOORS; floor++ {
		SetButtonLamp(BT_HallUp, floor, CCToBool(HallRequests[floor][BT_HallUp].Value))
		SetButtonLamp(BT_HallDown, floor, CCToBool(HallRequests[floor][BT_HallDown].Value))
	}
}

func CheckBarrier(Barrier ElevList_t, ElevAliveList ElevList_t) bool {
	for i := 0; i < N_ELEVATORS; i++ {
		if Barrier[i] != ElevAliveList[i] {
			return false
		}
	}
	return true
}

func BoolUnion(a []bool, b []bool) []bool {
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

func CCToBool(CC CyclicCounter_t) bool {
	if CC == CC_Uninit || CC == CC_No || CC == CC_Unconfirmed {
		return false
	}
	if CC == CC_Confirmed || CC == CC_Done {
		return true
	} else {
		print("wring CC Value")
		return false
	}
}

func DeepCopySystemData(src SystemData_t) SystemData_t {
	dst := src
	dst.ElevatorData = DeepCopyElevatorData(src.ElevatorData)
	dst.HallRequestData = DeepCopyHallRequests(src.HallRequestData)
	return dst
}

func DeepCopyElevatorData(src []ElevatorData_t) []ElevatorData_t {
	dst := make([]ElevatorData_t, len(src))

	for i := range src {
		dst[i] = DeepCopySingleElevatorData(src[i])
	}
	return dst
}

func DeepCopySingleElevatorData(src ElevatorData_t) ElevatorData_t {
	dst := src

	dst.IsAlive.Barrier = DeepCopyBarrier(src.IsAlive.Barrier)
	dst.IsFunctional.Barrier = DeepCopyBarrier(src.IsFunctional.Barrier)
	dst.Floor.Barrier = DeepCopyBarrier(src.Floor.Barrier)
	dst.ElevatorBehaviour.Barrier = DeepCopyBarrier(src.ElevatorBehaviour.Barrier)
	dst.MotorDirection.Barrier = DeepCopyBarrier(src.MotorDirection.Barrier)
	dst.CabRequests = DeepCopyCabRequests(src.CabRequests)
	return dst
}

func DeepCopyHallRequests(src [][2]RequestCyclicCounter_t) [][2]RequestCyclicCounter_t {
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

func DeepCopyCabRequests(src []RequestCyclicCounter_t) []RequestCyclicCounter_t {
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

func DeepCopyBarrier(src ElevList_t) ElevList_t {
	dst := make([]bool, len(src))
	copy(dst, src)
	return dst
}
