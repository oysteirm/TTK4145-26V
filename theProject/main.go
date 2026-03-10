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

student@NTNU24604:~/Documents/gruppe_14/TTK4145-26V/theProject/tests/networkTests$ ip addr
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host noprefixroute 
       valid_lft forever preferred_lft forever
2: enp1s0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether 00:13:3b:11:04:8d brd ff:ff:ff:ff:ff:ff
3: enp0s31f6: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether d0:8e:79:07:e0:b1 brd ff:ff:ff:ff:ff:ff
    inet 10.100.23.172/24 brd 10.100.23.255 scope global dynamic noprefixroute enp0s31f6
       valid_lft 8818sec preferred_lft 8818sec
    inet6 fe80::d7c9:2c07:c2a5:cd3b/64 scope link noprefixroute 
       valid_lft forever preferred_lft forever
4: wlp0s20f3: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000
    link/ether 2c:8d:b1:40:15:7d brd ff:ff:ff:ff:ff:ff
