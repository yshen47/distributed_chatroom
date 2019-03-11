package server

import (
	"bufio"
	"cs425_mp1/utils"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strings"
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
	messageQueueMutex sync.Mutex

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


				for {
					reader := bufio.NewReader(os.Stdin)
					s.ChatMutex.Lock()
					log.Print(": ")
					s.ChatMutex.Unlock()
					text, _ := reader.ReadString('\n')
					text = strings.TrimSuffix(text, "\n")
					if len(text) == 0 {
						continue
					}
					// bMulticast
					//s.updateVectorTimestamp()
					content := utils.Concatenate(s.Name, " ", text)
					messageTimestamp := make(map[string]int)

					for k, v := range s.VectorTimestamp {
						messageTimestamp[k] = v
					}
					newMessage := Message{Content:content, Sender:s.Name, Timestamp:messageTimestamp}
					newMessage.Timestamp[s.Name] += 1
					s.handleMessage(newMessage)
					s.bMuticast("Message", content)
				}

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
			//fmt.Println("dial error? ",err)
			if err == nil {
				//fmt.Println("suceess ip: ",ip)
				go s.HandleConnection(conn)
			}
		}
	}
}


func (s *Server) HandleConnection(conn net.Conn) {
	var remoteName string
	var remoteAddr string
	s.unicast(conn, "Introduce", "")
	for {
		buf := make([]byte, 16*1024)
		n, err := conn.Read(buf)
		//fmt.Println("read error = ",err)
		if err == io.EOF {
			//Failure detected
			//log.Println("Failure detected from ", s.MyAddress, remoteAddr, remoteName)
			s.ChatMutex.Lock()
			_, ok := s.EstablishedConns[remoteAddr]
			if ok {
				log.Println(remoteName, " left!")
			}
			s.ChatMutex.Unlock()

			s.ConnMutex.Lock()
			delete(s.EstablishedConns, remoteAddr)
			s.ConnMutex.Unlock()
			s.bMuticast("Leave", utils.Concatenate(remoteName, ";", remoteAddr))
			err = conn.Close()
			utils.CheckError(err)
			return
		}

		//received something
		var resultMap Action
		// parse resultMap to json format
		//fmt.Println("Received new resultMap, START", string(buf[0:n]), "END")
		err = json.Unmarshal(buf[0:n], &resultMap)
		utils.CheckError(err)
		if resultMap.ActionType == EncodeActionType("Introduce") {
			s.ConnMutex.Lock()
			_, ok := s.EstablishedConns[resultMap.SenderIP];
			s.ConnMutex.Unlock()
			if !ok {
				s.ConnMutex.Lock()
				s.EstablishedConns[resultMap.SenderIP] = conn
				remoteAddr = resultMap.SenderIP
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
			//add multicast here?
			if !s.isMessageReceived(newMessage) {
				s.handleMessage(newMessage)
				s.bMuticast("Message", resultMap.Metadata)
			}

		} else if resultMap.ActionType == EncodeActionType("Leave") {
			deleteRemoteAddr := strings.Split(resultMap.Metadata, ";")[1]
			deleteRemoteName := strings.Split(resultMap.Metadata,";")[0]

			s.ChatMutex.Lock()
			_, ok := s.EstablishedConns[deleteRemoteName]
			if ok {
				log.Println(remoteName, " left!")
			}
			s.ChatMutex.Unlock()

			s.ConnMutex.Lock()
			_, ok = s.EstablishedConns[deleteRemoteAddr]
			if ok {
				delete(s.EstablishedConns, deleteRemoteAddr)
				s.ConnMutex.Unlock()
				s.bMuticast("Leave", resultMap.Metadata)
			} else {
				s.ConnMutex.Unlock()
			}
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
	origSenderName := strings.Split(message.Content, " ")[0]
	s.messageQueueMutex.Lock()
	defer s.messageQueueMutex.Unlock()
	for _, old := range s.messageQueue {
		_, ok := old.Timestamp[origSenderName]
		if ok {
			if message.Timestamp[origSenderName] == old.Timestamp[origSenderName] {
				return true
			}
		}
	}

	if message.Timestamp[origSenderName] <= s.VectorTimestamp[origSenderName] {
		return true
	}



	return false
}

func (s * Server)handleMessage(message Message) {
	s.messageQueueMutex.Lock()
	s.messageQueue = append(s.messageQueue, message)
	deliver := make([]string,0)
	newQueue := make([]Message,0)
	for i:=0;i<len(s.messageQueue);i++{
		if s.isDeliverable(s.messageQueue[i]) {
			s.VectorTimestamp[s.messageQueue[i].Sender] += 1
			j := strings.Index(s.messageQueue[i].Content, " ")
			realContent := utils.Concatenate(s.messageQueue[i].Sender, ": ", s.messageQueue[i].Content[j+1:])
			deliver = append(deliver,realContent)

		}else{
			newQueue = append(newQueue,s.messageQueue[i])
		}
	}
	s.messageQueue = newQueue
	s.messageQueueMutex.Unlock()
	for _, message := range deliver {
		if message != "" {
			s.ChatMutex.Lock()
			log.Println(message)
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

	for k, conn := range s.EstablishedConns {
		if k == s.MyAddress {
			continue
		}
		action := Action{ActionType:EncodeActionType(actionType), SenderIP: s.MyAddress, SenderName:s.Name, Metadata:metaData, VectorTimestamp: s.VectorTimestamp}
		_, err := conn.Write(action.ToBytes())
		utils.CheckError(err)
	}
}
