package server

type Message struct {
	Content   string
	Sender    string
	Timestamp map[string]int
}
