package message_sync

import (
	"the_project/elevator"
	"fmt"
	"time"
)

func Init_System_Data(local_id int) (System_Data_t, System_Data_t){

    var tmp_Cab_Requests []Request_Cyclic_Counter_t = make([]Request_Cyclic_Counter_t, elevator.N_FLOORS)

    for floor := 0; floor < elevator.N_FLOORS; floor++ {
        tmp_Cab_Requests[floor] = Request_Cyclic_Counter_t{
            Value:   CC_Uninit,
            barrier: make(Elev_List_t, N_ELEVATORS),
        }
    }

    var tmp_Hall_Request_Data [][2]Request_Cyclic_Counter_t = make([][2]Request_Cyclic_Counter_t, elevator.N_FLOORS)

    for floor := 0; floor < elevator.N_FLOORS; floor++ {
        for btn := 0; btn < 2; btn++ {
            tmp_Hall_Request_Data[floor][btn] = Request_Cyclic_Counter_t{
                Value:   CC_Uninit,
                barrier: make(Elev_List_t, N_ELEVATORS),
            }
        }
    }
	
	var tmp_Elevator_Data []Elevator_Data_t = make([]Elevator_Data_t, N_ELEVATORS)

    for i := 0; i < N_ELEVATORS; i++ {
        tmp_Elevator_Data[i] = Elevator_Data_t{
            Id:          i,
            Msg_counter: 0,
            Is_Alive: Is_Alive_Data_t{
                Value:   false,
                barrier: make(Elev_List_t, N_ELEVATORS),
            },
            Is_Able: Is_Able_Data_t{
                Value:   false,
                barrier: make(Elev_List_t, N_ELEVATORS),
            },
            Floor: Floor_Data_t{
                Value:   -1,
                barrier: make(Elev_List_t, N_ELEVATORS),
            },
            Elevator_Behaviour: Elevator_Behaviour_Data_t{
                Value:   elevator.EB_Idle,
                barrier: make(Elev_List_t, N_ELEVATORS),
            },
            Motor_Direction: Motor_Direction_Data_t{
                Value:   elevator.MD_Stop,
                barrier: make(Elev_List_t, N_ELEVATORS),
            },
            Cab_Requests: Deep_Copy_Cab_Requests(tmp_Cab_Requests),
        }
    }

	var system_data System_Data_t = System_Data_t{Id: local_id, Elevator_Data: tmp_Elevator_Data, Hall_Request_Data: tmp_Hall_Request_Data}

	return system_data, Deep_Copy_System_Data(system_data)
}

func On_Recieved_Fresh_Data(system_data System_Data_t, confirmed_system_data System_Data_t, fresh_data System_Data_t) (System_Data_t, System_Data_t, bool){
	var updated_system_data System_Data_t = Deep_Copy_System_Data(system_data)
	var updated_confirmed_system_data System_Data_t = Deep_Copy_System_Data(confirmed_system_data)
	var is_confirmed_data_updated bool = false

	for i := 0; i < N_ELEVATORS; i++{
		//if the new data have newer information about a elevator, we accept it
		if fresh_data.Elevator_Data[i].Msg_counter > system_data.Elevator_Data[i].Msg_counter {
			updated_system_data.Elevator_Data[i] = Update_Single_Elevator_Data(system_data.Elevator_Data[i], fresh_data.Elevator_Data[i], system_data.Id)
		}
	}

	//update the confirmed data that have recieved consensus
	updated_confirmed_system_data, is_confirmed_data_updated = Update_Confirmed_System_Data(system_data, confirmed_system_data)

	updated_system_data.Hall_Request_Data = Update_Hall_Request_Data(system_data.Hall_Request_Data, fresh_data.Hall_Request_Data, system_data.Id)
	if updated_confirmed_system_data.Hall_Request_Data != updated_system_data.Hall_Request_Data {
		updated_confirmed_system_data.Hall_Request_Data = updated_system_data.Hall_Request_Data
		is_confirmed_data_updated = true
	}
	
	return updated_system_data, updated_confirmed_system_data, is_confirmed_data_updated
}

func Update_Hall_Request_Data(old_data [][2]Request_Cyclic_Counter_t, new_data [][2]Request_Cyclic_Counter_t, id int) [][2]Request_Cyclic_Counter_t {
	var updated_hall_requests [][2]Request_Cyclic_Counter_t = Deep_Copy_Hall_Requests(old_data)
	
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			updated_hall_requests[floor][btn] = Update_CC(old_data[floor][btn], new_data[floor][btn], id)
		}
	}
	return updated_hall_requests
}

