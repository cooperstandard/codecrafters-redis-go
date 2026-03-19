package parser

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Object struct {
	ExpiresAt time.Time
	Value     string
}

type Config struct {
	Mux     *sync.RWMutex
	Storage map[string]Object
	Lists   map[string][]string
}

func InitConfig() Config {
	var config Config
	config.Mux = &sync.RWMutex{}
	config.Storage = make(map[string]Object)
	config.Lists = make(map[string][]string)
	return config
}

func ByteEncodeString(input string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", input)
}

func WriteSimpleString(conn net.Conn, val string) {
	fmt.Fprintf(conn, "+%s\r\n", val)
}

func ByteDecodeString(input []byte) string {
	return string(input)
}

type Command struct {
	Command  string
	Callback func([]string, net.Conn, Config) error
}

func GetArgs(raw []string) []string {
	var ret []string

	for i := 4; i < len(raw); i += 2 {
		ret = append(ret, raw[i])
	}
	return ret
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
		Command: "rpush",
		Callback: rpushCommand,
	},
}

func ParseString(cmd []byte) (Command, []string) {
	str := strings.Split(string(cmd), "\r\n")

	return Commands[strings.ToLower(str[2])], str
}

func nullCommand(_args []string, _conn net.Conn, _config Config) error {
	return nil
}

func echoCommand(args []string, conn net.Conn, _config Config) error {
	WriteBulkString(conn, args[4])
	return nil
}

func rpushCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	config.Lists[args[0]] = append(config.Lists[args[0]], args[1:]...)
	WriteInteger(conn, len(config.Lists[args[0]]))

	return nil
}

func PingCommand(_args []string, conn net.Conn, _config Config) error {
	WriteSimpleString(conn, "PONG")
	return nil
}

func setCommand(args []string, conn net.Conn, config Config) error {
	expiresAt := time.Time{}
	if len(args) >= 10 {
		dur, err := strconv.Atoi(args[10])
		if err != nil {
			return err
		}
		if args[8] == "EX" {
			expiresAt = time.Now().Add(time.Second * time.Duration(dur))
		} else {
			expiresAt = time.Now().Add(time.Millisecond * time.Duration(dur))
		}
	}

	config.Mux.Lock()
	config.Storage[args[4]] = Object{Value: args[6], ExpiresAt: expiresAt}
	config.Mux.Unlock()

	WriteSimpleString(conn, "OK")
	return nil
}

func getCommand(args []string, conn net.Conn, config Config) error {
	fmt.Println(args)
	config.Mux.RLock()
	val, exists := config.Storage[args[4]]
	config.Mux.RUnlock()

	// remove value if expired
	if exists && !val.ExpiresAt.IsZero() && time.Now().After(val.ExpiresAt) {
		exists = false
		config.Mux.Lock()
		delete(config.Storage, args[4])
		config.Mux.Unlock()
	}

	if !exists {
		WriteBulkString(conn, "")
		return nil
	}

	WriteBulkString(conn, val.Value)
	return nil
}

func WriteBulkString(conn net.Conn, val string) {
	if len(val) == 0 {
		fmt.Fprintf(conn, "$-1\r\n")
		return
	}
	fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
}

func WriteInteger(conn net.Conn, val int) {
	fmt.Fprintf(conn, ":%d\r\n", val)
}
