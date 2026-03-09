package messagesynctests

import (
	"theProject/config"
	"theProject/messageSync"
)


func testUpdateHallRequests(){

	localID := 0;
	oldHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	fullBarrier := []bool{1, 1, 1}
	fakePeersList := []bool{1, 1, 1}


	newHallRequests := make([][config.N_UP_DOWN]messageSync.RequestCyclicCounter_t, config.N_FLOORS)

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++{
			newHallRequests[floor][btn].Value = messageSync.CC_Unconfirmed
			newHallRequests[floor][btn].Barrier = messageSync.DeepCopyBarrier(fullBarrier)
		}
	}

	oldHallRequests = messageSync.UpdateHallRequestData(oldHallRequests, newHallRequests, localID)
}