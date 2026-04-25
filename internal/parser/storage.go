package parser

import (
	"net"
	"sync"
	"time"
)

type object struct {
	ExpiresAt time.Time
	Value     string
}

type QueuedCommand struct {
	Args     []string
	Callback func([]string, net.Conn, Config) []byte
}

type stream struct {
	ID   string
	data map[string]string
}

type Config struct {
	Mux             *sync.RWMutex
	Storage         map[string]object
	Lists           map[string][]string
	Streams         map[string][]stream
	Queues          map[net.Conn][]QueuedCommand
	Source          string
	ReplicationConn net.Conn
}

func InitConfig() Config {
	var config Config
	config.Mux = &sync.RWMutex{}
	config.Storage = make(map[string]object)
	config.Lists = make(map[string][]string)
	config.Streams = make(map[string][]stream)
	config.Queues = make(map[net.Conn][]QueuedCommand)
	return config
}
