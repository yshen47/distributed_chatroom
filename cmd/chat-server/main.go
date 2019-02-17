package main

import (
	"fmt"
	"log"
	"mp1/server"
	"mp1/utils"
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
	utils.SetupLog(name)

	s := new(server.SwimServer)
	s.Constructor(name, peopleNum, portNum, globalServerIPs)

	log.Println("Start server with the: ", name, myAddress, portNum, peopleNum)

	//Join group except for introducer itself
	log.Println("Start joining the group",myAddress)
	s.Join(globalServerIPs)


	//Start the server
	ServerConn, err := net.Listen("tcp", utils.Concatenate(":", portNum))
	utils.CheckError(err)

	go s.DialOthers()
	go s.ListenForDial(ServerConn)


}