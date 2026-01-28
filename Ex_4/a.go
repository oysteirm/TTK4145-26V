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

	//var timer time.Timer
	timeout := 3* time.Second
	for {
		//timer = *time.NewTimer(2 * time.Second)
		recv_socket.SetReadDeadline(time.Now().Add(timeout))
		n, from, err := recv_socket.ReadFromUDP(buffer)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				fmt.Println("TIMEOUT: not heard from B in: ", timeout, " hence B is dead?")
				return
			}
			fmt.Println("read error:", err)
			continue
		}
		fmt.Printf("Received bytes from %s: %s\n", from.String(), string(buffer[:n]))
		//timer runs out
		// if (timer.C == nil) {
		// 	fmt.Println("No message received in 2 seconds")
		// }
		// timer.Stop()
	}
}