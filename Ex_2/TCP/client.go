package main

import (
	"fmt"
	"io"
	"net"
)

func sendFixed(conn net.Conn, s string){
	buffer := make([]byte, 1024)
	copy(buffer,[]byte(s))
	conn.Write(buffer)
}

func main (){
	server_IP := "10.100.23.11"
	port := 34933

	conn, err := net.Dial("tcp",fmt.Sprintf("%s:%d",server_IP,port))
	if err != nil {
		fmt.Println("connect error:",err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	//For å lese welcome meldingen
	io.ReadFull(conn,buffer)
	fmt.Println("WELCOME:",string(buffer))

	sendFixed(conn, "Hello server")
	sendFixed(conn, "Thnx for us gr 15")

	io.ReadFull(conn, buffer)
	fmt.Println("RECV:",string(buffer))

	io.ReadFull(conn, buffer)
	fmt.Println("RECV:",string(buffer))

}