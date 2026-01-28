package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	addr := net.UDPAddr {
	IP: net.IPv4zero,
	Port: 30000,
	}

	recv_socket, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer recv_socket.Close()

	fmt.Println("Listening")

	buffer := make([]byte, 1024)

	i := 1

	for {
		n, from, err := recv_socket.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("read error:", err)
			continue
		}
		fmt.Printf("Received bytes from %s: %s\n", from.String(), string(buffer[:n]))
		
		

	}
}