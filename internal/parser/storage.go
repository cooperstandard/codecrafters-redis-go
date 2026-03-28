package parser

import (
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
}

func InitConfig() Config {
	var config Config
	config.Mux = &sync.RWMutex{}
	config.Storage = make(map[string]object)
	config.Lists = make(map[string][]string)
	config.Streams = make(map[string][]stream)
	return config
}
