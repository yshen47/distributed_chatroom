package server

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type SwimServer struct {
	name string
	tDetection int64
	tSuspect int64
	tFailure int64
	tLeave int64
	IntroducerIpAddress string
	MembershipList *GlobalState
	MyAddress string
	InitialTimeStamp int64
}

func (s * SwimServer) Constructor(name string, introducerIP string, myIP string) {
	currTimeStamp := time.Now().Unix()
	s.MembershipList = new(GlobalState)
	s.MyAddress = myIP
	s.IntroducerIpAddress = introducerIP
	s.InitialTimeStamp = currTimeStamp
	s.tDetection = 2
	s.tSuspect = 3
	s.tFailure = 3
	s.tLeave = 3
	var entry Entry
	entry.lastUpdatedTime = 0
	entry.EntryType = EncodeEntryType("alive")
	entry.Incarnation = 0
	entry.InitialTimeStamp = currTimeStamp
	entry.IpAddress = myIP
	s.MembershipList.AddNewNode(entry)
}

func (s *SwimServer) StartPing(duration time.Duration) {
	for {
		time.Sleep(duration)
		s.ping()

		s.checkMembershipList()
	}
}

/*
	This function should ping to num processes. And at the same time, it should disseminate entries stored in the disseminateList
 */
func (s *SwimServer) ping() {
	log.Println("Start to ping...")
	targetIndices := s.getPingTargets()
	//fmt.Println("targetIndices", targetIndices)

	for _, index := range targetIndices {
		if s.MembershipList.List[index].lastUpdatedTime != 0 {
			continue
		}
		ipAddress := s.MembershipList.List[index].IpAddress
		s.sendMessageWithUDP("Ping", ipAddress)
		s.MembershipList.List[index].lastUpdatedTime = time.Now().Unix()
	}
	log.Println("server's membership list: ", s.MembershipList.List)
	log.Println("server's blacklist: ", s.MembershipList.printBlackList())
}

/*
	This function should reply to the ping from ipAddress, and disseminate its own disseminateList.
 */
func (s *SwimServer) Ack(ipAddress string) {
	log.Println("Sending ack")
	s.sendMessageWithUDP("Ack", ipAddress)
}


/*
	This function invoke when it attempts to connect with the introducer node. If success, it should update its membership list
 */
func (s *SwimServer) Join() {
	log.Println("Sending join request")
	s.sendMessageWithUDP("Join", s.IntroducerIpAddress)
}

/*
	This function invoke when it leaves the group
 */
func (s *SwimServer) Leave() {
	log.Println("Sending leave request")
	targetIndices := s.getPingTargets()
	s.MembershipList.UpdateNode2(s.InitialTimeStamp, s.MyAddress, 3, 0)
	//s.MembershipList.RemoveNode(s.MyAddress, s.InitialTimeStamp)
	for _, index := range targetIndices {
		ipAddress := s.MembershipList.List[index].IpAddress
		s.sendMessageWithUDP("Leave", ipAddress)
	}
}

func (s *SwimServer) MergeList(receivedRequest Action) {
	log.Println("Start to merge list...")
	for _, entry := range receivedRequest.Record {
		if entry.InitialTimeStamp != s.InitialTimeStamp && entry.IpAddress != s.MyAddress {
			index := s.MembershipList.UpdateNode(entry)
			if index != -1 {
				if s.MyAddress == s.MembershipList.List[index].IpAddress && s.InitialTimeStamp == s.MembershipList.List[index].InitialTimeStamp {
					//only process j can increase its own incarnation number
					s.MembershipList.List[index].Incarnation += 1
					s.MembershipList.List[index].EntryType = 0
				}
			}
		}
	}
}

func (s *SwimServer) checkMembershipList() {
	currTime := time.Now().Unix()
	//check if any process is MembershipList or failed
	for i:= len(s.MembershipList.List)-1; i>=0; i-- {
		entry := s.MembershipList.List[i]
		if entry.EntryType == 0 && currTime - entry.lastUpdatedTime >= s.tDetection&& entry.lastUpdatedTime != 0  {
			//alive now but passed detection timeout
			s.MembershipList.List[i].EntryType += 1
			s.MembershipList.List[i].lastUpdatedTime = 0
		} else if entry.EntryType == 1 && currTime - entry.lastUpdatedTime >= s.tSuspect && entry.lastUpdatedTime != 0 {
			//suspected now but passed suspected timeout
			s.MembershipList.List[i].EntryType += 1
			s.MembershipList.List[i].lastUpdatedTime = currTime
		} else if entry.EntryType == 2 && currTime - entry.lastUpdatedTime >= s.tFailure && entry.lastUpdatedTime != 0 {
			//failed now but passed failure timeout
			s.MembershipList.List = append(s.MembershipList.List[:i], s.MembershipList.List[i+1:]...)
		} else if entry.EntryType == 2 && entry.lastUpdatedTime == 0 {
			s.MembershipList.List = append(s.MembershipList.List[:i], s.MembershipList.List[i+1:]...)
			s.MembershipList.AddToBlacklist(entry)
		} else if entry.EntryType == 3 && currTime - entry.lastUpdatedTime >= s.tLeave {
			s.MembershipList.List = append(s.MembershipList.List[:i], s.MembershipList.List[i+1:]...)
			s.MembershipList.AddToBlacklist(entry)
		}
	}
}

func (s *SwimServer) sendMessageWithUDP ( actionType string, ipAddress string) {
	arr := strings.Split(ipAddress, ":")
	myPort, _ := strconv.Atoi(arr[1])
	Conn, _ := net.DialUDP("udp", nil, &net.UDPAddr{IP:[]byte{127,0,0,1},Port:myPort,Zone:""})
	defer Conn.Close()
	var listToSend []Entry
	for _, v := range s.MembershipList.List {
		//if v.EntryType != 2 {
		listToSend = append(listToSend, v)
		//}
	}
	action := Action{EncodeActionType(actionType), listToSend, s.InitialTimeStamp, s.MyAddress}
	Conn.Write(action.ToBytes())
}

func (s *SwimServer) findSelfInMembershipList() int {
	for ind, entry := range s.MembershipList.List {
		if s.MyAddress == entry.IpAddress && s.InitialTimeStamp == entry.InitialTimeStamp {
			return ind
		}
	}
	log.Fatalln("Fail to find self in membership list.")
	return -1
}

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	var list []int
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}


func (s *SwimServer) getPingTargets() []int {
	var res []int
	currPointer := s.findSelfInMembershipList()
	res = append(res, (currPointer + 1)%len(s.MembershipList.List), (currPointer - 1 + len(s.MembershipList.List))%len(s.MembershipList.List), (currPointer + 2)%len(s.MembershipList.List))
	uniqueRes := unique(res)
	for i, value := range uniqueRes {
		if value == currPointer {
			uniqueRes = append(uniqueRes[:i], uniqueRes[i+1:]...)
			break
		}
	}
	return  uniqueRes
}




