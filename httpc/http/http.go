package http

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type HTTP struct {
	operation string
	verbose   bool
	headers   map[string]string
	body      string
	file      bool
	inline    bool
	conn      net.Conn
	request   []string
	response  []string
}

func (h HTTP) Connect() error {
	defer h.conn.Close()

	buf := make([]byte, 1024)
	//stdin := bufio.NewReader(os.Stdin)

	for i := 0; i < len(h.request); i++ {
		line := []byte(h.request[i])
		_, err := h.conn.Write(line)
		if err != nil {
			return err
		}
		if h.verbose {
			fmt.Println(h.request[i])
		}
	}
	wait := true
	omit := true
	for wait {
		if _, err := io.ReadFull(h.conn, buf); err != nil {
			if err == io.EOF {
				wait = false
			} else if err == io.ErrUnexpectedEOF {
				wait = false
			} else {
				return err
			}
		}
		fmt.Printf("Output:\n")
		if h.verbose {
			h.response = append(h.response, string(buf))
			os.Stdout.Write(buf)
		} else if omit {
			h.response = append(h.response, string(buf))
			h.findNewLines(buf)
		} else {
			//data for verbose has been removed
			h.response = append(h.response, string(buf))
			os.Stdout.Write(buf)
		}
	}
	return nil
}

func (h HTTP) findNewLines(in []byte) {
	input := strings.Split(string(in[:]), "\n")
	print := false
	for i := 0; i < len(input); i++ {
		if print {
			buf := ([]byte)(input[i] + "\n")
			os.Stdout.Write(buf)
		}
		if len(input[i]) == 1 {
			print = true
		}
	}
}
