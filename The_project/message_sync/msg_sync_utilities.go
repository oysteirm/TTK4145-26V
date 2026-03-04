package message_sync

import (
	"fmt"
	"strconv"
	"the_project/Network_Driver/peers"
	"the_project/elevator"
)

//Initalizing the the system_data and confirmed_system_data in Message_Sync_Server
//All values are initialized to 0, -1 (not in a floor, EB_Idle, CC_Uninit and empty barriers
func Init_System_Data(local_id int) (System_Data_t, System_Data_t){

    var tmp_Cab_Requests []Request_Cyclic_Counter_t = make([]Request_Cyclic_Counter_t, elevator.N_FLOORS)

    for floor := 0; floor < elevator.N_FLOORS; floor++ {
        tmp_Cab_Requests[floor] = Request_Cyclic_Counter_t{
            Value:   CC_Uninit,
            Barrier: make([]bool, N_ELEVATORS),
        }
    }

    var tmp_Hall_Request_Data [][2]Request_Cyclic_Counter_t = make([][2]Request_Cyclic_Counter_t, elevator.N_FLOORS)

    for floor := 0; floor < elevator.N_FLOORS; floor++ {
        for btn := 0; btn < 2; btn++ {
            tmp_Hall_Request_Data[floor][btn] = Request_Cyclic_Counter_t{
                Value:   CC_Uninit,
                Barrier: make([]bool, N_ELEVATORS),
            }
        }
    }
	
	var tmp_Elevator_Data []Elevator_Data_t = make([]Elevator_Data_t, N_ELEVATORS)

    for i := 0; i < N_ELEVATORS; i++ {
        tmp_Elevator_Data[i] = Elevator_Data_t{
            Id:          i,
            //Msg_counter: 0,
            Is_Alive: Is_Alive_Data_t{
                Value:   false,
                Barrier: make([]bool, N_ELEVATORS),
            },
            Is_Functional: Is_Functional_Data_t{
                Value:   false,
                Barrier: make([]bool, N_ELEVATORS),
            },
            Floor: Floor_Data_t{
                Value:   -1,
                Barrier: make([]bool, N_ELEVATORS),
            },
            Elevator_Behaviour: Elevator_Behaviour_Data_t{
                Value:   elevator.EB_Idle,
                Barrier: make([]bool, N_ELEVATORS),
            },
            Motor_Direction: Motor_Direction_Data_t{
                Value:   elevator.MD_Stop,
                Barrier: make([]bool, N_ELEVATORS),
            },
            Cab_Requests: Deep_Copy_Cab_Requests(tmp_Cab_Requests),
        }
    }

	var system_data System_Data_t = System_Data_t{Id: local_id, Elevator_Data: tmp_Elevator_Data, Hall_Request_Data: tmp_Hall_Request_Data}

	return system_data, Deep_Copy_System_Data(system_data)
}

//Processing the fresh data and undating system_data and confirmed_system_data accordingly
func On_Received_Fresh_Data(system_data System_Data_t, 
							confirmed_system_data System_Data_t, 
							fresh_data System_Data_t) (System_Data_t, System_Data_t, bool){

	var updated_system_data System_Data_t = Deep_Copy_System_Data(system_data)
	var updated_confirmed_system_data System_Data_t = Deep_Copy_System_Data(confirmed_system_data)
	var is_confirmed_data_updated bool = false

	for i := 0; i < N_ELEVATORS; i++{
		//if the new data have newer information about a elevator, we accept it

		if system_data.Elevator_Data[i].Id == fresh_data.Id {
			updated_system_data.Elevator_Data[i] = Update_Elevator_Data_About_Self(system_data.Elevator_Data[i], fresh_data.Elevator_Data[i], system_data.Id)
		} else {
			updated_system_data.Elevator_Data[i] = Update_Elevator_Data_About_Other(system_data.Elevator_Data[i], fresh_data.Elevator_Data[i], system_data.Id)
		}
	}
	//update hall requests with the cyclic counter 
	updated_system_data.Hall_Request_Data = Update_Hall_Request_Data(system_data.Hall_Request_Data, fresh_data.Hall_Request_Data, system_data.Id)

	//update the confirmed data that have recieved consensus
	updated_confirmed_system_data, is_confirmed_data_updated = Update_Confirmed_System_Data(system_data, confirmed_system_data)
	
	return updated_system_data, updated_confirmed_system_data, is_confirmed_data_updated
}

//Functions for safely updating the system data
//-----------------------------------------------------------
func Update_Hall_Request_Data(	old_data [][2]Request_Cyclic_Counter_t, 
								new_data [][2]Request_Cyclic_Counter_t, 
								id int) [][2]Request_Cyclic_Counter_t {

	var updated_hall_requests [][2]Request_Cyclic_Counter_t = Deep_Copy_Hall_Requests(old_data)
	
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			updated_hall_requests[floor][btn] = Update_CC(old_data[floor][btn], new_data[floor][btn], id)
		}
	}
	return updated_hall_requests
}

