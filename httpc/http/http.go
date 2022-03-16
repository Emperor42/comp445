package http

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"io/ioutil"
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


func QuickConnect(verbose bool, request []string, conn net.Conn) error {
	defer conn.Close()

	buf := make([]byte, 1024)
	//stdin := bufio.NewReader(os.Stdin)

	for i := 0; i < len(request); i++ {
		line := []byte(request[i])
		_, err := conn.Write(line)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Println(request[i])
		}
	}
	wait := true
	omit := true
	for wait {
		if _, err := io.ReadFull(conn, buf); err != nil {
			if err == io.EOF {
				wait = false
			} else if err == io.ErrUnexpectedEOF {
				wait = false
			} else {
				return err
			}
		}
		fmt.Printf("Output:\n")
		if verbose {
			os.Stdout.Write(buf)
		} else if omit {
			findNewLines(buf)
		} else {
			//data for verbose has been removed
			os.Stdout.Write(buf)
		}
	}
	return nil
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


// usage: go run httpc.go [--host hostname] [--port port number]
//httpc (get|post) [-v] (-h "k:v")* [-d inline-data] [-f file] URL

func Run(arguments string) {
	args := strings.Split(arguments, " ")[1:]
	operation := args[0]
	//always store the headers
	headers := make(map[string]string)

	if processOperation(operation) {
		printHelp(args)
		return
	}
	//reset args to limit data being used
	args = args[1:]
	verbose := processVerbose(args)
	//remove the verbose flag if needed
	if verbose {
		args = removeFlag("-v", args)
	}
	headers, args = processHeaders(verbose, headers, args)

	//////
	//START PROCESS INLINE DATA AND FILES
	/////
	hasFile, hasInlineData, body := processFilesOrData(args)
	//remove proces data

	if operation == "get" {
		if hasFile {
			if verbose {
				fmt.Println("GET CAN'T USE -f FLAG!")
			}
			return
		}
		if hasInlineData {
			if verbose {
				fmt.Println("GET CAN'T USE -d FLAG!")
			}
			return
		}
	}

	if operation == "post" {
		if hasFile && hasInlineData {
			if verbose {
				fmt.Println("POST CAN'T USE -f ALONG WITH -d FLAG!")
			}
			return
		}
	}

	//////
	//START PROCESS URL
	/////

	url, full := findHTTPURL(args)
	port := 80

	/////
	//START BUILD REQUEST
	/////
	request := make([]string, 0)
	//append the operation
	startLine := strings.ToUpper(operation) + " " + full + " HTTP/1.0\r\n\r\n"
	request = append(request, startLine)

	if hasFile || hasInlineData {
		//include the headers for data
		headers["Content-Length"] = fmt.Sprint(len(body))
		if hasFile {
			headers["files"] = body
		}
		if hasInlineData {
			headers["data"] = body
		}
	}

	for k, v := range headers {
		headerLine := k + ": " + v+"\n"
		request = append(request, headerLine)
	}
	if hasFile || hasInlineData {
		//include the headers for data
		request = append(request, "\r\n\r\n"+body+"\r\n\r\n")
	}

	if verbose {
		fmt.Println("Attemping connection...")
	}

	addr := fmt.Sprintf("%s:%d", url, port)

	conn, err := net.Dial("tcp", addr)

	if err != nil {
		if verbose {
			fmt.Println(err)
		}
		fmt.Fprintf(os.Stderr, "failed to connect to %s\n", addr)
		return
	}
	if err := connect(verbose, request, conn); err != nil {
		if verbose {
			fmt.Println(err)
		}
		fmt.Fprintf(os.Stderr, "Error during Respond %v\n", err)
	}

}

func connect(verbose bool, request []string, conn net.Conn) error {
	defer conn.Close()

	buf := make([]byte, 1024)
	//stdin := bufio.NewReader(os.Stdin)

	for i := 0; i < len(request); i++ {
		line := []byte(request[i])
		_, err := conn.Write(line)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Println(request[i])
		}
	}
	wait := true
	omit := true
	for wait {
		if _, err := io.ReadFull(conn, buf); err != nil {
			if err == io.EOF {
				wait = false
			} else if err == io.ErrUnexpectedEOF {
				wait = false
			} else {
				return err
			}
		}
		fmt.Printf("Output:\n")
		if verbose {
			os.Stdout.Write(buf)
		} else if omit {
			findNewLines(buf)
		} else {
			//data for verbose has been removed
			os.Stdout.Write(buf)
		}
	}
	return nil
}

