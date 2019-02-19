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
	name              string
	MyAddress         string
	portNum           int
	PeopleNum         int
	GlobalServerAddrs [] string
	EstablishedConns  map[string] net.Conn
	Mutex             *sync.Mutex
	VectorTimestamp   map[string] int
	messageQueue      [] Message

}

func (s * Server) Constructor(name string, peopleNum int, portNum int, myAddr string, globalServerAddrs [] string) {
	s.MyAddress = myAddr
	s.name = name
	s.EstablishedConns = make(map[string] net.Conn)
	s.portNum = portNum
	s.GlobalServerAddrs = globalServerAddrs
	s.PeopleNum = peopleNum
	s.Mutex = &sync.Mutex{}
	s.VectorTimestamp = make(map[string] int)
	s.VectorTimestamp[s.MyAddress] = 0
}


func (s *Server) DialOthers() {
	isFirst := true
	for {
		if len(s.EstablishedConns) == s.PeopleNum - 1 {
			if isFirst {
				isFirst = false
				//TODO: READY
				log.Println("READY!")
				go s.startChat()
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

func (s *Server) startChat () {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// bMulticast
		s.updateVectorTimestamp()
		s.bMuticast("Message", text)
		log.Println(text)
	}

}

func (s *Server) HandleConnection(conn net.Conn) {
	var remoteName string
	var remoteAddr string
	s.unicast(conn, "Introduce", "")
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			//Failure detected
			s.Mutex.Lock()
			//log.Println("Failure detected from ", s.MyAddress, remoteAddr, remoteName)
			message := utils.Concatenate(remoteName, " left.")
			delete(s.EstablishedConns, remoteAddr)
			s.Mutex.Unlock()
			s.bMuticast("Leave", message)
			log.Println(message)
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
			s.handleMessage(Message{sender:resultMap.SenderName, content:resultMap.Metadata, timestamp:resultMap.VectorTimestamp})
		} else if resultMap.ActionType == EncodeActionType("Leave") {
			s.Mutex.Lock()
			_, ok := s.EstablishedConns[resultMap.Metadata]
			if ok {
				delete(s.EstablishedConns, resultMap.Metadata)
				s.Mutex.Unlock()
				s.bMuticast("Leave", utils.Concatenate(resultMap.Metadata))
				log.Println(resultMap.Metadata)
			}
			s.Mutex.Unlock()

		}
	}
}

func (s* Server)checktimestamp(message Message)bool{
	var flag bool = true
	for k,_ := range message.timestamp{
		if k == message.sender{
			if message.timestamp[k] != s.VectorTimestamp[k]{
				flag = false
				break
			}
		}else{
			if message.timestamp[k] > s.VectorTimestamp[k] {
				flag = false
				break
			}
		}
	}
	return flag
}


func (s * Server)handleMessage(message Message) []string{

	s.messageQueue = append(s.messageQueue, message)
	var newQueue []Message
	var deliver []string
	deliver = make([]string,10)
	newQueue = make([]Message,10)
	for i:=0;i<len(s.messageQueue);i++{
		if s.checktimestamp(s.messageQueue[i]){
			s.VectorTimestamp[s.messageQueue[i].sender] += 1
			deliver = append(deliver,s.messageQueue[i].content)

		}else{
			newQueue = append(newQueue,s.messageQueue[i])
		}
	}

	return deliver


	//able to deliever message immediately
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
	s.VectorTimestamp[s.MyAddress] += 1
}

func (s *Server) unicast(target net.Conn, actionType string, metaData string) {
	//s.updateVectorTimestamp()
	action := Action{ActionType:EncodeActionType(actionType), SenderIP: s.MyAddress, SenderName:s.name, Metadata:metaData, VectorTimestamp:s.VectorTimestamp}
	_, err := target.Write(action.ToBytes())
	utils.CheckError(err)
}

func (s *Server) bMuticast(actionType string, metaData string) {
	if EncodeActionType(actionType) == -1 {
		log.Println("Fatal error :actionType doesn't exist.")
		os.Exit(1)
	}
	for _, conn := range s.EstablishedConns {
		action := Action{ActionType:EncodeActionType(actionType), SenderIP: s.MyAddress, SenderName:s.name, Metadata:metaData, VectorTimestamp: s.VectorTimestamp}
		_, err := conn.Write(action.ToBytes())
		utils.CheckError(err)
	}
}
