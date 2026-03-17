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
	"theProject/processPairs"
)

func main() {
	//Start Process pairs
	processPairs.RunProcessPair()

	// Addresses from command inputs
	localID := flag.Int("id", 0, "ID of this elevator (0, 1, 2, ...)")
	ioAddr := flag.String("addr", "", "Elevator IO TCP address (e.g. localhost:15657)")
	flag.Parse()

	// Connecting to the elevator hardware
	if *ioAddr == "" {
		*ioAddr = fmt.Sprintf("localhost:%d", 15657+*localID)
	}
	elevator_IO.Init(*ioAddr, config.N_FLOORS)

	fmt.Println("Starting elevator with localID = ", *localID)
	fmt.Println("Using IO address = ", *ioAddr)

	// Channels and routines for updating peers
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(config.PEER_UPDATE_PORT, strconv.Itoa(*localID), peerTxEnable)
	go peers.Receiver(config.PEER_UPDATE_PORT, peerUpdateCh)


	// elevatorServer <-> messageSync channels
	elevatorDataToMsgSyncFrom_FSM := make(chan messageSync.ElevatorData_t,16)
	requestToMsgSyncFrom_FSM := make(chan []elevator_IO.ButtonEvent_t,16)
	systemDataTo_FSM_FromMsgSync := make(chan messageSync.SystemData_t,16)

	// Launch the go routines
	go messageSync.MessageSyncServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, peerUpdateCh, *localID)
	go elevatorServer.ElevatorServer(elevatorDataToMsgSyncFrom_FSM, requestToMsgSyncFrom_FSM, systemDataTo_FSM_FromMsgSync, *localID)

	for {
	}
}

