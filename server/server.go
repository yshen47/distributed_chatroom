package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mp1/utils"
	"net"
	"time"
)

type SwimServer struct {
	name               string
	tDetection         int64
	tSuspect           int64
	tFailure           int64
	tLeave             int64
	GlobalState        *GlobalState
	MyAddress          string
	portNum            int
	InitialTimeStamp   int64
	GlobalServerAddrs  [] string
	EstablishedInConns map[string] net.Conn
	EstablishedOutConns map[string] net.Conn
}

func (s * SwimServer) Constructor(name string, peopleNum int, portNum int, myAddr string, globalServerAddrs [] string) {
	currTimeStamp := time.Now().Unix()
	s.GlobalState = new(GlobalState)
	s.MyAddress = myAddr
	s.InitialTimeStamp = currTimeStamp
	s.tDetection = 2
	s.tSuspect = 3
	s.tFailure = 3
	s.tLeave = 3
	s.EstablishedInConns = make(map[string] net.Conn)
	s.EstablishedOutConns = make(map[string] net.Conn)
	s.portNum = portNum
	var entry Entry
	entry.lastUpdatedTime = 0
	entry.EntryType = EncodeEntryType("alive")
	entry.Incarnation = 0
	entry.InitialTimeStamp = currTimeStamp
	entry.IpAddress = s.MyAddress
	s.GlobalState.AddNewNode(entry)
	s.GlobalServerAddrs = globalServerAddrs
}

func (s *SwimServer) DialOthers(c chan ConnectionPair)  map[string]net.Conn {
	fmt.Println(s.MyAddress)
	for {
		for _, ip := range s.GlobalServerAddrs {
			if ip == s.MyAddress {
				time.Sleep(1*time.Second)
				continue
			}
			if _, ok := s.EstablishedOutConns[ip]; ok {
				//ip has already established, so skip
				continue
			}
			//fmt.Println("trying to dial ", ip)
			conn, err := net.DialTimeout("tcp", ip, 1*time.Second)
			if err == nil {
				log.Println("Established new out going connection to ", ip)
				s.EstablishedOutConns[ip] = conn
			}
			time.Sleep(1*time.Second)
		}
	}
}


func (s *SwimServer) HandleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			//Failure detected
			for k, v := range s.EstablishedInConns {
				if v == conn {
					delete(s.EstablishedInConns, k)
					err = conn.Close()
					utils.CheckError(err)
					return
				}
			}
		}
		var resultMap Action
		// parse resultMap to json format
		err = json.Unmarshal(buf[0:n], &resultMap)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
}

func (s *SwimServer) SendMessageWithTCP(message string) {
	for ip := range s.EstablishedOutConns {

	}
}