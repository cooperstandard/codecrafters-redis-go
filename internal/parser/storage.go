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

type stream struct {
	ID   string
	data map[string]string
}

type Config struct {
	Mux     *sync.RWMutex
	Storage map[string]object
	Lists   map[string][]string
	Streams map[string][]stream
	Queues  map[net.Conn][]Command
}

func InitConfig() Config {
	var config Config
	config.Mux = &sync.RWMutex{}
	config.Storage = make(map[string]object)
	config.Lists = make(map[string][]string)
	config.Streams = make(map[string][]stream)
	config.Queues = make(map[net.Conn][]Command)
	return config
}
