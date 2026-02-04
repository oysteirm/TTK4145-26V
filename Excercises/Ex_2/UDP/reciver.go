//UDP uses datagrams, so receiveFrom will return whenever it receives anything. The buffer size is just the maximum size of the message, it doesn't have to be "filled". This example is for broadcasting.

//### Receiver
/*C
// the address we are listening for messages on
// we have no choice in IP, so use 0.0.0.0, INADDR_ANY, or leave the IP field empty
// the port should be whatever the sender sends to
// alternate names: sockaddr, resolve(udp)addr, 
InternetAddress addr

// a socket that plugs our program to the network. This is the "portal" to the outside world
// alternate names: conn
// UDP is sometimes called SOCK_DGRAM. You will sometimes also find UDPSocket or UDPConn as separate types
recvSock = new Socket(udp)

// bind the address we want to use to the socket,
recvSock.bind(addr)


// a buffer where the received network data is stored
byte[1024] buffer  

// an empty address that will be filled with info about who sent the data
InternetAddress fromWho 

loop {
    // clear buffer (or just create a new one)
    
    // receive data on the socket
    // fromWho will be modified by ref here. Or it's a return value. Depends.
    // receive-like functions return the number of bytes received
    // alternate names: read, readFrom
    numBytesReceived = recvSock.receiveFrom(buffer, ref fromWho)
    
    // the buffer just contains a bunch of bytes, so you may have to explicitly convert it to a string
    
    // optional: filter out messages from ourselves
    if(fromWho.IP != localIP){
        // do stuff with buffer
    }

Received 38 bytes from 10.100.23.18:33973: Hello from UDP server at 10.100.23.18!
Received 38 bytes from 10.100.23.11:47102: Hello from UDP server at 10.100.23.11!


//UDP_port 30000
//IP_port 20000+ 
*/
package main

import (
	"fmt"
	"net"
)

func main (){

	addr := net.UDPAddr {
	IP: net.IPv4zero, 
	Port: 30000,
	}

	recv_socket, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer recv_socket.Close()

	buffer := make([]byte, 1024)

	fmt.Println("Listening for UDP on port 30000")

	for {
		n, from_who, err := recv_socket.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		message := string(buffer[:n])

		fmt.Printf("Received %d bytes from %s: %s\n", n, from_who.String(), message)
	}
}

