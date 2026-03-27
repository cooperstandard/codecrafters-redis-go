package parser

import (
	"cmp"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func Last[S ~[]E, E cmp.Ordered](x S) E {
	if len(x) < 1 {
		panic("Last: empty list")
	}
	return x[len(x)-1]
}

func WriteSimpleString(conn net.Conn, val string) {
	fmt.Fprintf(conn, "+%s\r\n", val)
}

func GetBulkString(val string) string {
	if len(val) == 0 {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
}

// StreamIDCompare takes in 2 ids and returns 0 if they are equal, -1 if id1 < id2, or 1 if id1 > id2
func StreamIDCompare(id1, id2 string) int {
	if id1 == "+" || id2 == "-" {
		return 1
	}

	if id2 == "+" || id1 == "-" {
		return -1
	}

	timestamp1, _ := strconv.Atoi(strings.Split(id1, "-")[0])
	timestamp2, _ := strconv.Atoi(strings.Split(id2, "-")[0])

	if timestamp1 < timestamp2 {
		return -1
	} else if timestamp1 > timestamp2 {
		return 1
	}

	seq1, _ := strconv.Atoi(strings.Split(id1, "-")[1])
	seq2, _ := strconv.Atoi(strings.Split(id2, "-")[1])

	if seq1 < seq2 {
		return -1
	} else if seq1 > seq2 {
		return 1
	}

	return 0
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

func WriteStringArray(conn net.Conn, list []string) {
	fmt.Fprintf(conn, CreateStringArray(list))
}

func CreateStringArray(list []string) string {
	str := fmt.Sprintf("*%d\r\n", len(list))
	for _, v := range list {
		str += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	}
	return str
}

func WriteStreamSlice(conn net.Conn, s []stream) {
	streamString := fmt.Sprintf("*%d\r\n", len(s))
	for _, v := range s {
		streamString += CreateArrayFromStream(v)
	}

	fmt.Fprint(conn, streamString)
}

func CreateArrayFromStream(s stream) string {
	ret := fmt.Sprintf("*2\r\n%s", GetBulkString(s.ID))
	vals := []string{}
	for k, v := range s.data {
		vals = append(vals, k)
		vals = append(vals, v)
	}

	ret += CreateStringArray(vals)

	return ret
}

func WriteEmptyArray(conn net.Conn) {
	fmt.Fprintf(conn, "*-1\r\n")
}

func WriteSimpleError(conn net.Conn, msg string) {
	fmt.Fprintf(conn, "-ERR %s\r\n", msg)
}

func validateAndGenerateID(conn net.Conn, config Config, id string, streamName string) (string, bool) {
	config.Mux.RLock()
	defer config.Mux.RUnlock()
	if strings.Compare(id, "0-0") <= 0 && id[len(id)-1] != '*' {
		errorMessage := "The ID specified in XADD must be greater than 0-0"
		WriteSimpleError(conn, errorMessage)
		return "", false
	}
	// this is wrong, I should be searching through the whole list for the latest (if exists) entry with a matching timestamp
	if s, exists := config.Streams[streamName]; exists {
		errorMessage := "The ID specified in XADD is equal or smaller than the target stream top item"
		if len(s) != 0 {
			if strings.Compare(id, s[len(s)-1].ID) <= 0 && id[len(id)-1] != '*' {
				WriteSimpleError(conn, errorMessage)
				return "", false
			}
		}
	}

	parts := strings.Split(id, "-")
	now := strconv.Itoa(int(time.Now().UnixMilli()))
	if parts[0] == "*" {
		parts[0] = now
		parts = append(parts, "")
	} else if parts[1] != "*" {
		return id, true
	}
	if s, exists := config.Streams[streamName]; exists {
		if strings.Split(s[len(s)-1].ID, "-")[0] == parts[0] {
			ordinal, _ := strconv.Atoi(strings.Split(s[len(s)-1].ID, "-")[1])
			ordinal += 1
			parts[1] = strconv.Itoa(ordinal)
		} else {
			parts[1] = "0"
			if parts[0] == "0" {
				parts[1] = "1"
			}
		}
	} else {
		parts[1] = "0"
		if parts[0] == "0" {
			parts[1] = "1"
		}
	}

	return fmt.Sprintf("%s-%s", parts[0], parts[1]), true
}
