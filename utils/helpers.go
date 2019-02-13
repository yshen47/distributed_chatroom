package utils

import (
	"log"
	"net"
	"os"
	"strings"
)

func SetupLog(name string) {
	f, err := os.OpenFile("../../data/logs/log.txt", os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	//log.SetOutput(f)
	log.SetPrefix(name + " ")
	log.Print("#########################################")
	log.Println("Start server...")
}


func GetCurrentIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}


func Concatenate(str ...string) string {
	var ipAddress []string
	for _, s := range str {
		ipAddress = append(ipAddress, s)
	}

	return strings.Join(ipAddress, "")
}