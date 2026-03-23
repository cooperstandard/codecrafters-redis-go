package parser

import (
	"cmp"
	"fmt"
	"net"
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

func WriteEmptyArray(conn net.Conn) {
	fmt.Fprintf(conn, "*-1\r\n")
}

func WriteSimpleError(conn net.Conn, msg string) {
	fmt.Fprintf(conn, "-ERR %s\r\n", msg)
}
