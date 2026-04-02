package parser

import (
	"fmt"
	"net"
)

func multiCommand(args []string, conn net.Conn, config Config) []byte {
	config.Queues[conn] = make([]QueuedCommand, 0)
	return GetSimpleString("OK")
}

func execCommand(args []string, conn net.Conn, config Config) []byte {
	fmt.Println("here")
	if _, exists := config.Queues[conn]; !exists {
		return GetSimpleError("EXEC without MULTI")
	}

	if len(config.Queues[conn]) == 0 {
		delete(config.Queues, conn)
		return GetStringArray([]string{})
	}
	outputs := [][]byte{}
	for _, v := range config.Queues[conn] {
		outputs = append(outputs, v.Callback(v.Args, conn, config))
	}

	return prefixAndFlattenArray(outputs)
}

func discardCommand(args []string, conn net.Conn, config Config) []byte {
	if _, exists := config.Queues[conn]; !exists {
		return GetSimpleError("DISCARD without MULTI")
	}

	delete(config.Queues, conn)
	return GetSimpleString("OK")
}

func prefixAndFlattenArray(content [][]byte) []byte {
	ret := fmt.Appendf(nil, "*%d\r\n", len(content))
	for _, v := range content {
		ret = append(ret, v...)
	}
	return ret
}
