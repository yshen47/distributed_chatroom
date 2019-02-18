package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mp1/utils"
	"net"
	"sync"
	"time"
)

type SwimServer struct {
	name              string
	MyAddress         string
	portNum           int
	PeopleNum         int
	GlobalServerAddrs [] string
	EstablishedConns  map[string] net.Conn
	Mutex             *sync.Mutex
}

func (s * SwimServer) Constructor(name string, peopleNum int, portNum int, myAddr string, globalServerAddrs [] string) {
	s.MyAddress = myAddr
	s.name = name
	s.EstablishedConns = make(map[string] net.Conn)
	s.portNum = portNum
	s.GlobalServerAddrs = globalServerAddrs
	s.PeopleNum = peopleNum
	s.Mutex = &sync.Mutex{}
}

func (s *SwimServer) DialOthers() {
	isFirst := true
	for {
		if len(s.EstablishedConns) == s.PeopleNum - 1 {
			if isFirst {
				isFirst = false
				//TODO: READY
				log.Println("READY!")
			}
			continue
		}
		for _, ip := range s.GlobalServerAddrs {
			if ip == s.MyAddress {
				time.Sleep(1*time.Second)
				continue
			}
			s.Mutex.Lock()
			_, ok := s.EstablishedConns[ip]
			s.Mutex.Unlock()
			if ok {
				//ip has already established, so skip
				continue
			}
			//fmt.Println("trying to dial ", ip)
			conn, err := net.DialTimeout("tcp", ip, 1*time.Second)
			if err == nil {
				go s.HandleConnection(conn)
			}
		}

	}
}


func (s *SwimServer) HandleConnection(conn net.Conn) {
	var remoteName string
	var remoteAddr string
	action := Action{ActionType:EncodeActionType("Introduce"), SenderIP: s.MyAddress, SenderName:s.name}
	conn.Write(action.ToBytes())

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			//Failure detected
			s.Mutex.Lock()
			//log.Println("Failure detected from ", s.MyAddress, remoteAddr, remoteName)
			log.Println(remoteName, " has left")
			delete(s.EstablishedConns, remoteAddr)
			s.Mutex.Unlock()
			//TODO:send someone left message
			err = conn.Close()
			utils.CheckError(err)
			return
		}
		var resultMap Action
		// parse resultMap to json format
		err = json.Unmarshal(buf[0:n], &resultMap)
		if err != nil {
			fmt.Println("error:", err)
		}
		if resultMap.ActionType == EncodeActionType("Introduce") {
			s.Mutex.Lock()
			_, ok := s.EstablishedConns[resultMap.SenderIP];
			s.Mutex.Unlock()
			if !ok {
				s.Mutex.Lock()
				s.EstablishedConns[resultMap.SenderIP] = conn
				remoteAddr = resultMap.SenderIP
				remoteName = resultMap.SenderName
				log.Println("Established new connection ", resultMap.SenderName, resultMap.SenderIP, " <=> ", s.MyAddress)
				s.Mutex.Unlock()
			} else {
				err = conn.Close()
				utils.CheckError(err)
				return
			}
		} else if resultMap.ActionType == EncodeActionType("Message") {
			//TODO: Print out message

		} else if resultMap.ActionType == EncodeActionType("Leave") {
			s.Mutex.Lock()
			_, ok := s.EstablishedConns[resultMap.Metadata]
			if ok {
				delete(s.EstablishedConns, resultMap.Metadata)
				s.Mutex.Unlock()
				//TODO:send someone left message
				log.Println(resultMap.Metadata, " left.")
			}
			s.Mutex.Unlock()

		}
	}
}