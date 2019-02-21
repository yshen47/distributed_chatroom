package main

import (
	"fmt"
	"log"
	"cs425_mp1/server"
	"cs425_mp1/utils"
	"net"
	"os"
	"strconv"
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

	s := new(server.Server)
	s.Constructor(name, peopleNum, portNum, myAddress, globalServerIPs)

	log.Println("Start server with the: ", name, myAddress, peopleNum)


	//Start the server
	ServerConn, err := net.Listen("tcp", myAddress)
	utils.CheckError(err)
	go s.DialOthers()

	for {
		conn, err := ServerConn.Accept()
		if err != nil {
			continue
		}
		go s.HandleConnection(conn)
	}
}