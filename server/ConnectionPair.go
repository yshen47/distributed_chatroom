package server

import "net"

type ConnectionPair struct{
	ip string
	conn net.Conn
}
