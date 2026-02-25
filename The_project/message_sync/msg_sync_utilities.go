package message_sync

import (
	"TTK4145-26V/elevator"
	"fmt"
	"time"
)

func Init_System_Data(local_id int) (System_Data_t, System_Data_t){

	var tmp_Cab_Requests []Request_Cyclic_Counter_t
	var tmp_Hall_Request_Data [][2]Request_Cyclic_Counter_t
	for floor := 0; floor < elevator.N_FLOORS; floor++{
		Cab_Requests[i] =  Request_Cyclic_Counter_t{value: CC_Uninit, barrier: Elev_List_t{}}
		Hall_Request_Data[i][0] =  Request_Cyclic_Counter_t{value: CC_Uninit, barrier: Elev_List_t{}}
		Hall_Request_Data[i][1] =  Request_Cyclic_Counter_t{value: CC_Uninit, barrier: Elev_List_t{}}
	}
	
	var tmp_Elevator_Data []Elevator_Data_t
	for i := 0; i < N_ELEVATORS; i++{
		Elevator_Data[i] = Elevator_Data_t{Id: i, Msg_counter: 0, Is_Alive: 0, Is_Able: 0, Floor: -1, Elevator_Behaviour: elevator.EB_Idle, Motor_Direction: elevator.MD_Stop, Cab_Requests: tmp_Cab_Requests}
	}

	var system_data System_Data_t = System_Data_t{Id: local_id, Elevator_Data: tmp_Elevator_Data, tmp_Hall_Request_Data}

	return system_data, system_data
}

func On_Recieved_Fresh_Data(system_data System_Data_t, confirmed_system_data System_Data_t, fresh_data System_Data_t) (System_Data_t, System_Data_t, bool){
	var updated_system_data System_Data_t = system_data
	var updated_confirmed_system_data System_Data_t = confirmed_system_data
	var is_confirmed_data_updated bool = false

	for i := 0; i < N_ELEVATORS; i++{
		//if the new data have newer information about a elevator, we accept it
		if fresh_data.Elevator_Data[i].Msg_counter > system_data.Elevator_Data[i].Msg_counter {
			updated_system_data.Elevator_Data[i] = Update_Single_Elevator_Data(system_data.Elevator_Data[i], fresh_data.Elevator_Data[i], system_data.Id)
		}
	}

	system_data.Hall_Request_Data = Update_Hall_Request_Data(system_data.Hall_Request_Data, fresh_data.Hall_Request_Data, system_data.Id)

	//update the confirmed data that have recieved consensus
	if updated_confirmed_system_data.Hall_Request_Data != updated_system_data.Hall_Request_Data {
		updated_confirmed_system_data.Hall_Request_Data = updated_system_data.Hall_Request_Data
		is_confirmed_data_updated = true
	}
	updated_confirmed_system_data, is_confirmed_data_updated = Update_Confirmed_System_Data(system_data, confirmed_system_data)
	
	return updated_system_data, updated_confirmed_system_data, is_confirmed_data_updated
}

func Update_Hall_Request_Data(old_data [][2]Request_Cyclic_Counter_t, new_data [][2]Request_Cyclic_Counter_t, id int) [][2]Request_Cyclic_Counter_t {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			Update_CC(old_data[floor][btn], new_data[floor][btn], id)
		}
	}
}

