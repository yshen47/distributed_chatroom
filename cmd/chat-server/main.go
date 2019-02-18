package main

import (
	"fmt"
	"log"
	"mp1/server"
	"mp1/utils"
	"net"
	"os"
	"strconv"
	"strings"
)

var DEBUG = true


func main() {
	if len(os.Args) != 4 {
		fmt.Print("Usage: go run main.go [server name] [port] n. \n")
		return
	}
	//Parse input argument
	name := os.Args[1]
	portNum, _ := strconv.Atoi(os.Args[2])
	peopleNum, _ := strconv.Atoi(os.Args[3])
	if DEBUG {
		if !utils.IsPortValid(portNum, peopleNum) {
			fmt.Print("portNum is invalid.\n")
			return
		}
	}

	myAddress := utils.GetCurrentIP(DEBUG, portNum)
	globalServerIPs := utils.GetServerIPs(portNum, peopleNum, DEBUG)
	//utils.SetupLog(name)

	s := new(server.SwimServer)
	s.Constructor(name, peopleNum, portNum, myAddress, globalServerIPs)

	log.Println("Start server with the: ", name, myAddress, peopleNum)


	//Start the server
	ServerConn, err := net.Listen("tcp", utils.Concatenate(":", portNum))
	utils.CheckError(err)
	dialChannel := make(chan server.ConnectionPair)

	go s.DialOthers(dialChannel)

	for {
		fmt.Println("hereerere")
		conn, err := ServerConn.Accept()
		if err != nil {
			continue
		}
		clientIP := conn.LocalAddr().String()
		temp := strings.Split(clientIP, ":")

		clientPortString := temp[len(temp)-1]

		clientPort, _ := strconv.Atoi(clientPortString)
		fmt.Println("clientPort", clientPort)
		fmt.Println("portNum", portNum)
		_, ok := s.EstablishedInConns[clientIP]
		fmt.Println("ok", ok)
		if !ok && clientPort != portNum {
			log.Println("Received new Client TCP connection from ", clientIP)
			s.EstablishedInConns[clientIP] = conn
			go s.HandleRequest(conn)
		}

	}
}