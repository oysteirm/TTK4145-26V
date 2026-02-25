package message_sync

import (
	"../elevator"
	"fmt"
	"time"
)

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


func Update_Hall_Request_Data(old_data [][2]Request_Cyclic_Counter_t, new_data [][2]Request_Cyclic_Counter_t, id int) [][2]Request_Cyclic_Counter_t {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			Update_CC(old_data[floor][btn], new_data[floor][btn], id)
		}
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