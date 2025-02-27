package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const ( 
	HEARTBEAT_PORT = ":9999" // UDP-port for heartbeat-meldinger
	HEARTBEAT_IP   = "127.0.0.1"
	HEARTBEAT_RATE = 1 * time.Second // Hvor ofte master sender heartbeats
	TIMEOUT        = 5 * time.Second // Hvor lenge backup venter før den tar over
	CHECK_INTERVAL = 1 * time.Second // Hvor ofte backup sjekker etter heartbeat
	COUNT_FILE     = "count.txt"     // Fil for å lagre tellingen
)

// Henter siste verdi av telleren fra fil
// Hvis filen ikke finnes, starter vi på 0
func getLastCount() int {
	data, err := os.ReadFile(COUNT_FILE)
	if err != nil {
		return 0 // Hvis filen ikke finnes, returner 0
	}
	count, err := strconv.Atoi(string(data)) // Konverter string -> int
	if err != nil {
		return 0
	}
	return count
}

// Skriver den oppdaterte telleren til fil
func updateCount(count int) {
	os.WriteFile(COUNT_FILE, []byte(strconv.Itoa(count)), 0644)
}

// Sender heartbeat over UDP til backup-prosessen
func sendHeartbeat() {
	conn, err := net.Dial("udp", HEARTBEAT_IP+HEARTBEAT_PORT)
	if err != nil {
		fmt.Println("Feil ved sending av heartbeat:", err)
		return
	}
	defer conn.Close() // Sørger for at vi lukker forbindelsen

	_, err = conn.Write([]byte("alive")) // Send "alive"-melding
	if err != nil {
		fmt.Println("Kunne ikke sende heartbeat:", err)
	}
}

// Starter en ny backup-prosess i et nytt terminalvindu
func spawnBackup() {
	fmt.Println("Starter backup-prosess...")
	cmd := exec.Command("go", "run", os.Args[0]) // Starter en ny instans av seg selv
	cmd.Start()                                  // Start prosessen, men vent ikke på at den fullfører
}

// Backup-prosessen lytter etter UDP-heartbeats fra master
// Hvis den ikke mottar noe heartbeat innen `TIMEOUT`, antar den at master er død
func listenForHeartbeat() bool {
	addr, err := net.ResolveUDPAddr("udp", HEARTBEAT_PORT)
	if err != nil {
		fmt.Println("Feil ved oppsett av UDP:", err)
		return false
	}

	// Åpner en UDP-forbindelse for å lytte etter heartbeats
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Feil ved lytting på UDP:", err)
		return false
	}
	defer conn.Close() // Lukker forbindelsen når vi er ferdige

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(TIMEOUT)) // Timeout hvis vi ikke får heartbeat

	_, _, err = conn.ReadFromUDP(buffer) // Lytt etter heartbeat
	if err != nil {
		return false // Ingen heartbeat mottatt = master er død
	}
	return true // Master lever
}

// Hovedfunksjonen som kjører enten som master eller backup
func main() {
	time.Sleep(1 * time.Second) // Unngå at backup starter for raskt

	// Backup starter her: sjekker om en master allerede lever
	fmt.Println("Backup sjekker om master er aktiv...")
	if listenForHeartbeat() { // Hvis vi får heartbeat, venter vi
		fmt.Println("Master lever. Backup venter...")
		for listenForHeartbeat() { // Lytt kontinuerlig etter heartbeats
			time.Sleep(CHECK_INTERVAL) // Sjekk med jevne mellomrom
		}
		fmt.Println("Master døde! Tar over...")
	}

	// Hvis vi kom hit, betyr det at vi er master
	spawnBackup() // Starter en ny backup-prosess

	count := getLastCount() // Henter siste verdi av telleren
	for {
		count++                       // Øk telleren
		updateCount(count)            // Lagre den nye verdien til filen
		sendHeartbeat()               // Send heartbeat til backup
		fmt.Println("Teller:", count) // Skriv ut verdien
		time.Sleep(HEARTBEAT_RATE)    // Vent før neste økning
	}
}
