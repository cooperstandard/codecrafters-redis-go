package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	var err error = nil

	buffer := make([]byte, 1024)

	for {
		_, err = conn.Read(buffer)
		if err != nil {
			break
		}

		conn.Write(byteEncodeString("PONG"))
	}
}

func byteEncodeString(input string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", input)
}

func byteDecodeString(input []byte) string {
	return string(input)
}
