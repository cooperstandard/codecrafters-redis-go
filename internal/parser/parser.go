package parser

import (
	"fmt"
	"net"
	"slices"
	"strconv"
	"time"
)

func GetArgs(raw []string) []string {
	var ret []string

	for i := 4; i < len(raw); i += 2 {
		ret = append(ret, raw[i])
	}
	return ret
}

func nullCommand(_args []string, _conn net.Conn, _config Config) error {
	return nil
}

func echoCommand(args []string, conn net.Conn, _config Config) error {
	WriteBulkString(conn, args[4])
	return nil
}

func typeCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)
	if _, ok := config.Storage[args[0]]; ok {
		WriteSimpleString(conn, "string")
		return nil
	}
	if _, ok := config.Lists[args[0]]; ok {
		WriteSimpleString(conn, "list")
		return nil
	}
	if _, ok := config.Streams[args[0]]; ok {
		WriteSimpleString(conn, "stream")
		return nil
	}

	WriteSimpleString(conn, "none")
	return nil
}

func blpopCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	if len(args) < 2 {
		WriteEmptyArray(conn)
		return nil
	}

	tryPop := func() bool {
		// NOTE: this is inefficient use of locks but in reality 90% of the time any go routine is waiting will be sleeping, not waiting for the lock
		config.Mux.Lock()
		defer config.Mux.Unlock()
		if len(config.Lists[args[0]]) > 0 {
			items := []string{args[0], config.Lists[args[0]][:1][0]}
			config.Lists[args[0]] = config.Lists[args[0]][1:]
			WriteStringArray(conn, items)
			return true
		}
		return false
	}

	if tryPop() {
		return nil
	}

	if args[1] == "0" {
		for {
			time.Sleep(time.Millisecond * 10)
			if tryPop() {
				return nil
			}
		}
	}

	timeoutSeconds, _ := strconv.ParseFloat(args[1], 64)

	deadline := time.Now().Add(time.Duration(timeoutSeconds * float64(time.Second)))
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		if tryPop() {
			return nil
		}
	}

	WriteEmptyArray(conn)
	return nil
}

func lpopCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	if len(args) == 1 {

		config.Mux.Lock()
		if len(config.Lists[args[0]]) == 0 {
			WriteBulkString(conn, "")
		} else {
			item := config.Lists[args[0]][0]
			config.Lists[args[0]] = config.Lists[args[0]][1:]
			WriteBulkString(conn, item)
		}
		config.Mux.Unlock()
	} else {
		config.Mux.Lock()
		if len(config.Lists[args[0]]) == 0 {
			WriteBulkString(conn, "")
		} else {
			pivot, _ := strconv.Atoi(args[1])
			pivot = min(len(config.Lists[args[0]])-1, pivot)
			items := config.Lists[args[0]][:pivot]
			config.Lists[args[0]] = config.Lists[args[0]][pivot:]
			WriteStringArray(conn, items)
		}
		config.Mux.Unlock()
	}

	return nil
}

func xrangeCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	fmt.Println(args)

	config.Mux.RLock()
	defer config.Mux.RUnlock()

	start := args[1]
	end := args[2]

	matched := []stream{}

	s := config.Streams[args[0]]

	for _, v := range s {
		inRange := StreamIDCompare(start, v.ID) != 1
		inRange = inRange && StreamIDCompare(end, v.ID) != -1
		if inRange {
			matched = append(matched, v)
		}
	}

	WriteStreamSlice(conn, matched)
	return nil
}

