package server

import "encoding/json"

type Action struct {
	ActionType int // 0: Message, 1: Leave
	Timestamp string
	SenderName string
	SenderIP string
	Metadata string //for Message: then it stores the actual message, for Leave: it stores the failed server address
}

func (a *Action)  ToBytes() []byte {
	res, _ := json.Marshal(a)
	return res
}

func EncodeActionType(actionType string) int {
	if actionType == "Message" {
		return 0
	} else if actionType == "Leave" {
		return 1
	} else if actionType == "Introduce" {
		return 2
	}
	return -1
}
