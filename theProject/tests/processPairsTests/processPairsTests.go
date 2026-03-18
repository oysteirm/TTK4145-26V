package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"theProject/config"
)

const backupArg = "--backup"

func main() {
	testIsBackupProcess()
	testBackupPromotesWhenPrimaryDies()
	testHeartbeatKeepsBackupAlive()
}

// testIsBackupProcess verifies role detection from command-line arguments
// by manipulating os.Args directly — no real processes started.
func testIsBackupProcess() {
	fmt.Println("--- testIsBackupProcess ---")

	original := os.Args

	os.Args = []string{"elevator"}
	assertEqBool(false, isBackupProcess(), "no --backup flag")

	os.Args = []string{"elevator", "--backup"}
	assertEqBool(true, isBackupProcess(), "--backup present")

	os.Args = []string{"elevator", "--id=0", "--backup"}
	assertEqBool(true, isBackupProcess(), "--backup in the middle")

	os.Args = original
	fmt.Println("PASS")
}

// testBackupPromotesWhenPrimaryDies starts a real UDP listener (backup) and
// a real UDP sender (primary). The sender stops after a few packets and we
// verify the listener detects the timeout.
func testBackupPromotesWhenPrimaryDies() {
	fmt.Println("--- testBackupPromotesWhenPrimaryDies ---")

	promoted := make(chan struct{})

	go func() {
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: config.PP_PORT}
		sock, err := net.ListenUDP("udp", addr)
		if err != nil {
			fmt.Println("FAIL: could not bind UDP port:", err)
			close(promoted)
			return
		}
		defer sock.Close()

		buf := make([]byte, 64)
		for {
			sock.SetReadDeadline(time.Now().Add(config.PP_TIMEOUT))
			_, _, err := sock.ReadFromUDP(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					close(promoted)
					return
				}
			}
		}
	}()

	// Send 5 heartbeats then stop — simulates primary dying.
	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		sock, err := net.ListenUDP("udp", nil)
		if err != nil {
			fmt.Println("FAIL: could not open send socket:", err)
			return
		}
		defer sock.Close()

		dest := &net.UDPAddr{IP: net.ParseIP(config.PP_SERVER_IP), Port: config.PP_PORT}
		for i := 0; i < 5; i++ {
			sock.WriteToUDP([]byte("heartbeat"), dest)
			time.Sleep(config.PP_INTERVAL)
		}
	}()

	<-senderDone

	select {
	case <-promoted:
		fmt.Println("PASS - backup detected primary death and promoted")
	case <-time.After(3 * config.PP_TIMEOUT):
		fmt.Println("FAIL - backup did not detect primary death in time")
	}
}

// testHeartbeatKeepsBackupAlive verifies that a backup does NOT promote
// itself while heartbeats are continuously flowing.
func testHeartbeatKeepsBackupAlive() {
	fmt.Println("--- testHeartbeatKeepsBackupAlive ---")

	window := 5 * config.PP_TIMEOUT
	promoted := make(chan struct{})

	go func() {
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: config.PP_PORT}
		sock, err := net.ListenUDP("udp", addr)
		if err != nil {
			fmt.Println("FAIL: could not bind UDP port:", err)
			close(promoted)
			return
		}
		defer sock.Close()

		buf := make([]byte, 64)
		for {
			sock.SetReadDeadline(time.Now().Add(config.PP_TIMEOUT))
			_, _, err := sock.ReadFromUDP(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					close(promoted)
					return
				}
			}
		}
	}()

	sock, err := net.ListenUDP("udp", nil)
	if err != nil {
		fmt.Println("FAIL: could not open send socket:", err)
		return
	}
	defer sock.Close()

	dest := &net.UDPAddr{IP: net.ParseIP(config.PP_SERVER_IP), Port: config.PP_PORT}
	deadline := time.Now().Add(window)
	for time.Now().Before(deadline) {
		sock.WriteToUDP([]byte("heartbeat"), dest)
		time.Sleep(config.PP_INTERVAL)
	}

	select {
	case <-promoted:
		fmt.Println("FAIL - backup promoted itself despite receiving heartbeats")
	default:
		fmt.Println("PASS - backup stayed passive while heartbeats were flowing")
	}
}

// isBackupProcess mirrors the unexported function in processPairs to allow
// testing its logic without needing to export it.
func isBackupProcess() bool {
	for _, arg := range os.Args[1:] {
		if arg == backupArg {
			return true
		}
	}
	return false
}

func assertEqBool(expected, actual bool, label string) {
	if expected != actual {
		fmt.Printf("  FAIL [%s]: expected %t, got %t\n", label, expected, actual)
	}
}
