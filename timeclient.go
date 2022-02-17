package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	host := flag.String("host", "localhost", "time server hostname")
	port := flag.Int("port", 8037, "time server port")
	flag.Parse()
	addr := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s\n", addr)
		return
	}

	if err := printTime(conn); err != nil {
		fmt.Fprintf(os.Stderr, "failed to retrieve time %v\n", err)
		return
	}
}

func printTime(conn net.Conn) error {
	// This value is the elapsed seconds from 1/1/1990 to 1/1/1970
	var time1970 uint32 = 2208988800
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	seconds := binary.BigEndian.Uint32(buf) - time1970
	rtime := time.Unix(int64(seconds), 0)
	fmt.Printf("%v\n", rtime)
	return nil
}
