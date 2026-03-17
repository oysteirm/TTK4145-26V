package processPairs

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"theProject/config"
)

/*
-----------------------------------
Behaviour:
	- Initial primary: spawns backup, starts heartbeat sender, returns
	- Backup: blocks waiting for heartbeats
	- On timeout: backup promotes itself, spawns a new backup, starts heartbeats,
	then returns as the new primary
	- RunProcessPair ensures only the current PRIMARY continues with main startup
-----------------------------------
*/

// Runs instance as backup or primary based on arguments.
func RunProcessPair() {
	if isBackupProcess() {
		runBackup()
		return
	}
	runPrimary()
}

// Checks if this process was spawned as a backup.
func isBackupProcess() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--backup" {
			return true
		}
	}
	return false
}

// Spawns a backup, starts heartbeats in a goroutine, then returns.
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

// Launches a new instance of this binary with the --backup flag.
func spawnBackup() {
	// prefer the executable path returned by os.Executable, fallback to argv[0]
	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}
	args := append(os.Args[1:], "--backup")

	// On Linux, try to launch the backup in a visible terminal window first.
	if runtime.GOOS == "linux" {
		if trySpawnBackupInLinuxTerminal(self, args) {
			fmt.Println("[ProcessPair] Backup spawned in new terminal")
			return
		}
	}

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

// Attempts common terminal emulators.
// Returns true if launching succeeded.
func trySpawnBackupInLinuxTerminal(self string, args []string) bool {
	quotedSelf := shellQuote(self)
	quotedArgs := make([]string, 0, len(args))
	for _, a := range args {
		quotedArgs = append(quotedArgs, shellQuote(a))
	}
	runCmd := strings.TrimSpace(quotedSelf + " " + strings.Join(quotedArgs, " "))

	type terminalSpec struct {
		bin    string
		params []string
	}

	specs := []terminalSpec{
		{bin: "x-terminal-emulator", params: []string{"-e", runCmd}},
		{bin: "gnome-terminal", params: []string{"--", "bash", "-lc", runCmd}},
		{bin: "konsole", params: []string{"-e", "bash", "-lc", runCmd}},
		{bin: "xfce4-terminal", params: []string{"-e", runCmd}},
		{bin: "xterm", params: []string{"-e", runCmd}},
	}

	for _, spec := range specs {
		if _, err := exec.LookPath(spec.bin); err != nil {
			continue
		}
		if err := exec.Command(spec.bin, spec.params...).Start(); err == nil {
			return true
		}
	}

	return false
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// Re-launches the primary role in this same process,
// without restarting the binary, just switches role internally.
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
