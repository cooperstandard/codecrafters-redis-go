package parser

import (
	"net"
	"strings"
)

type Command struct {
	Command  string
	Callback func([]string, net.Conn, Config) error
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
	"set": {
		Command:  "set",
		Callback: setCommand,
	},
	"get": {
		Command:  "get",
		Callback: getCommand,
	},
	"rpush": {
		Command:  "rpush",
		Callback: rpushCommand,
	},
	"lrange": {
		Command:  "lrange",
		Callback: lrangeCommand,
	},
	"lpush": {
		Command:  "lpush",
		Callback: lpushCommand,
	},
	"llen": {
		Command:  "llen",
		Callback: llenCommand,
	},
	"lpop": {
		Command:  "lpop",
		Callback: lpopCommand,
	},
	"blpop": {
		Command:  "blpop",
		Callback: blpopCommand,
	},
	"type": {
		Command: "type",
		Callback: typeCommand,
	},
	"xadd": {
		Command: "xadd",
		Callback: xaddCommand,
	},
	"xrange": {
		Command: "xrange",
		Callback: xrangeCommand,
	},
	"xread": {
		Command: "xread",
		Callback: xreadCommand,
	},
	"incr": {
		Command: "incr",
		Callback: incrCommand,
	},
	"multi": {
		Command: "multi",
		Callback: multiCommand,
	},
}

func ParseString(cmd []byte) (Command, []string) {
	str := strings.Split(string(cmd), "\r\n")

	return Commands[strings.ToLower(str[2])], str
}

