package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/parser"
)

const BufferSize int = 1024

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

	config := parser.InitConfig()

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config parser.Config) {
	defer func() {
		conn.Close()
	}()

	buffer := make([]byte, BufferSize)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		cmd := buffer[:n]
		command, args := parser.ParseString(cmd)
		if _, exists := config.Queues[conn]; exists {
			config.Queues[conn] = append(config.Queues[conn], parser.QueuedCommand{ Args: args, Callback: command.Callback})
			conn.Write(parser.GetSimpleString("QUEUED"))
			return
		} else {
			output := command.Callback(args, conn, config)
			conn.Write(output)
		}
	}
}