func findNewLines(in []byte) {
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

func findHTTPURL(input []string) (string, string) {
	for i := 0; i < len(input); i++ {
		if strings.Contains(input[i], "http") {
			if input[i][0] != 'h' {
				coreURL := input[i][(strings.Index(input[i], "://") + 3):(len(input[i]) - 1)]
				fullURL := input[i][1:(len(input[i]) - 1)]
				index := strings.Index(coreURL, "/")
				if index != -1 {
					return coreURL[:index], fullURL
				}
				return coreURL, fullURL
			} else {
				coreURL := input[i][(strings.Index(input[i], "://") + 3):]
				fullURL := input[i]
				index := strings.Index(coreURL, "/")
				if index != -1 {
					return coreURL[:index], fullURL
				}
				return coreURL, fullURL
			}
		}

	}
	return "", ""
}

func processFilesOrData(input []string) (bool, bool, string) {
	result := make([]string, 0)
	fFlag := false
	dFlag := false
	output := ""
	for i := 0; i < len(input); i++ {
		//check if its a header
		if input[i] == "-f" {
			fFlag = true
			//start reading out the files, checking if it starts with a single or double quote
			checking := input[i+1]
			endOfData := 'C'
			if checking[0] == '\'' {
				endOfData = '\''
			} else if checking[0] == '"' {
				endOfData = '"'
			}
			if endOfData != 'C' {
				raw := input[i+1][1:]
				for x := i + 2; x < len(input); x++ {
					if rune(input[x][len(input[x])-1]) == endOfData {
						raw = raw + input[x][:len(input[x])-1]
						i = x + 1
						break
					} else {
						raw = raw + input[x]
					}
				}
				data, err := ioutil.ReadFile(raw)
				if err != nil {
					fmt.Println("File reading error", err)
				}
				output = string(data)

			} else {
				i = i + 1
				data, err := ioutil.ReadFile(input[i])
				if err != nil {
					fmt.Println("File reading error", err)
				}
				output = string(data)
			}

		} else if input[i] == "-d" {
			dFlag = true
			fmt.Println("i run")
			//start reading out the files, checking if it starts with a single or double quote
			checking := input[i+1]
			endOfData := 'C'
			if checking[0] == '\'' {
				endOfData = '\''
			} else if checking[0] == '"' {
				endOfData = '"'
			}
			if endOfData != 'C' {
				raw := input[i+1][1:]
				for x := i + 2; x < len(input); x++ {
					if rune(input[x][len(input[x])-1]) == endOfData {
						raw = raw + input[x][:len(input[x])-1]
						i = x + 1
						break
					} else {
						raw = raw + input[x]
					}
				}
				output = raw
			} else {
				i = i + 1
				output = input[i]
			}

		} else {
			result = append(result, input[i])
		}
	}
	return fFlag, dFlag, output
}

func processHeaders(verbose bool, currentHeaders map[string]string, input []string) (map[string]string, []string) {
	result := make([]string, 0)
	for i := 0; i < len(input); i++ {
		//check if its a header
		if input[i] == "-h" {
			//check if there is in fact something next
			if (i + 1) < len(input) {
				//set the headers to the desired values
				tmp := strings.Split(input[i+1], ":")
				currentHeaders[tmp[0]] = tmp[1]
				i++
			}
		} else {
			result = append(result, input[i])
		}
	}
	return currentHeaders, result
}

func printHelp(args []string) {
	//you are printing some help screen
	if len(args) == 2 {
		switch args[1] {
		case "get":
			fmt.Println("usage: httpc get [-v] [-h key:value] URL\nGet executes a HTTP GET request for a given URL.\n-v Prints the detail of the response such as protocol, status, and headers.\n-h key:value Associates headers to HTTP Request with the format\n'key:value'.")
			return

		case "post":
			fmt.Println("usage: httpc post [-v] [-h key:value] [-d inline-data] [-f file] URL\nPost executes a HTTP POST request for a given URL with inline data or fromfile.\n-v Prints the detail of the response such as protocol, status,and headers.\n-h key:value Associates headers to HTTP Request with the format'key:value'.\n-d string Associates an inline data to the body HTTP POST request.\n-f file Associates the content of a file to the body HTTP POST request.\nEither [-d] or [-f] can be used but not both.")
			return
		}
	}
	fmt.Println("httpc is a curl-like application but supports HTTP protocol only.\nUsage:\nhttpc command [arguments]\nThe commands are:\nget executes a HTTP GET request and prints the response.\npost executes a HTTP POST request and prints the response.\nhelp prints this screen.\nUse \"httpc help [command]\" for more information about a command.")
	return
}

func removeFlag(flag string, input []string) []string {
	result := make([]string, 0)
	for i := 0; i < len(input); i++ {
		if input[i] == flag {
			continue
		}
		result = append(result, input[i])
	}
	return result
}

func removeFlagAndData(flag string, input []string) []string {
	result := make([]string, 0)
	skip := false
	for i := 0; i < len(input); i++ {
		if skip {
			skip = false
			continue
		}
		if input[i] == flag {
			skip = true
			continue
		}
		result = append(result, input[i])
	}
	return result
}

func processOperation(operation string) bool {
	switch operation {
	case "help":
		return true
	case "get":
		return false
	case "post":
		return false
	default:
		fmt.Println(operation + " is not a valid operation, use httpc help for more information")
		return false
	}
}

func processVerbose(input []string) bool {
	//looking for verbose flag
	for i := 0; i < len(input); i++ {
		if input[i] == "-v" {
			return true
		}
	}
	return false
}

func processURL(url string) bool {
	return false
}
