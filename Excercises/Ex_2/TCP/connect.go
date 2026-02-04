//Our IP: 10.100.23.25

package main

import (
	"fmt"
	"io"
	"net"
)

func send_fixed(conn net.Conn, s string) {
	buffer := make([]byte, 1024)
	copy(buffer, []byte(s))
	conn.Write(buffer)
}

func main() {

	ln, _ := net.Listen("tcp", ":20015")

	server_IP := "10.100.23.11"
	port := 34933

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server_IP, port))
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	//For å lese welcome meldingen
	io.ReadFull(conn, buffer)
	fmt.Println("WELCOME:", string(buffer))

	send_fixed(conn, "Connect to: 10.100.23.25:20015\x00")
	server_conn, _ := ln.Accept()

	//server_conn.Write([]byte("hello back!\x00"))
	send_fixed(server_conn, "hello back!")
	send_fixed(conn, "Thnx for us gr 15")

	io.ReadFull(conn, buffer)
	fmt.Println("RECV:", string(buffer))

	io.ReadFull(server_conn, buffer)
	fmt.Println("RECV:", string(buffer))

	io.ReadFull(server_conn, buffer)
	fmt.Println("RECV:", string(buffer))

	io.ReadFull(conn, buffer)
	fmt.Println("RECV:", string(buffer))

}
