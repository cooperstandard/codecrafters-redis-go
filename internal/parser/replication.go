package parser

import (
	"fmt"
	"net"
	"strings"
)

func EstablishReplicaConnection(config Config) (net.Conn, error) {
	conn, err := net.Dial("tcp", config.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source: %w", err)
	}
	conn.Write(GetStringArray([]string{"PING"}))

	conn.Write(GetStringArray([]string{"REPLCONF", "listening-port", strings.Split(config.Source, ":")[1]}))

	conn.Write(GetStringArray([]string{"REPLCONF", "capa", "eof"}))

	return conn, nil
}
