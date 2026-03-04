package message_sync

import (
	"the_project/elevator"
)

func InitSystemData(local_id int) (SystemData_t, SystemData_t) {

	var tmpCabRequests []RequestCyclicCounter_t = make([]RequestCyclicCounter_t, elevator.N_FLOORS)

	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		tmpCabRequests[floor] = RequestCyclicCounter_t{
			Value:   CC_Uninit,
			Barrier: make(ElevList_t, N_ELEVATORS),
		}
	}

	var tmpHallRequestData [][2]RequestCyclicCounter_t = make([][2]RequestCyclicCounter_t, elevator.N_FLOORS)

	for floor := 0; floor < elevator.N_FLOORS; floor++ {
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
			Is_Alive: IsAliveData_t{
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
			Elevator_Behaviour: ElevatorBehaviourData_t{
				Value:   elevator.EB_Idle,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			Motor_Direction: MotorDirectionData_t{
				Value:   elevator.MD_Stop,
				Barrier: make(ElevList_t, N_ELEVATORS),
			},
			Cab_Requests: DeepCopyCabRequests(tmpCabRequests),
		}
	}

	var systemData SystemData_t = SystemData_t{Id: local_id, ElevatorData: tmpElevatorData, HallRequestData: tmpHallRequestData}

	return systemData, DeepCopySystemData(systemData)
}

func OnReceivedFreshData(systemData SystemData_t, confirmedSystemData SystemData_t, freshData SystemData_t) (SystemData_t, SystemData_t, bool) {
	var updatedSystemData SystemData_t = DeepCopySystemData(systemData)
	var updatedConfirmedSystemData SystemData_t = DeepCopySystemData(confirmedSystemData)
	var isConfirmedDataUpdated bool = false

	for i := 0; i < N_ELEVATORS; i++ {
		//if the new data have newer information about a elevator, we accept it

		if systemData.Elevator_Data[i].Id == freshData.Id {
			updatedSystemData.Elevator_Data[i] = UpdateElevatorDataAboutSelf(systemData.Elevator_Data[i], freshData.Elevator_Data[i], systemData.Id)
		} else {
			updatedSystemData.Elevator_Data[i] = UpdateElevatorDataAboutOther(systemData.Elevator_Data[i], freshData.Elevator_Data[i], systemData.Id)
		}
	}
	//update hall requests with the cyclic counter
	updatedSystemData.Hall_Request_Data = UpdateHallRequestData(systemData.Hall_Request_Data, freshData.Hall_Request_Data, systemData.Id)

	//update the confirmed data that have recieved consensus
	updatedConfirmedSystemData, isConfirmedDataUpdated = UpdateConfirmedSystemData(systemData, confirmedSystemData)

	return updatedSystemData, updatedConfirmedSystemData, isConfirmedDataUpdated
}

func UpdateHallRequestData(oldData [][2]RequestCyclicCounter_t, newData [][2]RequestCyclicCounter_t, id int) [][2]RequestCyclicCounter_t {
	var updatedHallRequests [][2]RequestCyclicCounter_t = DeepCopyHallRequests(oldData)

	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			updatedHallRequests[floor][btn] = UpdateCC(oldData[floor][btn], newData[floor][btn], id)
		}
	}
	return updatedHallRequests
}

// We trust info an elevator tells about itself.
func UpdateElevatorDataAboutSelf(oldData ElevatorData_t, newData ElevatorData_t, id int) ElevatorData_t {
	var updatedData ElevatorData_t = DeepCopySingleElevatorData(oldData)

	if oldData.Is_Alive.Value == newData.Is_Alive.Value {
		updatedData.Is_Alive.Barrier = BoolUnion(oldData.Is_Alive.Barrier, newData.Is_Alive.Barrier)
	} else {
		updatedData.Is_Alive = newData.Is_Alive
		updatedData.Is_Alive.Barrier[id] = true
	}

	if oldData.Is_Functional.Value == newData.Is_Functional.Value {
		updatedData.Is_Functional.Barrier = BoolUnion(oldData.Is_Functional.Barrier, newData.Is_Functional.Barrier)
	} else {
		updatedData.Is_Functional = newData.Is_Functional
		updatedData.Is_Functional.Barrier[id] = true
	}

	if oldData.Floor.Value == newData.Floor.Value {
		updatedData.Floor.Barrier = BoolUnion(oldData.Floor.Barrier, newData.Floor.Barrier)
	} else {
		updatedData.Floor = newData.Floor
		updatedData.Floor.Barrier[id] = true
	}

	if oldData.Elevator_Behaviour.Value == newData.Elevator_Behaviour.Value {
		updatedData.Elevator_Behaviour.Barrier = BoolUnion(oldData.Elevator_Behaviour.Barrier, newData.Elevator_Behaviour.Barrier)
	} else {
		updatedData.Elevator_Behaviour = newData.Elevator_Behaviour
		updatedData.Elevator_Behaviour.Barrier[id] = true
	}

	if oldData.Motor_Direction.Value == newData.Motor_Direction.Value {
		updatedData.Motor_Direction.Barrier = BoolUnion(oldData.Motor_Direction.Barrier, newData.Motor_Direction.Barrier)
	} else {
		updatedData.Motor_Direction = newData.Motor_Direction
		updatedData.Motor_Direction.Barrier[id] = true
	}

	for i := 0; i < N_ELEVATORS; i++ {
		updatedData.Cab_Requests[i] = UpdateCC(oldData.Cab_Requests[i], newData.Cab_Requests[i], id)
	}

	return updatedData
}

