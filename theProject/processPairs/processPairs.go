package processPairs

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"theProject/config"
)

const (
	// heartbeat interval taken from configuration; this is the delay between
	// the primary sending packets to the backup.
	HEARTBEAT = config.PP_INTERVAL
)

// RunProcessPair listens for heartbeats from the backup.
// If the backup stops sending, this process becomes the backup,
// spawns a new primary, and starts sending heartbeats to it.
func RunProcessPair() {
	role := roleFromArgs() // "primary" or "backup"

	switch role {
	case "backup":
		runBackup()
	default:
		runPrimary()
	}
}

// roleFromArgs checks if this process was spawned as a backup.
func roleFromArgs() string {
	for _, arg := range os.Args[1:] {
		if arg == "--backup" {
			return "backup"
		}
	}
	return "primary"
}

// runPrimary is the ACTIVE role.
// It sends a heartbeat to the backup on PP_PORT every HEARTBEAT interval.
// It never returns.
func runPrimary() {
	fmt.Println("[ProcessPair] Running as PRIMARY")

	// Spawn the backup process headlessly, then start sending packets.
	spawnBackup()
	time.Sleep(200 * time.Millisecond) // give the backup time to start listening

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
		time.Sleep(HEARTBEAT)
	}
}

// runBackup is the PASSIVE role.
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
		// Heartbeat received — primary is alive, stay passive
	}
}

// spawnBackup launches a new instance of this binary with the --backup flag.
// It does not rely on a graphical terminal so that the library works on
// macOS, Linux and Windows in headless test environments.
func spawnBackup() {
	// prefer the executable path returned by os.Executable, fallback to argv[0]
	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}
	args := append(os.Args[1:], "--backup")
	cmd := exec.Command(self, args...)

	// redirect I/O so messages from the child appear in the parent's console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("[ProcessPair] Failed to spawn backup:", err)
	} else {
		fmt.Println("[ProcessPair] Backup spawned")
	}
}

// promoteToPrimary re-launches the primary role in this same process,
// without restarting the binary — just switches role internally.
func promoteToPrimary() {
	// remove --backup flag from args so future children are
	// always started in backup mode only
	for i, a := range os.Args {
		if a == "--backup" {
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}

	fmt.Println("[ProcessPair] Taking over as PRIMARY and spawning new backup")
	runPrimary()
}
