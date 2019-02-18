package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func send (conn net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text + "\n")
	}

}

func main() {

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")

	go send(conn)

	for {
		// read in input from stdin

		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: "+message)
	}
}
