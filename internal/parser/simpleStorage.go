package parser

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

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
	config.Storage[args[4]] = object{Value: args[6], ExpiresAt: expiresAt}
	config.Mux.Unlock()

	WriteSimpleString(conn, "OK")
	return nil
}

func incrCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)
	fmt.Println(args)

	config.Mux.Lock()
	defer config.Mux.Unlock()

	if config.Storage[args[0]].Value == "" {
		config.Storage[args[0]] = object{Value: "1", ExpiresAt: time.Time{}}

		WriteInteger(conn, 1)
		return nil
	}

	v, err := strconv.Atoi(config.Storage[args[0]].Value)
	if err != nil {
		// not an int
		WriteSimpleError(conn, "value is not an integer or out of range")
		return nil
	}

	config.Storage[args[0]] = object{Value: fmt.Sprintf("%d", v+1), ExpiresAt: config.Storage[args[0]].ExpiresAt}

	WriteInteger(conn, v+1)

	return nil
}

func getCommand(args []string, conn net.Conn, config Config) error {
	// TODO: update all the callbacks that use args to call this helper
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
