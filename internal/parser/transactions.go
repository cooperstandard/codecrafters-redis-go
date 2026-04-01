package parser

import (
	"fmt"
	"net"
)


func multiCommand(args []string, conn net.Conn, config Config) []byte {
	args = GetArgs(args)	
	fmt.Println(args)
	return GetSimpleString("OK")
}

func execCommand(args []string, conn net.Conn, config Config) []byte {
	//TODO: return the byte slice

	WriteSimpleError(conn, "EXEC without MULTI")
	return nil
}
