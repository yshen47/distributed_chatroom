package server

import (
	"fmt"
	"sort"
	"time"
)

type GlobalState struct {
	List [] Entry
	blacklist [] Entry //failed server
}

/*
	@param ipAddress string
	@param incarnation int
	@param entryType int
	Invoke when the server receives response from ping.  Update the membershipList
 */
func (m *GlobalState) UpdateNode(entry Entry) int{
	if m.inBlacklist(entry) {
		return -1
	}
	for i, elem := range m.List {
		if elem.IpAddress == entry.IpAddress && elem.InitialTimeStamp == entry.InitialTimeStamp {
			if entry.EntryType == 1 {
				return i
			}
			if entry.Incarnation > elem.Incarnation {
				m.List[i].Incarnation = entry.Incarnation
				m.List[i].EntryType = entry.EntryType
			} else if entry.Incarnation == elem.Incarnation {
				if entry.EntryType == 1 && entry.EntryType == 2 {
					//suspected or failed or left
					m.List[i].EntryType = entry.EntryType
				} else if entry.EntryType == 3 && elem.EntryType != 3 {
					m.List[i].EntryType = entry.EntryType
					m.List[i].lastUpdatedTime = time.Now().Unix()
					m.AddToBlacklist(entry)
				}
			}
			return -1
		}
	}
	if entry.EntryType != 2 && entry.EntryType != 3 {
		m.AddNewNode(entry)
		m.SortMembership()
	}
	return -1
}
func (m *GlobalState) SortMembership(){

	sort.Slice(m.List, func(i, j int) bool {
		key1 := fmt.Sprintf("%s%d",m.List[i].IpAddress,m.List[i].InitialTimeStamp )
		key2 := fmt.Sprintf("%s%d",m.List[j].IpAddress,m.List[j].InitialTimeStamp )
		return  key1<key2
	})
}
func (m *GlobalState) UpdateNode2(initialTimeStamp int64, ipAddress string, entryType int, lastUpdatedTime int64) {
	for i, elem := range m.List {
		if elem.IpAddress == ipAddress && elem.InitialTimeStamp == initialTimeStamp {
			m.List[i].EntryType = entryType
			m.List[i].lastUpdatedTime = lastUpdatedTime
			return
		}
	}
}

/*
	@param ipAddress string
	@param initialTimeStamp int64
	Invoke when the server receives response from ping.  Update the membershipList
 */
func (m *GlobalState) AddNewNode(entry Entry) {
	//fmt.Println("addnewnode", m.List)
	//fmt.Println(entry)
	if m.ContainsNode(entry) {
		panic("ip address is already in the list")
	}
	m.List = append(m.List, entry)
}

/*
	@param ipAddress string
	@param initialTimeStamp int64
	Invoke when the server receives response from ping.  Update the membershipList and the disseminateList
 */
func (m *GlobalState) RemoveNode(ipAddress string, initialTimeStamp int64) {
	for ind, elem := range m.List {
		if elem.IpAddress == ipAddress && elem.InitialTimeStamp == initialTimeStamp {
			m.List = append(m.List[:ind], m.List[ind+1:]...)
			return
		}
	}
}

func (m *GlobalState) ContainsNode(entry Entry) bool {
	for _, elem := range m.List {
		if elem.IpAddress == entry.IpAddress && elem.InitialTimeStamp == entry.InitialTimeStamp {
			return true
		}
	}
	return false
}

func (m *GlobalState) inBlacklist(entry Entry) bool {
	for _, elem := range m.blacklist {
		if elem.IpAddress == entry.IpAddress && elem.InitialTimeStamp == entry.InitialTimeStamp {
			return true
		}
	}
	return false
}

func (m *GlobalState) AddToBlacklist(entry Entry) {
	if !m.inBlacklist(entry) {
		m.blacklist = append(m.blacklist, entry)
	}
}

func (m *GlobalState) printBlackList() [] Entry {
	return m.blacklist
}
