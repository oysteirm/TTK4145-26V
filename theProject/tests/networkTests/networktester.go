package main

//Usage example fetched from https://github.com/TTK4145/Network-go/tree/master
import (
	"theProject/networkDriver/bcast"
	"theProject/networkDriver/localip"
	"theProject/networkDriver/peers"
	"theProject/messageSync"
	"theProject/tests/helperFunctionsForTests"
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
	
	//defined in usage example: peerPort was 15647 and bcastPort 16569
	peerPort := 20021
	bcastPort := 20022

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
	go peers.Transmitter(peerPort, id, peerTxEnable)
	go peers.Receiver(peerPort, peerUpdateCh)

	peerTxEnable <- true

	// We make channels for sending and receiving our custom data types

	//for HelloMSG test
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)

	//for SystemData_t test
	sysTX := make(chan messageSync.SystemData_t)
	sysRX := make(chan messageSync.SystemData_t)

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(bcastPort, sysTX)
	go bcast.Receiver(bcastPort, sysRX)

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
		sd:= testHelpers.MakeFakeConfirmedSystemData(4, 3)
		for {
			sysTX <- sd
			time.Sleep(1 * time.Second)
		}
	}()


	fmt.Println("Started")
	ip, _ := localip.LocalIP()
	println("Local ip: ", ip)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		// case a := <-sysRX:
		// 	//testHelpers.MoreReadablePrint_JSON("Received SystemData",a)
		}
	}
}