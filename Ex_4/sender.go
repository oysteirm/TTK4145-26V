package main

import (
	"fmt"
	"net"
	"time"
)

func recv_msg(recv_socket *net.UDPConn) {
	buffer :=make([]byte, 1024)
	for {
		n, from, _ := recv_socket.ReadFromUDP(buffer)
		fmt.Printf("Received bytes from %s: %s\n", from.String(), string(buffer[:n]))
	}
}

func main() {
	server_IP := "10.100.23.11"
	n:= 15

	workspace_port := 20000 + n

	recv_addr := net.UDPAddr{IP: net. IPv4zero, Port: workspace_port}
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

	for {




		
	}
}