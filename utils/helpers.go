package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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

func Concatenate(elem ...interface{}) string {
	var ipAddress []string
	for _, e := range elem {
		switch v := e.(type) {
		case string:
			ipAddress = append(ipAddress, v)
		case int:
			t := strconv.Itoa(v)
			ipAddress = append(ipAddress, t)
		default:
			fmt.Printf("unexpected type %T", v)
		}
	}

	return strings.Join(ipAddress, "")
}

func GetServerIPs(port int, num int, debug bool) [] string {
	ips := make([]string, num)
	if debug {
		for i := range ips {
			ips[i] = Concatenate("localhost:", 5800 + i * 100)
		}
	} else {
		for i := range ips {
			if i == 9 {
				ips[9] = Concatenate("sp19-cs425-g18-10.cs.illinois.edu:", port)
			} else {
				ips[i] = Concatenate("sp19-cs425-g18-0", i+1, ".cs.illinois.edu:", port)
			}

		}
	}
	return ips
}

func GetCurrentIP(debug bool, port int) string {
	if !debug {
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
	} else {
		return Concatenate("localhost:", port)
	}
}

//Only for debug mode
func IsPortValid(port int, num int) bool {
	for i := 0; i < num; i++ {
		if port == 5800 + i * 100 {
			return true
		}
	}
	return false
}

func CheckError(err error) {
	if err != nil {
		log.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
