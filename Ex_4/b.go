package main

import (
	"os/exec"
	"fmt"
	"net"
	"time"
)

func main() {
	exec.Command("gnome-terminal", "--", "go", "run", "a.go").Run()

	time.Sleep(1 * time.Second)

	server_IP := "127.0.0.1"

	workspace_port := 30000

	send_socket, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(err)
	}
	defer send_socket.Close()

	remote_addr := &net.UDPAddr{IP: net.ParseIP(server_IP),Port:workspace_port}

	//go recv_msg(recv_socket)

	
	for {

		send_socket.WriteToUDP([]byte("hello from Ø H H \n"),remote_addr)
		fmt.Println("SENT hello from Ø H H")
		time.Sleep(1 * time.Second)
	}
}