//We trust info an elevator tells about itself. 
//If the data is the same: update barriers
//If the data is not the same: accept new data
func Update_Elevator_Data_About_Self(	old_data Elevator_Data_t, 
										new_data Elevator_Data_t, 
										id int) Elevator_Data_t { 

	var updated_data Elevator_Data_t = Deep_Copy_Single_Elevator_Data(old_data)

	if old_data.Is_Alive.Value == new_data.Is_Alive.Value {
		 updated_data.Is_Alive.Barrier = Bool_Union(old_data.Is_Alive.Barrier, new_data.Is_Alive.Barrier)
	} else {
		updated_data.Is_Alive = new_data.Is_Alive
		updated_data.Is_Alive.Barrier[id] = true
	}

	if old_data.Is_Functional.Value == new_data.Is_Functional.Value {
		 updated_data.Is_Functional.Barrier = Bool_Union(old_data.Is_Functional.Barrier, new_data.Is_Functional.Barrier)
	} else {
		updated_data.Is_Functional = new_data.Is_Functional
		updated_data.Is_Functional.Barrier[id] = true
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

//Only update cab requests CC and update barriers
func Update_Elevator_Data_About_Other(	old_data Elevator_Data_t, 
										new_data Elevator_Data_t, 
										id int) Elevator_Data_t {

	var updated_elevator_data Elevator_Data_t = Deep_Copy_Single_Elevator_Data(old_data)

	if old_data.Is_Alive.Value == new_data.Is_Alive.Value {
		 updated_elevator_data.Is_Alive.Barrier = Bool_Union(old_data.Is_Alive.Barrier, new_data.Is_Alive.Barrier)
	} 

	if old_data.Is_Functional.Value == new_data.Is_Functional.Value {
		 updated_elevator_data.Is_Functional.Barrier = Bool_Union(old_data.Is_Functional.Barrier, new_data.Is_Functional.Barrier)
	} 

	if old_data.Floor.Value == new_data.Floor.Value {
		 updated_elevator_data.Floor.Barrier = Bool_Union(old_data.Floor.Barrier, new_data.Floor.Barrier)
	} 

	if old_data.Elevator_Behaviour.Value == new_data.Elevator_Behaviour.Value {
		 updated_elevator_data.Elevator_Behaviour.Barrier = Bool_Union(old_data.Elevator_Behaviour.Barrier, new_data.Elevator_Behaviour.Barrier)
	} 

	if old_data.Motor_Direction.Value == new_data.Motor_Direction.Value {
		 updated_elevator_data.Motor_Direction.Barrier = Bool_Union(old_data.Motor_Direction.Barrier, new_data.Motor_Direction.Barrier)
	} 
	
	for i := 0; i < N_ELEVATORS; i++{
		updated_elevator_data.Cab_Requests[i] = Update_CC(old_data.Cab_Requests[i], new_data.Cab_Requests[i], id)
	}

	return updated_elevator_data 
}

//Checking the Barrier 
func Update_Confirmed_System_Data(	unconfirmed_data System_Data_t, 
									confirmed_data System_Data_t) (System_Data_t, bool) {

	var is_updated bool = false
	
	for i := 0; i < N_ELEVATORS; i++ {
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Alive.Barrier) {
			confirmed_data.Elevator_Data[i].Is_Alive.Value = unconfirmed_data.Elevator_Data[i].Is_Alive.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Functional.Barrier) {
			confirmed_data.Elevator_Data[i].Is_Functional.Value = unconfirmed_data.Elevator_Data[i].Is_Functional.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Floor.Barrier) {
			confirmed_data.Elevator_Data[i].Floor.Value = unconfirmed_data.Elevator_Data[i].Floor.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.Barrier) {
			confirmed_data.Elevator_Data[i].Elevator_Behaviour.Value = unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.Value
			is_updated = true
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Motor_Direction.Barrier) {
			confirmed_data.Elevator_Data[i].Motor_Direction.Value = unconfirmed_data.Elevator_Data[i].Motor_Direction.Value
			is_updated = true
		}
		//Dont need Barrier check since Update_CC() have Barrier checks 
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].Value != confirmed_data.Elevator_Data[i].Cab_Requests[floor].Value {
				confirmed_data.Elevator_Data[i].Cab_Requests[floor].Value = unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].Value
				is_updated = true
			}
		}
	}

	//Dont need Barrier check since Update_CC() have Barrier checks 
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			if unconfirmed_data.Hall_Request_Data[floor][btn].Value != confirmed_data.Hall_Request_Data[floor][btn].Value {
				unconfirmed_data.Hall_Request_Data[floor][btn] = confirmed_data.Hall_Request_Data[floor][btn]
				is_updated = true
			}
		}
	}

	return confirmed_data, is_updated
}

