package main

//Usage example fetched from https://github.com/TTK4145/Network-go/tree/master
import (
	"TTK4145-26V/Network_Driver/bcast"
	"TTK4145-26V/Network_Driver/localip"
	"TTK4145-26V/Network_Driver/peers"
	"TTK4145-26V/message_sync"
	"TTK4145-26V/Tests/Helper_functions_for_tests"
	"flag"
	"fmt"
	"os"
	"time"
)




// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	
	//defined in usage example: peer_port was 15647 and bcast_port 16569
	peer_port := 20011
	bcast_port := 20012

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(peer_port, id, peerTxEnable)
	go peers.Receiver(peer_port, peerUpdateCh)

	// We make channels for sending and receiving our custom data types

	//for HelloMSG test
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)

	//for System_Data_t test
	sys_TX := make(chan message_sync.System_Data_t)
	sys_RX := make(chan message_sync.System_Data_t)

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(bcast_port, sys_TX)
	go bcast.Receiver(bcast_port, sys_RX)

	// The example message. We just send one of these every second.
	// go func() {
	// 	helloMsg := HelloMsg{"Hello from " + id, 0}
	// 	for {
	// 		helloMsg.Iter++
	// 		helloTx <- helloMsg
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()
	go func() {
		sd:= test_helpers.Make_Fake_Confirmed_System_Data_t(4, 3)
		for {
			sys_TX <- sd
			time.Sleep(1 * time.Second)
		}
	}()


	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-sys_RX:
			fmt.Printf("Received:")
			test_helpers.More_readable_json_print("Received System_Data",a)
		}
	}
}