func xreadBlocking(args []string, conn net.Conn, config Config) error {
	// replace all the dollars with real ids
	func() {
		streamKeys := args[3 : (len(args))/2+2]

		streamIDs := args[(len(args))/2+2:]

		for i, id := range streamIDs {
			if id == "$" {
				realIndex := (len(args))/2 + 2 + i
				s := config.Streams[streamKeys[i]]
				if len(s) == 0 {
					args[realIndex] = "0-0"
					continue
				}
				maxSeen := "0-0"
				for _, v := range s {
					if StreamIDCompare(v.ID, maxSeen) == 1 {
						maxSeen = v.ID
					}
				}
				args[realIndex] = maxSeen
			}
		}
	}()

	tryRead := func() bool {
		config.Mux.RLock()
		defer config.Mux.RUnlock()

		streamKeys := args[3 : (len(args))/2+2]

		streamIDs := args[(len(args))/2+2:]
		allMatched := [][]stream{}

		for i, key := range streamKeys {
			start := streamIDs[i]

			matched := []stream{}

			s := config.Streams[key]

			for _, v := range s {
				inRange := StreamIDCompare(start, v.ID) == -1
				if inRange {
					matched = append(matched, v)
				}
			}

			if len(matched) > 0 {
				allMatched = append(allMatched, matched)
			}
		}

		if len(allMatched) > 0 {
			WriteStreamSliceWithName(conn, allMatched, streamKeys)
			return true
		}

		return false
	}

	if tryRead() {
		return nil
	}

	if args[1] == "0" {
		for {
			time.Sleep(time.Millisecond * 10)
			if tryRead() {
				return nil
			}
		}
	}

	timeoutSeconds, _ := strconv.ParseFloat(args[1], 64)

	deadline := time.Now().Add(time.Duration(time.Duration(timeoutSeconds) * time.Millisecond))
	for time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		if tryRead() {
			return nil
		}
	}

	WriteEmptyArray(conn)
	return nil
}

func xreadCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)
	if args[0] == "block" {
		xreadBlocking(args, conn, config)
		return nil
	}

	args = args[1:]
	streamKeys := args[:(len(args))/2]

	streamIDs := args[(len(args))/2:]
	fmt.Printf("streamIDs: %v\n", streamIDs)

	config.Mux.RLock()
	defer config.Mux.RUnlock()

	allMatched := [][]stream{}

	for i, key := range streamKeys {
		start := streamIDs[i]

		matched := []stream{}

		s := config.Streams[key]

		for _, v := range s {
			inRange := StreamIDCompare(start, v.ID) != 1
			if inRange {
				matched = append(matched, v)
			}
		}

		allMatched = append(allMatched, matched)
	}

	WriteStreamSliceWithName(conn, allMatched, streamKeys)
	return nil
}

func xaddCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	var ok bool
	args[1], ok = validateAndGenerateID(conn, config, args[1], args[0])
	if !ok {
		return nil
	}

	config.Mux.Lock()
	streamEntry := stream{ID: args[1]}
	streamEntry.data = make(map[string]string)
	for i := 2; i < len(args); i += 2 {
		streamEntry.data[args[i]] = args[i+1]
	}
	config.Streams[args[0]] = append(config.Streams[args[0]], streamEntry)
	config.Mux.Unlock()

	WriteBulkString(conn, args[1])

	return nil
}

func llenCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	config.Mux.RLock()
	WriteInteger(conn, len(config.Lists[args[0]]))
	config.Mux.RUnlock()

	return nil
}

func lrangeCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)
	start, err := strconv.Atoi(args[1])
	if err != nil {
		WriteStringArray(conn, []string{})
		return err
	}

	end, _ := strconv.Atoi(args[2])

	config.Mux.RLock()
	list := config.Lists[args[0]]

	if start < 0 {
		start = len(list) + start
		start = max(start, 0)
	}

	end = min(end, len(list)-1)

	if end < 0 {
		end = len(list) + end
		start = max(start, 0)
	}

	if start > end {
		WriteStringArray(conn, []string{})
		config.Mux.RUnlock()
		return nil
	}

	WriteStringArray(conn, list[start:end+1])
	config.Mux.RUnlock()
	return nil
}

func lpushCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	config.Mux.Lock()
	elems := args[1:]
	slices.Reverse(elems)
	config.Lists[args[0]] = append(elems, config.Lists[args[0]]...)
	WriteInteger(conn, len(config.Lists[args[0]]))
	config.Mux.Unlock()

	return nil
}

func rpushCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)

	config.Mux.Lock()
	config.Lists[args[0]] = append(config.Lists[args[0]], args[1:]...)
	WriteInteger(conn, len(config.Lists[args[0]]))
	config.Mux.Unlock()

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
	config.Storage[args[4]] = object{Value: args[6], ExpiresAt: expiresAt}
	config.Mux.Unlock()

	WriteSimpleString(conn, "OK")
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
