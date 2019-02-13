package server

import "encoding/json"

type Action struct {
	ActionType int // 0: join, 1: ping, 2: ack 3: leave
	Record []Entry
	InitialTimeStamp int64
	IpAddress string
}

func (a *Action)  ToBytes() []byte {
	res, _ := json.Marshal(a)
	return res
}

func EncodeActionType(actionType string) int {
	if actionType == "Join" {
		return 0
	} else if actionType == "Ping" {
		return 1
	} else if actionType == "Ack" {
		return 2
	} else if actionType == "Leave" {
		return 3
	}
	return -1
}
