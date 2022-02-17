package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	port := flag.Int("port", 8007, "echo server port")
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen on %d\n", *port)
		return
	}
	defer listener.Close()

	fmt.Println("echo server is listening on", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error occured during accept connection %v\n", err)
			continue
		}
		go handleConn(conn)
	}
}

//echo reads data and sends back what it received until the channel is closed
func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("new connection from %v\n", conn.RemoteAddr())

	//we can use io.Copy(conn, conn) but this function demonstrates read&write methods
	buf := make([]byte, 1024)
	for {
		n, re := conn.Read(buf)
		if n > 0 {
			if _, we := conn.Write(buf[:n]); we != nil {
				fmt.Fprintf(os.Stderr, "write error %v\n", we)
				break
			}
		}
		if re == io.EOF {
			break
		}
		if re != nil {
			fmt.Fprintf(os.Stderr, "read error %v\n", re)
			break
		}
	}
}
