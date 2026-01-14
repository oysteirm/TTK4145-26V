//Our IP: 10.100.23.25

package main

import (
	"fmt"
	"io"
	"net"
)

func sendFixed(conn net.Conn, s string) {
	buffer := make([]byte, 1024)
	copy(buffer, []byte(s))
	conn.Write(buffer)
}

func main() {

	localAddr := &net.TCPAddr{
		IP:   net.ParseIP("10.100.23.25"), // or nil for any interface
		Port: 20099,
	}

	dialer := net.Dialer{
		LocalAddr: localAddr,
	}

	server_IP := "10.100.23.11"
	port := 34933

	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", server_IP, port))
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	//For å lese welcome meldingen
	io.ReadFull(conn, buffer)
	fmt.Println("WELCOME:", string(buffer))

	sendFixed(conn, "Connect to: 10.100.23.25:20099\x00")
	sendFixed(conn, "Thnx for us gr 15")

	fmt.Println("Local addr:", conn.LocalAddr())
	fmt.Println("Remote addr:", conn.RemoteAddr())

	io.ReadFull(conn, buffer)
	fmt.Println("RECV:", string(buffer))

	io.ReadFull(conn, buffer)
	fmt.Println("RECV:", string(buffer))

}
