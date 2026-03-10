package main

import (
	"flag"
	"fmt"
	"strconv"
	"theProject/config"
	"theProject/elevatorServer"
	"theProject/elevator_IO"
	"theProject/messageSync"
	"theProject/networkDriver/peers"
)

func main(){
	//TODO: make process pairs with primary-backup topology

	//TODO: get the local ID from terminal command
	localID := flag.Int("id", 0, "ID of this elevator (0, 1, 2, ...)") 
	flag.Parse()

	fmt.Println("Starting elevator with localID = ", *localID)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(config.PEER_UPDATE_PORT, strconv.Itoa(*localID), peerTxEnable)
	go peers.Receiver(config.PEER_UPDATE_PORT, peerUpdateCh)

	//TODO: initialize all variables and constants
	//TODO: initialize all the channels
	//elevatorServer channels
	elevatorDataToMsgSyncFrom_FSM := make(chan messageSync.ElevatorData_t)
    requestToMsgSyncFrom_FSM := make(chan []elevator_IO.ButtonEvent_t)
	systemDataTo_FSM_FromMsgSync := make(chan messageSync.SystemData_t)	

	//TODO: launch the go routines 
	go messageSync.MessageSyncServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, peerUpdateCh, *localID)
	go elevatorServer.ElevatorServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, *localID)

	//TODO: forever loop?
	for{}
}

//NOTES: barrier to wait for the initalization to be done before starting program
//check the book on semafores for example
