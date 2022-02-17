package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

// usage: go run echoclient.go [--host hostname] [--port port number]
func main() {
	host := flag.String("host", "localhost", "echo server hostname")
	port := flag.Int("port", 8007, "echo server port")
	flag.Parse()
	addr := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s\n", addr)
		return
	}
	fmt.Println("Type any thing then ENTER. Press Ctrl+C to terminate")
	if err := repl(conn); err != nil {
		fmt.Fprintf(os.Stderr, "Error during repl %v\n", err)
	}
}

func repl(conn net.Conn) error {
	defer conn.Close()

	buf := make([]byte, 1024)
	stdin := bufio.NewReader(os.Stdin)

	for {
		line, err := stdin.ReadSlice('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		n, err := conn.Write(line)
		if err != nil {
			return err
		}
		if cap(buf) < n {
			buf = make([]byte, n)
		}
		if _, err := io.ReadFull(conn, buf[:n]); err != nil {
			return err
		}
		fmt.Printf("Replied: ")
		os.Stdout.Write(buf[:n])
	}
	return nil
}
