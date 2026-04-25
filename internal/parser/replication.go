package parser

import (
	"fmt"
	"net"
	"strings"
)

func EstablishReplicaConnection(config Config) (net.Conn, error) {
	buffer := make([]byte, 1024)

	conn, err := net.Dial("tcp", config.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source: %w", err)
	}

	conn.Write(GetStringArray([]string{"PING"}))
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read PING response: %w", err)
	}

	if err = handleHandshakeResponses(buffer[:n], "+PONG\r\n"); err != nil {
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	conn.Write(GetStringArray([]string{"REPLCONF", "listening-port", strings.Split(config.Source, ":")[1]}))
	n, _ = conn.Read(buffer)

	if err = handleHandshakeResponses(buffer[:n], "+OK\r\n"); err != nil {
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	conn.Write(GetStringArray([]string{"REPLCONF", "capa", "eof"}))

	n, _ = conn.Read(buffer)

	if err = handleHandshakeResponses(buffer[:n], "+OK\r\n"); err != nil {
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	return conn, nil
}

func handleHandshakeResponses(actual []byte, expected string) error {
	if string(actual) != expected {
		return fmt.Errorf("unexpected response: %s, expected: %s", string(actual), expected)
	}
	return nil
}
