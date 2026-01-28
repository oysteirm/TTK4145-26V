package main

import (
	"fmt"
	"net"
)

func main () {

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

	fmt.Println("Listening")

	for (
		
	)
}
