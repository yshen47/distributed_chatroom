package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mp1/server"
	"mp1/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Print("Usage: go run main.go [server name] [my ip address] [introducer ip address]")
		return
	}
	//Parse input argument
	name := os.Args[1]
	myAddress := os.Args[2]
	introducerAddress := os.Args[3]
	utils.SetupLog(name)
	log.Println("Start server with the: ", name, myAddress, introducerAddress )


	s := new(server.SwimServer)
	s.Constructor(name, introducerAddress,myAddress)

	//Join group except for introducer itself
	if myAddress != introducerAddress {
		log.Println("Start joining the group",myAddress)
		s.Join()
	} else {
		log.Println("Introducer from",myAddress, "starts the group")
	}


	//Start the server
	arr := strings.Split(myAddress, ":")
	myPort, _ := strconv.Atoi(arr[1])
	ServerConn, err := net.ListenUDP("udp", &net.UDPAddr{IP:[]byte{0,0,0,0},Port:myPort,Zone:""})
	//ServerConn, err := net.Dial("udp", &net.UDPAddr{IP:[]byte{0,0,0,0},Port:myPort,Zone:""})
	//ServerConn, err := net.Dial("udp",myAddress)
	if err != nil {
		log.Println(err)
		return
	}

	defer ServerConn.Close()

	go s.StartPing(1 * time.Second)

	//wait for incoming response
	buf := make([]byte, 1024)

	for {
		n, _ := ServerConn.Read(buf)
		var resultMap server.Action
		// parse resultMap to json format
		json.Unmarshal(buf[0:n], &resultMap)


		//Customize different action
		if resultMap.ActionType == 0 {
			//received join
			log.Println("Received Join from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.MembershipList.List)
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.MembershipList.List)
			s.Ack(resultMap.IpAddress)
		} else if resultMap.ActionType == 1 {
			//received ping
			log.Println("Received Ping from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.MembershipList.List)
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.MembershipList.List)
			s.Ack(resultMap.IpAddress)
		} else if resultMap.ActionType == 2 {
			//received ack
			log.Println("Received Ack from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.MembershipList.List)
			for _, entry := range s.MembershipList.List {
				if entry.InitialTimeStamp == resultMap.InitialTimeStamp && entry.IpAddress == resultMap.IpAddress {
					s.MembershipList.UpdateNode2(resultMap.InitialTimeStamp, resultMap.IpAddress, 0, 0)
					break
				}
			}
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.MembershipList.List)
		} else if resultMap.ActionType == 3{
			log.Println("Received Leave from ", resultMap.IpAddress)
			log.Println("Data received:", resultMap.Record)
			log.Println("server's membership list: ", s.MembershipList.List)
			//received leave
			//s.MembershipList.RemoveNode(incomingIP)
			s.MergeList(resultMap)
			log.Println("After merging, server's membership list", s.MembershipList.List)
		}

	}

}