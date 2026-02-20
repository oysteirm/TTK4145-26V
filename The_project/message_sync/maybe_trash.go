package message_sync

func On_Recieved_Fresh_Data(system_data System_Data_t, confirmed_system_data System_Data_t, fresh_data System_Data_t) (System_Data_t, System_Data_t){
	var updated_system_data System_Data_t = system_data
	var updated_confirmed_system_data System_Data_t = confirmed_system_data

	for i := 0; i < N_ELEVATORS; i++{
		if fresh_data.Id == i {
			updated_system_data.Elevator_Data[i] = Update_Single_Elevator_Data(system_data.Elevator_Data[i], fresh_data.Elevator_Data[i], system_data.Id)
		} else{
			updated_system_data.Elevator_Data[i] = Update_Single_Elevator_Barriers(system_data.Elevator_Data[i], fresh_data.Elevator_Data[i], system_data.Id)
		}
	}
	system_data.Hall_Request_Data = Update_Hall_Request_Data(system_data.Hall_Request_Data, fresh_data.Hall_Request_Data)

	updated_confirmed_system_data = Update_Confirmed_System_Data(system_data, confirmed_system_data)

	return updated_system_data, updated_confirmed_system_data
}

func Update_Single_Elevator_Data(old_data Elevator_Data_t, new_data Elevator_Data_t, id int) Elevator_Data_t{ //maybe problem, no safe update of not requests. we take what we are given. 
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

func Update_Single_Elevator_Barriers(old_data Elevator_Data_t, new_data Elevator_Data_t, id int) Elevator_Data_t{
	var updated_data Elevator_Data_t = old_data

	if old_data.Is_Alive.value == new_data.Is_Alive.value {
		 updated_data.Is_Alive.barrier = Bool_Union(old_data.Is_Alive.barrier, new_data.Is_Alive.barrier)
	} 
	if old_data.Is_Able.value == new_data.Is_Able.value {
		 updated_data.Is_Able.barrier = Bool_Union(old_data.Is_Able.barrier, new_data.Is_Able.barrier)
	} 
	if old_data.Floor.value == new_data.Floor.value {
		 updated_data.Floor.barrier = Bool_Union(old_data.Floor.barrier, new_data.Floor.barrier)
	} 
	if old_data.Elevator_Behaviour.value == new_data.Elevator_Behaviour.value {
		 updated_data.Elevator_Behaviour.barrier = Bool_Union(old_data.Elevator_Behaviour.barrier, new_data.Elevator_Behaviour.barrier)
	} 
	if old_data.Motor_Direction.value == new_data.Motor_Direction.Motor_Direction {
		 updated_data.Motor_Direction.barrier = Bool_Union(old_data.Motor_Direction.barrier, new_data.Motor_Direction.barrier)
	} 

	for i := 0; i < N_ELEVATORS; i++{
		updated_data.Cab_Requests[i] = Update_CC(old_data.Cab_Requests[i], new_data.Cab_Requests[i], id)
	}

	return updated_data
}

func Update_Confirmed_System_Data(unconfirmed_data System_Data_t, confirmed_data System_Data_t) System_Data_t{
	for i := 0; i < N_ELEVATORS; i++ {
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Alive.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Alive.value = unconfirmed_data.Elevator_Data[i].Is_Alive.value
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Is_Able.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Is_Able.value = unconfirmed_data.Elevator_Data[i].Is_Able.value
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Floor.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Floor.value = unconfirmed_data.Elevator_Data[i].Floor.value
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Elevator_Behaviour.value = unconfirmed_data.Elevator_Data[i].Elevator_Behaviour.value
		}
		if Check_Barrier(unconfirmed_data.Elevator_Data[i].Motor_Direction.barrier, elevator_network_list) {
			confirmed_data.Elevator_Data[i].Motor_Direction.value = unconfirmed_data.Elevator_Data[i].Motor_Direction.value
		}
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if Check_Barrier(unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].barrier, elevator_network_list) {
				confirmed_data.Elevator_Data[i].Cab_Requests[floor].value = unconfirmed_data.Elevator_Data[i].Cab_Requests[floor].value
			}
		}
	}
}