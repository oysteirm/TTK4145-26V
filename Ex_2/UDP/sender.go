/*
### Sender
C

// if sending directly to a single remote machine:
    addr = new Address(remoteIP, remotePort)
    sock = new Socket(udp)

    // either: set up the socket to use a single remote address
        sock.connect(addr)
        sock.send(message)
    // or: set up the remote address when sending
        sock.sendTo(message, addr)

// if sending on broadcast:
// you have to set up the BROADCAST socket option before calling connect / sendTo
    broadcastIP = #.#.#.255 // First three bytes are from the local IP, or just use 255.255.255.255
    addr = new InternetAddress(broadcastIP, port)
    sendSock = new Socket(udp) // UDP, aka SOCK_DGRAM
    sendSock.setOption(broadcast, true)
    sendSock.sendTo(message, addr)


Received 38 bytes from 10.100.23.18:33973: Hello from UDP server at 10.100.23.18!
Received 38 bytes from 10.100.23.11:47102: Hello from UDP server at 10.100.23.11!

*/

package main

import (
	"fmt"
	"net"
	"time"
)

func recv_msg(recv_socket *net.UDPConn) {
	buffer := make([]byte, 1024)
	for{
		n, from,_ := recv_socket.ReadFromUDP(buffer)
		fmt.Printf("Received bytes from %s: %s\n", from.String(), string(buffer[:n]))
	}

	}

func main() {
	server_IP := "10.100.23.11"
	n:=15


	workspace_port := 20000 + n
	

	recv_addr := net.UDPAddr{IP: net.IPv4zero,Port:workspace_port}
	recv_socket, err := net.ListenUDP("udp", &recv_addr)
	if err != nil {
		panic(err)
	}
	defer recv_socket.Close()

	send_socket, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(err)
	}
	defer send_socket.Close()

	remote_addr := &net.UDPAddr{IP: net.ParseIP(server_IP),Port:workspace_port}
	
	go recv_msg(recv_socket)
	
	for{
		send_socket.WriteToUDP([]byte("hello from Ø H H \n"),remote_addr)
		fmt.Println("SENT hello from Ø H H")
		time.Sleep(1 * time.Second)

	}

}