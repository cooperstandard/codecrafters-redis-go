package parser

import (
	"fmt"
	"net"
)

func multiCommand(args []string, conn net.Conn, config Config) []byte {
	args = GetArgs(args)
	fmt.Println(args)
	config.Queues[conn] = make([]Command, 0)
	return GetSimpleString("OK")
}

func execCommand(args []string, conn net.Conn, config Config) []byte {
	// TODO: return the byte slice
	if _, exists := config.Queues[conn]; !exists {
		return GetSimpleError(conn, "EXEC without MULTI")
	}

	if len(config.Queues[conn]) == 0 {
		delete(config.Queues, conn)
		return GetStringArray([]string{})
	}

	return GetSimpleString("OK")
}
