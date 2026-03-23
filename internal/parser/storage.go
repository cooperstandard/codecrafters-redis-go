package parser

import (
	"sync"
	"time"
)

type Object interface {
	Type() string
	Expires() time.Time
}

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

type stringObject struct {
	ExpiresAt time.Time
	Value     string
}

func (s stringObject) Type() string {
	return "string"
}

func (s stringObject) Expires() time.Time {
	return s.ExpiresAt
}

type intObject struct {
	ExpiresAt time.Time
	Value     string
}

func (_ intObject) Type() string {
	return "int"
}

func (i intObject) Expires() time.Time {
	return i.ExpiresAt
}

type streamObject struct {
	ExpiresAt time.Time
	Value     []string
}

func (_ streamObject) Type() string {
	return "stream"
}

func (s streamObject) Expires() time.Time {
	return s.ExpiresAt
}
