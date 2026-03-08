package theproject

import (
	"theProject/elevatorServer"
	"theProject/messageSync"
	"theProject/networkDriver/peers"
	"theProject/elevator_IO"
)

func main(){
	//TODO: make process pairs with primary-backup topology

	//TODO: get the local ID from terminal command
	localID := 0//hjelp henning

	//TODO: initialize all variables and constants
	//TODO: initialize all the channels
	//elevatorServer channels
	elevatorDataToMsgSyncFrom_FSM := make(chan messageSync.ElevatorData_t)
    requestToMsgSyncFrom_FSM := make(chan []elevator_IO.ButtonEvent_t)
	systemDataTo_FSM_FromMsgSync := make(chan messageSync.SystemData_t)	
	peersReciever := make(chan peers.PeerUpdate)

	//TODO: launch the go routines 
	go messageSync.MessageSyncServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, peersReciever, localID)
	go elevatorServer.ElevatorServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, localID)

	//TODO: forever loop?
	for{}
}