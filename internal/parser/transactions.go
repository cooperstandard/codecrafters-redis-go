package parser

import (
	"fmt"
	"net"
)


func multiCommand(args []string, conn net.Conn, config Config) error {
	args = GetArgs(args)	
	fmt.Println(args)
	WriteSimpleString(conn, "OK")


	return nil
}

func execCommand(args []string, conn net.Conn, config Config) error {

	WriteSimpleError(conn, "EXEC without MULTI")
	return nil
}