// Only update cabcalls CC or update Barriers
func UpdateElevatorDataAboutOther(oldData ElevatorData_t, newData ElevatorData_t, id int) ElevatorData_t {
	var updatedData ElevatorData_t = DeepCopySingleElevatorData(oldData)

	if oldData.Is_Alive.Value == newData.Is_Alive.Value {
		updatedData.Is_Alive.Barrier = BoolUnion(oldData.Is_Alive.Barrier, newData.Is_Alive.Barrier)
	}

	if oldData.Is_Functional.Value == newData.Is_Functional.Value {
		updatedData.Is_Functional.Barrier = BoolUnion(oldData.Is_Functional.Barrier, newData.Is_Functional.Barrier)
	}

	if oldData.Floor.Value == newData.Floor.Value {
		updatedData.Floor.Barrier = BoolUnion(oldData.Floor.Barrier, newData.Floor.Barrier)
	}

	if oldData.Elevator_Behaviour.Value == newData.Elevator_Behaviour.Value {
		updatedData.Elevator_Behaviour.Barrier = BoolUnion(oldData.Elevator_Behaviour.Barrier, newData.Elevator_Behaviour.Barrier)
	}

	if oldData.Motor_Direction.Value == newData.Motor_Direction.Value {
		updatedData.Motor_Direction.Barrier = BoolUnion(oldData.Motor_Direction.Barrier, newData.Motor_Direction.Barrier)
	}

	for i := 0; i < N_ELEVATORS; i++ {
		updatedData.Cab_Requests[i] = UpdateCC(oldData.Cab_Requests[i], newData.Cab_Requests[i], id)
	}

	return updatedData
}

func UpdateConfirmedSystemData(unconfinedData SystemData_t, confirmedData SystemData_t) (SystemData_t, bool) {
	var isUpdated bool = false

	for i := 0; i < N_ELEVATORS; i++ {
		if CheckBarrier(unconfinedData.Elevator_Data[i].Is_Alive.Barrier, elevator_network_list) {
			confirmedData.Elevator_Data[i].Is_Alive.Value = unconfinedData.Elevator_Data[i].Is_Alive.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.Elevator_Data[i].Is_Functional.Barrier, elevator_network_list) {
			confirmedData.Elevator_Data[i].Is_Functional.Value = unconfinedData.Elevator_Data[i].Is_Functional.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.Elevator_Data[i].Floor.Barrier, elevator_network_list) {
			confirmedData.Elevator_Data[i].Floor.Value = unconfinedData.Elevator_Data[i].Floor.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.Elevator_Data[i].Elevator_Behaviour.Barrier, elevator_network_list) {
			confirmedData.Elevator_Data[i].Elevator_Behaviour.Value = unconfinedData.Elevator_Data[i].Elevator_Behaviour.Value
			isUpdated = true
		}
		if CheckBarrier(unconfinedData.Elevator_Data[i].Motor_Direction.Barrier, elevator_network_list) {
			confirmedData.Elevator_Data[i].Motor_Direction.Value = unconfinedData.Elevator_Data[i].Motor_Direction.Value
			isUpdated = true
		}
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if CheckBarrier(unconfinedData.Elevator_Data[i].Cab_Requests[floor].Barrier, elevator_network_list) {
				confirmedData.Elevator_Data[i].Cab_Requests[floor].Value = unconfinedData.Elevator_Data[i].Cab_Requests[floor].Value
				isUpdated = true
			}
		}
	}
	//Dont need Barrier check since the UpdateCC() have Barrier checks
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if unconfinedData.Hall_Request_Data[floor][btn].Value != confirmedData.Hall_Request_Data[floor][btn].Value {
				unconfinedData.Hall_Request_Data[floor][btn] = confirmedData.Hall_Request_Data[floor][btn]
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

	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		elevator.SetButtonLamp(elevator.BT_Cab, floor, CCToBool(CabRequests[floor].Value))
	}
}

func LightHallLights(HallRequests [][2]RequestCyclicCounter_t) {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		elevator.SetButtonLamp(elevator.BT_HallUp, floor, CCToBool(HallRequests[floor][elevator.BT_HallUp].Value))
		elevator.SetButtonLamp(elevator.BT_HallDown, floor, CCToBool(HallRequests[floor][elevator.BT_HallDown].Value))
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
	dst.Elevator_Data = DeepCopyElevatorData(src.Elevator_Data)
	dst.Hall_Request_Data = DeepCopyHallRequests(src.Hall_Request_Data)
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

	dst.Is_Alive.Barrier = DeepCopyBarrier(src.Is_Alive.Barrier)
	dst.Is_Functional.Barrier = DeepCopyBarrier(src.Is_Functional.Barrier)
	dst.Floor.Barrier = DeepCopyBarrier(src.Floor.Barrier)
	dst.Elevator_Behaviour.Barrier = DeepCopyBarrier(src.Elevator_Behaviour.Barrier)
	dst.Motor_Direction.Barrier = DeepCopyBarrier(src.Motor_Direction.Barrier)
	dst.Cab_Requests = DeepCopyCabRequests(src.Cab_Requests)
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
