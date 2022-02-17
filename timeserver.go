package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	port := flag.Int("port", 8037, "time server port")
	flag.Parse()
	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen %s with %v\n", addr, err)
		return
	}

	defer listener.Close()
	fmt.Println("Time server is listening at", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection %v\n", err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			if err := reportTime(conn); err != nil {
				fmt.Fprintf(os.Stderr, "failed to report time %v\n", err)
			}
		}(conn)
	}
}

func reportTime(conn net.Conn) error {
	// Number of seconds elapsed from 1900 to 1970
	var time1970 int64 = 2208988800
	now := time.Now().Unix() + time1970
	buf := make([]byte, 4)

	// Must send the uint32 in big-endian over the network
	binary.BigEndian.PutUint32(buf, uint32(now))
	_, err := conn.Write(buf)
	return err
}
