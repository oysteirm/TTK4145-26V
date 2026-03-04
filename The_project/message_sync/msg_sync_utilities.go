package message_sync

import (
	"../elevator"
	"fmt"
	"time"
)

func Init_System_Data(local_id int) (System_Data_t, System_Data_t){

	var tmp_Cab_Requests []Request_Cyclic_Counter_t
	var tmp_Hall_Request_Data [][2]Request_Cyclic_Counter_t
	for floor := 0; floor < elevator.N_FLOORS; floor++{
		Cab_Requests[i] =  Request_Cyclic_Counter_t{Value: CC_Uninit, Barrier: Elev_List_t{}}
		Hall_Request_Data[i][0] =  Request_Cyclic_Counter_t{Value: CC_Uninit, Barrier: Elev_List_t{}}
		Hall_Request_Data[i][1] =  Request_Cyclic_Counter_t{Value: CC_Uninit, Barrier: Elev_List_t{}}
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

	if old_data.Is_Alive.Value == new_data.Is_Alive.Value {
		 updated_data.Is_Alive.Barrier = Bool_Union(old_data.Is_Alive.Barrier, new_data.Is_Alive.Barrier)
	} else {
		updated_data.Is_Alive = new_data.Is_Alive
		updated_data.Is_Alive.Barrier[id] = true
	}

	if old_data.Is_Able.Value == new_data.Is_Able.Value {
		 updated_data.Is_Able.Barrier = Bool_Union(old_data.Is_Able.Barrier, new_data.Is_Able.Barrier)
	} else {
		updated_data.Is_Able = new_data.Is_Able
		updated_data.Is_Able.Barrier[id] = true
	}

	if old_data.Floor.Value == new_data.Floor.Value {
		 updated_data.Floor.Barrier = Bool_Union(old_data.Floor.Barrier, new_data.Floor.Barrier)
	} else {
		updated_data.Floor = new_data.Floor
		updated_data.Floor.Barrier[id] = true
	}

	if old_data.Elevator_Behaviour.Value == new_data.Elevator_Behaviour.Value {
		 updated_data.Elevator_Behaviour.Barrier = Bool_Union(old_data.Elevator_Behaviour.Barrier, new_data.Elevator_Behaviour.Barrier)
	} else {
		updated_data.Elevator_Behaviour = new_data.Elevator_Behaviour
		updated_data.Elevator_Behaviour.Barrier[id] = true
	}

	if old_data.Motor_Direction.Value == new_data.Motor_Direction.Value {
		 updated_data.Motor_Direction.Barrier = Bool_Union(old_data.Motor_Direction.Barrier, new_data.Motor_Direction.Barrier)
	} else {
		updated_data.Motor_Direction = new_data.Motor_Direction
		updated_data.Motor_Direction.Barrier[id] = true
	}

	for i := 0; i < N_ELEVATORS; i++{
		updated_data.Cab_Requests[i] = Update_CC(old_data.Cab_Requests[i], new_data.Cab_Requests[i], id)
	}

	return updated_data
}

func Update_Confirmed_System_Data(unconfirmed_data System_Data_t, confirmed_data System_Data_t) (System_Data_t, bool){
	var is_updated bool = false
	
	for i := 0; i < N_ELEVATORS; i++ {
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Alive.Barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Alive.Value = unconfirmed_data.Elevator_Data[i].Is_Alive.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Able.Barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Able.Value = unconfirmed_data.Elevator_Data[i].Is_Able.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Floor.Barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Floor.Value = unconfirmed_data.Elevator_Data[i].Floor.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.Barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Elevator_Behaviour.Value = unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Motor_Direction.Barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Motor_Direction.Value = unconfirmed_data.Elevator_Data[i].Motor_Direction.Value
			is_updated = true
		}
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if Check_Barrier(unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].Barrier, elevator_network_list) {
				confirmed_data.Elevator_Data[i].Cab_Requests[floor].Value = unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].Value
				is_updated = true
			}
		}
	}
	return confirmed_data, is_updated
}

func Update_CC(old_CC Request_Cyclic_Counter_t, new_CC Request_Cyclic_Counter_t, id int) Request_Cyclic_Counter_t{
	 var updated_CC Request_Cyclic_Counter_t = old_CC

	if old_CC.Value == CC_Done && new_CC.Value == CC_No{
		updated_CC = new_CC
		updated_CC.Barrier[id] = 1
	} 
	else if old_CC.Value == CC_No && new_CC.Value == CC_Done{
		updated_CC = old_CC
	} 
	else if old_CC.Value == new_CC.Value{
		old_CC.Barrier = Bool_Union(old_CC.Barrier, new_CC.Barrier)
	}
	else if old_CC.Value < new_CC.Value {
		updated_CC = new_CC
		updated_CC.Barrier[id] = 1
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