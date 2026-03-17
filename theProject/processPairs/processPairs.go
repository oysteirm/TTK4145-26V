package processPairs

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
	"theProject/config"
)

const backupArg = "--backup"

/*
-----------------------------------
Functionality:
	- Initial primary: spawns backup, starts heartbeat sender, returns
	- Backup: blocks waiting for heartbeats
	- On timeout: backup promotes itself, spawns a new backup, starts heartbeats,
	then returns as the new primary
	- RunProcessPair ensures only the current PRIMARY continues with main startup
-----------------------------------
*/

func RunProcessPair() {
	if isBackupProcess() {
		runBackup()
		return
	}
	runPrimary()
}

func isBackupProcess() bool {
	for _, arg := range os.Args[1:] {
		if arg == backupArg {
			return true
		}
	}
	return false
}

func runPrimary() {
	fmt.Println("[ProcessPair] Running as PRIMARY")

	spawnBackup()
	time.Sleep(200 * time.Millisecond)

	go heartbeatLoop()
}

func heartbeatLoop() {

	sendSocket, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(fmt.Sprintf("[ProcessPair] Failed to open send socket: %v", err))
	}
	defer sendSocket.Close()

	backupAddr := &net.UDPAddr{
		IP:   net.ParseIP(config.PP_SERVER_IP),
		Port: config.PP_PORT,
	}

	for {
		_, err := sendSocket.WriteToUDP([]byte("heartbeat"), backupAddr)
		if err != nil {
			fmt.Println("[ProcessPair] Heartbeat send error:", err)
		}
		time.Sleep(config.PP_INTERVAL)
	}
}

// It listens for heartbeats from the primary on PP_PORT.
// If the primary times out, it promotes itself to primary.
func runBackup() {
	fmt.Println("[ProcessPair] Running as BACKUP — listening for primary heartbeat")

	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: config.PP_PORT,
	}

	recvSocket, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(fmt.Sprintf("[ProcessPair] Failed to open recv socket: %v", err))
	}

	buf := make([]byte, 64)

	for {
		recvSocket.SetReadDeadline(time.Now().Add(config.PP_TIMEOUT))
		_, _, err := recvSocket.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				fmt.Println("[ProcessPair] Primary heartbeat lost — promoting to PRIMARY")
				recvSocket.Close()
				promoteToPrimary()
				return
			}
			fmt.Println("[ProcessPair] Read error:", err)
		}
		// Heartbeat received, primary is alive, stay passive
	}
}

// Spawn backup process.
func spawnBackup() {
	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}
	args := append(os.Args[1:], backupArg)
	terminalArgs := append([]string{"--", self}, args...)
	exec.Command("gnome-terminal", terminalArgs...).Start()
}

// Remove --backup and continue as primary in this process.
func promoteToPrimary() {
	for i, a := range os.Args {
		if a == backupArg {
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}

	fmt.Println("[ProcessPair] Taking over as PRIMARY and spawning new backup")
	runPrimary()
}