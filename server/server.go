package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mp1/utils"
	"net"
	"os"
	"sync"
	"time"

)

type Server struct {
	Name              string
	MyAddress         string
	portNum           int
	PeopleNum         int
	GlobalServerAddrs [] string
	EstablishedConns  map[string] net.Conn
	ConnMutex         *sync.Mutex
	ChatMutex         *sync.Mutex
	VectorTimestamp   map[string] int
	messageQueue      [] Message

}

func (s * Server) Constructor(name string, peopleNum int, portNum int, myAddr string, globalServerAddrs [] string) {
	s.MyAddress = myAddr
	s.Name = name
	s.EstablishedConns = make(map[string] net.Conn)
	s.portNum = portNum
	s.GlobalServerAddrs = globalServerAddrs
	s.PeopleNum = peopleNum
	s.ConnMutex = &sync.Mutex{}
	s.ChatMutex = &sync.Mutex{}
	s.VectorTimestamp = make(map[string] int)
	s.VectorTimestamp[s.Name] = 0
}


func (s *Server) DialOthers() {
	isFirst := true
	for {
		if len(s.EstablishedConns) == s.PeopleNum - 1 {
			if isFirst {
				isFirst = false
				//TODO: READY
				s.ChatMutex.Lock()
				log.Println("READY!")
				s.ChatMutex.Unlock()
				go s.startChat()
			}
			continue
		}
		for _, ip := range s.GlobalServerAddrs {
			if ip == s.MyAddress {
				time.Sleep(1*time.Second)
				continue
			}
			s.ConnMutex.Lock()
			_, ok := s.EstablishedConns[ip]
			s.ConnMutex.Unlock()
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

func (s *Server) startChat () {
	for {
		reader := bufio.NewReader(os.Stdin)
		s.ChatMutex.Lock()
		log.Print(": ")
		s.ChatMutex.Unlock()
		text, _ := reader.ReadString('\n')
		// bMulticast
		s.updateVectorTimestamp()
		s.bMuticast("Message", text)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	var remoteName string
	//var remoteAddr string
	s.unicast(conn, "Introduce", "")
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			//Failure detected
			//log.Println("Failure detected from ", s.MyAddress, remoteAddr, remoteName)
			message := utils.Concatenate(remoteName, " left.")
			s.bMuticast("Leave", message)
			err = conn.Close()
			utils.CheckError(err)
			return
		}

		//received something
		var resultMap Action
		// parse resultMap to json format
		err = json.Unmarshal(buf[0:n], &resultMap)

		if err != nil {
			fmt.Println("error:", err)

		}
		if resultMap.ActionType == EncodeActionType("Introduce") {
			s.ConnMutex.Lock()
			_, ok := s.EstablishedConns[resultMap.SenderIP];
			s.ConnMutex.Unlock()
			if !ok {
				s.ConnMutex.Lock()
				s.EstablishedConns[resultMap.SenderIP] = conn
				//remoteAddr = resultMap.SenderIP
				remoteName = resultMap.SenderName
				log.Println("Established new connection ", resultMap.SenderName, resultMap.SenderIP, " <=> ", s.MyAddress)
				s.ConnMutex.Unlock()
			} else {
				err = conn.Close()
				utils.CheckError(err)
				return
			}
		} else if resultMap.ActionType == EncodeActionType("Message") {
			newMessage := Message{Sender: resultMap.SenderName, Content:resultMap.Metadata, Timestamp:resultMap.VectorTimestamp}
			if !s.isMessageReceived(newMessage) {
				s.handleMessage(newMessage)
			}

		} else if resultMap.ActionType == EncodeActionType("Leave") {
			s.ConnMutex.Lock()
			_, ok := s.EstablishedConns[resultMap.Metadata]
			if ok {
				delete(s.EstablishedConns, resultMap.Metadata)
				s.ConnMutex.Unlock()
				s.bMuticast("Leave", utils.Concatenate(resultMap.Metadata))
				s.ChatMutex.Lock()
				log.Println(resultMap.Metadata)
				s.ChatMutex.Unlock()
			}
			s.ConnMutex.Unlock()

		}
	}
}

func (s* Server) isDeliverable(message Message)bool{
	for k,_ := range message.Timestamp {
		if k == message.Sender {
			if message.Timestamp[k] != s.VectorTimestamp[k] + 1 {
				return false
			}
		}else{
			if message.Timestamp[k] > s.VectorTimestamp[k] {
				return false
			}
		}
	}
	return true
}

func (s *Server) isMessageReceived(message Message) bool {
	for _, old := range s.messageQueue {
		for k, v := range old.Timestamp {
			if message.Timestamp[k] != v {
				return false
			}
		}
		if old.Sender == message.Sender {
			return true
		}
	}
	return false
}

func (s * Server)handleMessage(message Message) {
	s.messageQueue = append(s.messageQueue, message)
	deliver := make([]string,10)
	newQueue := make([]Message,10)
	for i:=0;i<len(s.messageQueue);i++{
		if s.isDeliverable(s.messageQueue[i]){
			s.VectorTimestamp[s.messageQueue[i].Sender] += 1
			realContent := utils.Concatenate(s.messageQueue[i].Sender, ": ", s.messageQueue[i].Content)
			fmt.Println(realContent)
			deliver = append(deliver,realContent)

		}else{
			newQueue = append(newQueue,s.messageQueue[i])
		}
	}
	s.messageQueue = newQueue
	for _, message := range deliver {
		if message != "" {
			s.ChatMutex.Lock()
			log.Print(message)
			s.ChatMutex.Unlock()
		}

	}
}


func (s *Server) mergeVectorTimestamp(newTimestamp map[string] int) {
	for k, newVal := range newTimestamp {
		oldVal, ok := s.VectorTimestamp[k]
		if ok {
			if oldVal < newVal {
				s.VectorTimestamp[k] = newVal
			}
		} else {
			s.VectorTimestamp[k] = newVal
		}
	}
}

func (s *Server) updateVectorTimestamp() {
	s.VectorTimestamp[s.Name] += 1
}

func (s *Server) unicast(target net.Conn, actionType string, metaData string) {
	//s.updateVectorTimestamp()
	action := Action{ActionType:EncodeActionType(actionType), SenderIP: s.MyAddress, SenderName:s.Name, Metadata:metaData, VectorTimestamp:s.VectorTimestamp}
	_, err := target.Write(action.ToBytes())
	utils.CheckError(err)
}

func (s *Server) bMuticast(actionType string, metaData string) {
	if EncodeActionType(actionType) == -1 {
		log.Println("Fatal error :actionType doesn't exist.")
		os.Exit(1)
	}
	for _, conn := range s.EstablishedConns {
		action := Action{ActionType:EncodeActionType(actionType), SenderIP: s.MyAddress, SenderName:s.Name, Metadata:metaData, VectorTimestamp: s.VectorTimestamp}
		_, err := conn.Write(action.ToBytes())
		utils.CheckError(err)
	}
}
