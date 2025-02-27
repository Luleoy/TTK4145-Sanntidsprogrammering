package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const BackupAddress = "localhost:8080"
const ListenAddr = "localhost:8080"
const Timeout = 2 * time.Second

func main() {
	// ------ BACKUP MODE ------ //
	// Connect to socket
	addr, err := net.ResolveUDPAddr("udp", ListenAddr)
	if err != nil {
		fmt.Println("Error resolving address", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening for UDP:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Backup listening for heartbeat...")
	buf := make([]byte, 1024)
	count := 0

	// Listen to heartbeats
	for {
		conn.SetReadDeadline(time.Now().Add(Timeout))

		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				fmt.Println("Heartbeat timed out! Taking over...")
				break // Exit loop and start counting
			}
			fmt.Println("Error receiving heartbeat:", err)
			continue
		}

		count, _ = strconv.Atoi(string(buf[:n])) // Extract count
	}

	// ------ TAKE OVER ------ //

	conn.Close()

	// Start backup in new terminal

	BackupFilePath := "~/Documents/nyheismonster/TTK4145-Sanntidsprogrammering/exercise4/main.go"

	cmd := exec.Command("gnome-terminal", "--", "bash", "-c", "go run "+BackupFilePath+"; exec bash")
	error := cmd.Start()
	if error != nil {
		fmt.Println("Failed to start backup process:", err)
		return
	}

	backupAddr, err := net.ResolveUDPAddr("udp", BackupAddress)
	if err != nil {
		fmt.Println("Failed to resolve backup address:", err)
		os.Exit(1)
	}

	conn, err = net.DialUDP("udp", nil, backupAddr)
	if err != nil {
		fmt.Println("Failed to connect to backup process:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// ------- PRIMARY MODE ------- //

	count++
	for {
		// Send a heartbeat with count
		msg := fmt.Sprintf("%d", count)
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Failed to send heartbeat:", err)
		}
		// Print the count
		fmt.Println(count)
		time.Sleep(time.Second)
		count++
	}
}
