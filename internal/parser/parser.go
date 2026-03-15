package parser

import (
	"fmt"
	"net"
	"strings"
)

func ByteEncodeString(input string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", input)
}

func WriteString(conn net.Conn, val string) {
	fmt.Fprintf(conn, "+%s\r\n", val)
}

func ByteDecodeString(input []byte) string {
	return string(input)
}

type Command struct {
	Command  string
	Callback func([]string, net.Conn) error
}

var Commands = map[string]Command{
	"ping": {
		Command:  "ping",
		Callback: PingCommand,
	},
	"null": {
		Command:  "",
		Callback: nullCommand,
	},
	"echo": {
		Command:  "echo",
		Callback: echoCommand,
	},
}

func ParseString(cmd []byte) (Command, []string) {
	str := strings.Split(string(cmd), "\r\n")

	return Commands[strings.ToLower(str[2])], str
}

func nullCommand(_args []string, _conn net.Conn) error {
	return nil
}

func echoCommand(args []string, conn net.Conn) error {
	WriteBulkString(conn, args[4])
	return nil
}

func PingCommand(_ []string, conn net.Conn) error {
	WriteString(conn, "PONG")
	return nil
}

func WriteBulkString(conn net.Conn, val string) {
	fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
}
