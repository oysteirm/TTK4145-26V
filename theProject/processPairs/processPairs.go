package processPairs

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
	"theProject/config"
)

func RunProcessPairs(localID int) {
	var state int = 0
	
	for {
		switch state {
		case 0:
			addr := net.UDPAddr {
				IP: net.IPv4zero,
				Port: config.PP_PORT,
			}

			recvSocket, err := net.ListenUDP("udp", &addr)
			if err != nil {
				panic(err)
			}
			buffer := make([]byte, 1024)

			timeout := config.PP_TIMEOUT

			for {
				recvSocket.SetReadDeadline(time.Now().Add(timeout))
				n, _, err := recvSocket.ReadFromUDP(buffer)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						state = 1
						recvSocket.Close()
						break
					}
					fmt.Println("Read error: ", err)
					continue
				}
			}

		case 2:
			exec.Command("gnome.terminal", "--", "go", "run", "theProject/main.go").Start()

			time.Sleep(1 * time.Second) // Is this necessary??

			server_IP := config.PP_SERVER_IP

			sendSocket, err := net.ListenUDP("udp", nil)
			if err != nil {
				panic(err)
			}
			defer sendSocket.Close()

			
			remoteAddr := &net.UDPAddr{IP: net.ParseIP(server_IP), Port: config.PP_PORT}


			// TODO: Find out what to be in this FOR-loop
			for {
				
			}
		}
	}
}
