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

func main() {
	if len(os.Args) != 4 {
		fmt.Print("Usage: go run main.go [server name] [port] n. \n")
		return
	}
	//Parse input argument
	name := os.Args[1]
	portNum, _ := strconv.Atoi(os.Args[2])
	peopleNum, _ := strconv.Atoi(os.Args[3])
	myAddress := utils.GetCurrentIP()
	utils.SetupLog(name)
	log.Println("Start server with the: ", name, myAddress, portNum, peopleNum)


	s := new(server.SwimServer)
	s.Constructor(name, peopleNum, portNum)

	//Join group except for introducer itself
	log.Println("Start joining the group",myAddress)
	s.Join()


	//Start the server
	ServerConn, err := net.ListenUDP("udp", &net.UDPAddr{IP:[]byte{0,0,0,0},Port:portNum,Zone:""})
	//ServerConn, err := net.Dial("udp", &net.UDPAddr{IP:[]byte{0,0,0,0},Port:myPort,Zone:""})
	//ServerConn, err := net.Dial("udp",myAddress)
	if err != nil {
		log.Println(err)
		return
	}

	defer ServerConn.Close()
/*
	go s.StartPing(1 * time.Second)

	//wait for incoming response
	buf := make([]byte, 1024)

	for {
		n, _ := ServerConn.Read(buf)
		var resultMap server.Action
		// parse resultMap to json format
		err := json.Unmarshal(buf[0:n], &resultMap)
		if err != nil {
			fmt.Println("error:", err)
		}

		//Customize different action
		if resultMap.ActionType == 0 {
			//received join
			log.Println("Received Join from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.GlobalState.List)
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.GlobalState.List)
			s.Ack(resultMap.IpAddress)
		} else if resultMap.ActionType == 1 {
			//received ping
			log.Println("Received Ping from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.GlobalState.List)
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.GlobalState.List)
			s.Ack(resultMap.IpAddress)
		} else if resultMap.ActionType == 2 {
			//received ack
			log.Println("Received Ack from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.GlobalState.List)
			for _, entry := range s.GlobalState.List {
				if entry.InitialTimeStamp == resultMap.InitialTimeStamp && entry.IpAddress == resultMap.IpAddress {
					s.GlobalState.UpdateNode2(resultMap.InitialTimeStamp, resultMap.IpAddress, 0, 0)
					break
				}
			}
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.GlobalState.List)
		} else if resultMap.ActionType == 3{
			log.Println("Received Leave from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.GlobalState.List)
			//received leave
			//s.GlobalState.RemoveNode(incomingIP)
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.GlobalState.List)
		}

	}
*/
}