package main

import (
	"fmt"
	"net"
	"time"
	"os/exec"
	"strconv"
)


func main() {

	var state int = 0
	var counter int = 0

	for {
		switch state {
		case 0: //be A
			addr := net.UDPAddr {
			IP: net.IPv4zero,
			Port: 30000,
			}

			recv_socket, err := net.ListenUDP("udp", &addr)
			if err != nil {
				panic(err)
			}
			buffer := make([]byte, 1024)

			timeout := 3* time.Second
			for {
				recv_socket.SetReadDeadline(time.Now().Add(timeout))
				n, _, err := recv_socket.ReadFromUDP(buffer)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						//fmt.Println("TIMEOUT: not heard from B in: ", timeout, " hence B is dead?")
						state = 1
						recv_socket.Close()
						counter++
						break
					}
					fmt.Println("read error:", err)
					continue
				}
				counter, _ = strconv.Atoi(string(buffer[:n]))
				//fmt.Printf("%d\n",  counter)
			}

		case 1: // B
			exec.Command("gnome-terminal", "--", "go", "run", "ab.go").Start()

			time.Sleep(1 * time.Second)

			server_IP := "127.0.0.1"

			workspace_port := 30000

			send_socket, err := net.ListenUDP("udp", nil)
			if err != nil {
				panic(err)
			}
			defer send_socket.Close()

			remote_addr := &net.UDPAddr{IP: net.ParseIP(server_IP),Port:workspace_port}

			for {
				msg := strconv.Itoa(counter)
				send_socket.WriteToUDP([]byte(msg), remote_addr)
				fmt.Println(counter)
				counter++
				time.Sleep(1 * time.Second)
			}
		}
	}
}
