package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/parser"
)

const BufferSize int = 1024

func main() {
	fmt.Println("Logs from your program will appear here!")
	port := 6379

	args := make(map[string]string)
	if len(os.Args) > 1 {
		args = parseFlags(os.Args[1:])
	}

	for k, v := range args {
		switch k {
		case "--port":
			port, _ = strconv.Atoi(v)
			fmt.Printf("Using port: %d\n", port)

		case "--replicaof":
			fmt.Printf("Replication flag set to: %s\n", v)

		default:
			fmt.Printf("Unknown flag: %s\n", k)
		}
	}

	port, _ = strconv.Atoi(os.Args[2])

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
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
		if _, exists := config.Queues[conn]; exists && !parser.DoesCommandEndTransaction(command) {
			config.Queues[conn] = append(config.Queues[conn], parser.QueuedCommand{Args: args, Callback: command.Callback})
			conn.Write(parser.GetSimpleString("QUEUED"))
		} else {
			output := command.Callback(args, conn, config)
			conn.Write(output)
		}
	}
}

func parseFlags(flags []string) map[string]string {
	// val := prior[:2]
	// parse the flags slice (which is in the form of [name, value, name, value, etc]) and return a map of flag name to value
	ret := make(map[string]string)
	for i := 0; i < len(flags); i += 2 {
		ret[flags[i]] = flags[i+1]
	}
	return ret
}