func Update_Single_Elevator_Data(old_data Elevator_Data_t, new_data Elevator_Data_t, id int) Elevator_Data_t{ 
	var updated_data Elevator_Data_t = old_data

	if old_data.Is_Alive.value == new_data.Is_Alive.value {
		 updated_data.Is_Alive.barrier = Bool_Union(old_data.Is_Alive.barrier, new_data.Is_Alive.barrier)
	} else {
		updated_data.Is_Alive = new_data.Is_Alive
		updated_data.Is_Alive.barrier[id] = true
	}

	if old_data.Is_Able.value == new_data.Is_Able.value {
		 updated_data.Is_Able.barrier = Bool_Union(old_data.Is_Able.barrier, new_data.Is_Able.barrier)
	} else {
		updated_data.Is_Able = new_data.Is_Able
		updated_data.Is_Able.barrier[id] = true
	}

	if old_data.Floor.value == new_data.Floor.value {
		 updated_data.Floor.barrier = Bool_Union(old_data.Floor.barrier, new_data.Floor.barrier)
	} else {
		updated_data.Floor = new_data.Floor
		updated_data.Floor.barrier[id] = true
	}

	if old_data.Elevator_Behaviour.value == new_data.Elevator_Behaviour.value {
		 updated_data.Elevator_Behaviour.barrier = Bool_Union(old_data.Elevator_Behaviour.barrier, new_data.Elevator_Behaviour.barrier)
	} else {
		updated_data.Elevator_Behaviour = new_data.Elevator_Behaviour
		updated_data.Elevator_Behaviour.barrier[id] = true
	}

	if old_data.Motor_Direction.value == new_data.Motor_Direction.value {
		 updated_data.Motor_Direction.barrier = Bool_Union(old_data.Motor_Direction.barrier, new_data.Motor_Direction.barrier)
	} else {
		updated_data.Motor_Direction = new_data.Motor_Direction
		updated_data.Motor_Direction.barrier[id] = true
	}

	for i := 0; i < N_ELEVATORS; i++{
		updated_data.Cab_Requests[i] = Update_CC(old_data.Cab_Requests[i], new_data.Cab_Requests[i], id)
	}

	return updated_data
}

func Update_Confirmed_System_Data(unconfirmed_data System_Data_t, confirmed_data System_Data_t) (System_Data_t, bool){
	var is_updated bool = false
	
	for i := 0; i < N_ELEVATORS; i++ {
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Alive.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Alive.value = unconfirmed_data.Elevator_Data[i].Is_Alive.value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Able.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Able.value = unconfirmed_data.Elevator_Data[i].Is_Able.value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Floor.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Floor.value = unconfirmed_data.Elevator_Data[i].Floor.value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Elevator_Behaviour.value = unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Motor_Direction.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Motor_Direction.value = unconfirmed_data.Elevator_Data[i].Motor_Direction.value
			is_updated = true
		}
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if Check_Barrier(unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].barrier, elevator_network_list) {
				confirmed_data.Elevator_Data[i].Cab_Requests[floor].value = unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].value
				is_updated = true
			}
		}
	}
	return confirmed_data, is_updated
}

func Update_CC(old_CC Request_Cyclic_Counter_t, new_CC Request_Cyclic_Counter_t, id int) Request_Cyclic_Counter_t{
	 var updated_CC Request_Cyclic_Counter_t = old_CC

	if old_CC.value == CC_Done && new_CC.value == CC_No{
		updated_CC = new_CC
		updated_CC.barrier[id] = 1
	} 
	else if old_CC.value == CC_No && new_CC.value == CC_Done{
		updated_CC = old_CC
	} 
	else if old_CC.value == new_CC.value{
		old_CC.barrier = Bool_Union(old_CC.barrier, new_CC.barrier)
	}
	else if old_CC.value < new_CC.value {
		updated_CC = new_CC
		updated_CC.barrier[id] = 1
	}

	return updated_CC
}


func Check_Barrier(barrier Elev_List_t, Elev_Alive_List Elev_List_t)bool{
	for i := 0; i < N_ELEVATORS; i++{
		if barrier[i] != Elev_Alive_List[i]{
			return false
		}
	}
	return true
}

func Bool_Union(a []bool, b []bool) []bool {
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

func CC_To_Bool(CC Cyclic_Counter_t)bool{
	if CC == CC_Uninit || CC == CC_No || CC = CC_Unconfirmed {
		return false
	}
	if CC == CC_Confirmed || CC == CC_Done {
		return true
	}
	else {
		return nil
	}
}