package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func listen (conn net.Conn) {
	for {
		// read in input from stdin

		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: "+message)
	}
}

func main() {

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")
	go listen(conn)

}