func Update_Single_Elevator_Data(old_data Elevator_Data_t, new_data Elevator_Data_t, id int) Elevator_Data_t{ 
	var updated_data Elevator_Data_t = Deep_Copy_Single_Elevator_Data(old_data)

	if old_data.Is_Alive.Value == new_data.Is_Alive.Value {
		 updated_data.Is_Alive.barrier = Bool_Union(old_data.Is_Alive.barrier, new_data.Is_Alive.barrier)
	} else {
		updated_data.Is_Alive = new_data.Is_Alive
		updated_data.Is_Alive.barrier[id] = true
	}

	if old_data.Is_Able.Value == new_data.Is_Able.Value {
		 updated_data.Is_Able.barrier = Bool_Union(old_data.Is_Able.barrier, new_data.Is_Able.barrier)
	} else {
		updated_data.Is_Able = new_data.Is_Able
		updated_data.Is_Able.barrier[id] = true
	}

	if old_data.Floor.Value == new_data.Floor.Value {
		 updated_data.Floor.barrier = Bool_Union(old_data.Floor.barrier, new_data.Floor.barrier)
	} else {
		updated_data.Floor = new_data.Floor
		updated_data.Floor.barrier[id] = true
	}

	if old_data.Elevator_Behaviour.Value == new_data.Elevator_Behaviour.Value {
		 updated_data.Elevator_Behaviour.barrier = Bool_Union(old_data.Elevator_Behaviour.barrier, new_data.Elevator_Behaviour.barrier)
	} else {
		updated_data.Elevator_Behaviour = new_data.Elevator_Behaviour
		updated_data.Elevator_Behaviour.barrier[id] = true
	}

	if old_data.Motor_Direction.Value == new_data.Motor_Direction.Value {
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
			confirmed_data.Elevator_Data[i].Is_Alive.Value = unconfirmed_data.Elevator_Data[i].Is_Alive.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Able.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Able.Value = unconfirmed_data.Elevator_Data[i].Is_Able.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Floor.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Floor.Value = unconfirmed_data.Elevator_Data[i].Floor.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Elevator_Behaviour.Value = unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Motor_Direction.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Motor_Direction.Value = unconfirmed_data.Elevator_Data[i].Motor_Direction.Value
			is_updated = true
		}
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if Check_Barrier(unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].barrier, elevator_network_list) {
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
		//Accept new value
		updated_CC.Value = new_CC.Value
		updated_CC.barrier = Deep_Copy_Barrier(new_CC.barrier)
		updated_CC.barrier[id] = true
	} else if old_CC.Value == CC_No && new_CC.Value == CC_Done{
		//Keep old value
		updated_CC.Value = old_CC.Value
		updated_CC.barrier = Deep_Copy_Barrier(old_CC.barrier)
	} else if old_CC.Value == new_CC.Value{
		//They are the same, only update barrier
		updated_CC.barrier = Bool_Union(old_CC.barrier, new_CC.barrier)
	} else if old_CC.Value < new_CC.Value {
		//Accept new value
		updated_CC.Value = new_CC.Value
		updated_CC.barrier = Deep_Copy_Barrier(new_CC.barrier)
		updated_CC.barrier[id] = true
	}
	return updated_CC
}

func Light_Cab_Lights(Cab_Requests []Request_Cyclic_Counter_t){

	for floor := 0; floor < elevator.N_FLOORS; floor++{
		elevator.SetButtonLamp(elevator.BT_Cab, floor, CC_To_Bool(Cab_Requests[floor].Value))
	}
}

func Light_Hall_Lights(Hall_Requests [][2]Request_Cyclic_Counter_t){
	for floor := 0; floor < elevator.N_FLOORS; floor++{
		elevator.SetButtonLamp(elevator.BT_HallUp, floor, CC_To_Bool(Hall_Requests[floor][elevator.BT_HallUp].Value))
		elevator.SetButtonLamp(elevator.BT_HallDown, floor, CC_To_Bool(Hall_Requests[floot][elevator.BT_HallDown].Value))
	}
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
	if (CC == CC_Uninit || CC == CC_No || CC == CC_Unconfirmed) {
		return false
	}
	if CC == CC_Confirmed || CC == CC_Done {
		return true
	} else {
		print("wring CC Value")
		return false
	}
}

func Deep_Copy_Cab_Requests(src []Request_Cyclic_Counter_t) []Request_Cyclic_Counter_t {
    dst := make([]Request_Cyclic_Counter_t, len(src))
    for i := range src {
        dst[i] = src[i]
        if src[i].barrier != nil {
            barrierCopy := make([]bool, len(src[i].barrier))
            copy(barrierCopy, src[i].barrier)
            dst[i].barrier = barrierCopy
        }
    }
    return dst
}

func Deep_Copy_Hall_Requests(src [][2]Request_Cyclic_Counter_t) [][2]Request_Cyclic_Counter_t {
    dst := make([][2]Request_Cyclic_Counter_t, len(src))

    for floor := range src {
        for btn := 0; btn < 2; btn++ {
            dst[floor][btn] = src[floor][btn]

			barrierCopy := make([]bool, len(src[floor][btn].barrier))
			copy(barrierCopy, src[floor][btn].barrier)
			dst[floor][btn].barrier = barrierCopy
        }
    }
    return dst
}

func Deep_Copy_Barrier(src Elev_List_t) Elev_List_t{
	dst := make([]bool, len(src))
	copy(dst, src)
	return dst
}

func Deep_Copy_Single_Elevator_Data(src Elevator_Data_t) Elevator_Data_t{
	 dst := src

    dst.Is_Alive.barrier = Deep_Copy_Barrier(src.Is_Alive.barrier)
    dst.Is_Able.barrier = Deep_Copy_Barrier(src.Is_Able.barrier)
    dst.Floor.barrier = Deep_Copy_Barrier(src.Floor.barrier)
    dst.Elevator_Behaviour.barrier = Deep_Copy_Barrier(src.Elevator_Behaviour.barrier)
    dst.Motor_Direction.barrier = Deep_Copy_Barrier(src.Motor_Direction.barrier)
	dst.Cab_Requests = Deep_Copy_Cab_Requests(src.Cab_Requests)
	return dst
}

func Deep_Copy_Elevator_data(src []Elevator_Data_t) []Elevator_Data_t{
	dst := make([]Elevator_Data_t, len(src))

	for i := range src {
		dst[i] = Deep_Copy_Single_Elevator_Data(src[i])
	}
	return dst
}

func Deep_Copy_System_Data(src System_Data_t)System_Data_t{
	dst := src 
	dst.Elevator_Data = Deep_Copy_Elevator_data(src.Elevator_Data)
	dst.Hall_Request_Data = Deep_Copy_Hall_Requests(src.Hall_Request_Data)
	return dst
}