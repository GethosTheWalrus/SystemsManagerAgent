package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

var before *cpu.Stats
var after *cpu.Stats

func main() {

	args := os.Args[1:]

	_, portErr := strconv.Atoi(args[1])
	translatedInterval, intervalErr := strconv.Atoi(args[2])

	if len(args) < 3 || portErr != nil || intervalErr != nil {
		fmt.Println("Args[0] must be STRING\nArgs[1] must be INT\nArgs[2] must be INT")
		return
	}

	SERVER := args[0]              //"localhost"
	PORT := args[1]                //"3333"
	INTERVAL := translatedInterval //10

	/*
	 * Connect to the server
	 * Compile the message payload
	 * Send the payload
	 * Wait for reply
	 * Close the server
	 */
	for {

		updateServerWithPayload(SERVER, PORT)

		time.Sleep(time.Duration(INTERVAL) * time.Second)

	}

}

func updateServerWithPayload(SERVER string, PORT string) {

	con, err := net.Dial("tcp", SERVER+":"+PORT)
	checkErr(err)
	defer con.Close()

	operatingSystem := runtime.GOOS
	switch operatingSystem {
	case "windows":
		fmt.Println("Windows")
	case "darwin":
		fmt.Println("MAC operating system")
	case "linux":
		fmt.Println("Linux")
	default:
		fmt.Printf("%s.\n", operatingSystem)
	}

	systemStats := make(map[string]interface{})

	cpuTotal, cpuUser, cpuSystem, cpuIdle := getCPUStats()

	systemStats["cpu"] = map[string]float64{
		"total":  cpuTotal,
		"user":   cpuUser,
		"system": cpuSystem,
		"idle":   cpuIdle,
	}

	memoryTotal, memoryUsed, memoryCached, memoryFree := getMemoryStats(operatingSystem)

	systemStats["memory"] = map[string]uint64{
		"total":  memoryTotal,
		"used":   memoryUsed,
		"cached": memoryCached,
		"free":   memoryFree,
	}

	ip, hostname := getNetworkStats()

	systemStats["network"] = map[string]interface{}{
		"preferred_ip": ip,
		"hostname":     hostname,
	}

	systemStats["general"] = map[string]interface{}{
		"operating_system": operatingSystem,
		"last_seen":        time.Now(),
	}

	payloadJson, err := json.Marshal(systemStats)
	checkErr(err)
	payload := string(payloadJson)

	_, err = con.Write([]byte(payload))
	checkErr(err)

	reply := make([]byte, 1024)
	_, err = con.Read(reply)
	checkErr(err)

	fmt.Println(string(reply))

}

func getNetworkStats() (ip net.IP, hostname string) {
	ip = getOutboundIP()
	hostname, _ = os.Hostname()

	return
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func getMemoryStats(operatingSystem string) (memoryTotal uint64, memoryUsed uint64, memoryCached uint64, memoryFree uint64) {

	memory, err := memory.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	memoryTotal = memory.Total
	memoryUsed = memory.Used
	memoryFree = memory.Free
	//memoryCached = memory.Cached

	return

}

func getCPUStats() (cpuTotal float64, cpuUser float64, cpuSystem float64, cpuIdle float64) {
	var err error

	// get CPU stats
	if before, err = cpu.Get(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	time.Sleep(time.Duration(1) * time.Second)

	if after, err = cpu.Get(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	total := float64(after.Total - before.Total)

	cpuUser = float64(after.User-before.User) / total * 100
	cpuSystem = float64(after.System-before.System) / total * 100
	cpuIdle = float64(after.Idle-before.Idle) / total * 100

	return

}

func checkErr(err error) {

	if err != nil {

		log.Fatal(err)
	}
}
