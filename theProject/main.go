package main

import (
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
	localID := 0//hjelp meg, henning

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(config.PEER_UPDATE_PORT, strconv.Itoa(localID), peerTxEnable)
	go peers.Receiver(config.PEER_UPDATE_PORT, peerUpdateCh)

	//TODO: initialize all variables and constants
	//TODO: initialize all the channels
	//elevatorServer channels
	elevatorDataToMsgSyncFrom_FSM := make(chan messageSync.ElevatorData_t)
    requestToMsgSyncFrom_FSM := make(chan []elevator_IO.ButtonEvent_t)
	systemDataTo_FSM_FromMsgSync := make(chan messageSync.SystemData_t)	

	//TODO: launch the go routines 
	go messageSync.MessageSyncServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, peerUpdateCh, localID)
	go elevatorServer.ElevatorServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, localID)

	//TODO: forever loop?
	for{}
}

//NOTES: barrier to wait for the initalization to be done before starting program
//check the book on semafores for example

student@NTNU24604:~/Documents/gruppe_14/TTK4145-26V/theProject/tests/networkTests$ go run networktester.go -id=1
Started
Broadcast socket bound to: 0.0.0.0:20022
Broadcast socket bound to: 0.0.0.0:20022
Bcast sent 1839 bytes to 255.255.255.255:20022
Local ip:  10.100.23.172
Bcast sent 1839 bytes to 255.255.255.255:20022
Bcast sent 1839 bytes to 255.255.255.255:20022
Bcast sent 1839 bytes to 255.255.255.255:20022
