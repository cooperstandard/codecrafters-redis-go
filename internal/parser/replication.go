package parser

import (
	"fmt"
	"net"
)

func EstablishReplicaConnection(config Config) (net.Conn, error) {
	conn, err := net.Dial("tcp", config.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source: %w", err)
	}
	conn.Write(GetStringArray([]string{"PING"}))

	return conn, nil
}