func Update_CC(	old_CC Request_Cyclic_Counter_t, 
				new_CC Request_Cyclic_Counter_t, 
				Id int) Request_Cyclic_Counter_t {

	 var updated_CC Request_Cyclic_Counter_t = old_CC

	//update the CC based on rules
	if old_CC.Value == CC_Done && new_CC.Value == CC_No{
		//Accept new value
		updated_CC.Value = new_CC.Value
		updated_CC.Barrier = Deep_Copy_Barrier(new_CC.Barrier)
		updated_CC.Barrier[Id] = true
	} else if old_CC.Value == CC_No && new_CC.Value == CC_Done{
		//Keep old value
		updated_CC.Value = old_CC.Value
		updated_CC.Barrier = Deep_Copy_Barrier(old_CC.Barrier)
	} else if old_CC.Value == new_CC.Value{
		//They are the same, only update Barrier
		updated_CC.Barrier = Bool_Union(old_CC.Barrier, new_CC.Barrier)
	} else if old_CC.Value < new_CC.Value {
		//Accept bigger value
		updated_CC.Value = new_CC.Value
		updated_CC.Barrier = Deep_Copy_Barrier(new_CC.Barrier)
		updated_CC.Barrier[Id] = true
	}

	//update the CC if barriers are fulliled 
	if (updated_CC.Value == CC_Unconfirmed && Check_Barrier(updated_CC.Barrier)){
		updated_CC.Value = CC_Confirmed
		updated_CC.Barrier = make([]bool, N_ELEVATORS)
		updated_CC.Barrier[Id] = true
	}
	if (updated_CC.Value == CC_Done && Check_Barrier(updated_CC.Barrier)){
		updated_CC.Value = CC_No
		updated_CC.Barrier = make([]bool, N_ELEVATORS)
		updated_CC.Barrier[Id] = true
	}

	return updated_CC
}
//-----------------------------------------------------------

//TODO: these do not belong here
func Light_Cab_Lights(Cab_Requests []Request_Cyclic_Counter_t) {

	for floor := 0; floor < elevator.N_FLOORS; floor++{
		elevator.SetButtonLamp(elevator.BT_Cab, floor, CC_To_Bool(Cab_Requests[floor].Value))
	}
}
func Light_Hall_Lights(Hall_Requests [][2]Request_Cyclic_Counter_t) {
	for floor := 0; floor < elevator.N_FLOORS; floor++{
		elevator.SetButtonLamp(elevator.BT_HallUp, floor, CC_To_Bool(Hall_Requests[floor][elevator.BT_HallUp].Value))
		elevator.SetButtonLamp(elevator.BT_HallDown, floor, CC_To_Bool(Hall_Requests[floor][elevator.BT_HallDown].Value))
	}
}
func CC_To_Bool(CC Cyclic_Counter_t) bool {
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

//Helper functions
//-----------------------------------------------------------
func Check_Barrier(Barrier []bool) bool {
	for i := 0; i < N_ELEVATORS; i++{
		if Barrier[i] != Active_Peers[i]{
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

func From_Peers_Update_To_Active_Peers(Peers_Update peers.PeerUpdate) []bool { 
	Active_Peers:= make([]bool, N_ELEVATORS)
	
	for _, peer := range Peers_Update.Peers {
		idx := Peer_Str_To_Int(peer)
		Active_Peers[idx] = true
	}

	return Active_Peers
}

func Peer_Str_To_Int(peer_str string) int {
	num, err := strconv.Atoi(peer_str)
	if err != nil {
		fmt.Println("Invalid number:", err)
		return -1
	}
	return num
}
//-----------------------------------------------------------


//Deep copy funtions for msg_sync_types
//-----------------------------------------------------------
func Deep_Copy_System_Data(src System_Data_t)System_Data_t{
	dst := src 
	dst.Elevator_Data = Deep_Copy_Elevator_data(src.Elevator_Data)
	dst.Hall_Request_Data = Deep_Copy_Hall_Requests(src.Hall_Request_Data)
	return dst
}

func Deep_Copy_Elevator_data(src []Elevator_Data_t) []Elevator_Data_t {
	dst := make([]Elevator_Data_t, len(src))

	for i := range src {
		dst[i] = Deep_Copy_Single_Elevator_Data(src[i])
	}
	return dst
}

func Deep_Copy_Single_Elevator_Data(src Elevator_Data_t) Elevator_Data_t {
	 dst := src

	dst.Is_Alive.Barrier = Deep_Copy_Barrier(src.Is_Alive.Barrier)
	dst.Is_Functional.Barrier = Deep_Copy_Barrier(src.Is_Functional.Barrier)
	dst.Floor.Barrier = Deep_Copy_Barrier(src.Floor.Barrier)
	dst.Elevator_Behaviour.Barrier = Deep_Copy_Barrier(src.Elevator_Behaviour.Barrier)
	dst.Motor_Direction.Barrier = Deep_Copy_Barrier(src.Motor_Direction.Barrier)
	dst.Cab_Requests = Deep_Copy_Cab_Requests(src.Cab_Requests)
	return dst
}

func Deep_Copy_Hall_Requests(src [][2]Request_Cyclic_Counter_t) [][2]Request_Cyclic_Counter_t {
	dst := make([][2]Request_Cyclic_Counter_t, len(src))
	
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

func Deep_Copy_Cab_Requests(src []Request_Cyclic_Counter_t) []Request_Cyclic_Counter_t {
	dst := make([]Request_Cyclic_Counter_t, len(src))
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

func Deep_Copy_Barrier(src []bool) []bool {
	dst := make([]bool, len(src))
	copy(dst, src)
	return dst
}
//-----------------------------------------------------------

