package main

import (
	"bufio"
	"comp445/udp"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	//"comp445/httpc/http"
)

func handleConnection(conn net.Conn, path string, verbose bool) {
	defer conn.Close()
	//read in data from the connection in a loop, using boolean values to determine what to do
	opset := false
	opGET := false
	opPOST := false
	filePath := ""
	opLIST := false
	//headers have been processed
	headers := true
	//data for file system
	fileData := ""
	//start loop
	connection := bufio.NewReader(conn)
	//out of loop
	readData := ""
	//var err error

	for status, err := connection.ReadString('\n'); err == nil && status != "\n" && status != "\r"; status, err = connection.ReadString('\n') {
		if connection.Buffered() < 4 {
			break
		}
		fmt.Println(strings.ReplaceAll(status, "\r", ";"))
		fmt.Println(err)
		fmt.Println(status)
		fmt.Println(connection.Buffered())
		//do stuff
		if opset {
			//check if we have been processing the headers
			if status == "\r" {
				//we are exiting a section
				if headers {
					headers = false
				}
				continue
			}
			//we are processing headers if this is true
			if headers {
				//TODO: Process headers
			} else {
				//append file data
				fileData = fileData + status
			}
			//do stuff
		} else {
			//its the start line, reset flags needed
			opset = false
			opGET = false
			opPOST = false
			opLIST = false
			pathSET := false
			//read data
			tmp := strings.Split(status, " ")
			for i := 0; i < len(tmp); i++ {
				if tmp[i] == "GET" {
					opGET = true
					opset = true
					continue
				}
				if tmp[i] == "POST" {
					opPOST = true
					opset = true
					continue
				}
				//error check
				if opPOST && opGET {
					panic("Incorrect Format")
				}
				// check if the value is the endline if so then exit the loop
				if tmp[i] == "HTTP/1.0\r" {
					break
				}
				//check if its root
				if tmp[i] == "/" {
					opLIST = true
					opGET = false
					opset = true
					continue
				}
				//finally one can assume that this is a path to be used (omit the first /)
				if len(tmp[i]) > 1 {
					if strings.Contains(tmp[i], "/") && !pathSET {
						filePath = tmp[i][1:]
						pathSET = true
					}
				}
			}
			if !opset {
				panic("Incorrect Format")
			}
		}
		if verbose {
			fmt.Println(">" + status + "<")
		}
	}
	var buf []byte
	flush, err := io.ReadFull(connection, buf)
	fmt.Println(flush)
	readData = readData + string(buf)
	//we execute the functions generated
	fmt.Println("Rrequest Reading Complete")
	fmt.Print("opset: ")
	fmt.Println(opset)
	fmt.Print("opGET: ")
	fmt.Println(opGET)
	fmt.Print("opPOST: ")
	fmt.Println(opPOST)
	fmt.Print("opLIST: ")
	fmt.Println(opLIST)
	if opset {
		//clean the path (***SECURITY MECHANISM***)
		filePath = path + "/" + strings.ReplaceAll(filePath, "/", "")
		fmt.Println(filePath)
		if opLIST {
			//list files in working directory
			fmt.Println("OPLIST")
			info, erra := os.ReadDir(path)
			fmt.Println(erra)
			for i := 0; i < len(info); i++ {
				fmt.Println(info[i].Name())
				readData = readData + info[i].Name()
			}

		} else if opGET {
			fmt.Println("OPGET")
			//read from a file found
			tmpData, errb := ioutil.ReadFile(filePath)
			if errb == nil {
				fmt.Println(string(readData))
				readData = string(tmpData)
			} else {
				fmt.Println(errb)
			}
		} else if opPOST {
			fmt.Println("OPPOST")
			//write to a file found
			msg := []byte(fileData)
			err := ioutil.WriteFile(filePath, msg, 0644)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			panic("THERE WAS NO CORRECT OPERATION SET!")
		}
	} else {
		panic("THERE WAS NO CORRECT OPERATION SET!")
	}
	resp := "404"
	s := []byte(resp + readData)
	//http.QuickConnect(verbose, s, conn)
	dataList, err := conn.Write(s)
	if err == nil {
		fmt.Println(dataList)
	} else {
		fmt.Println(err)
	}
	//the connection should close
}

func main() {
	//for the directory default
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(ex)
	verbose := false
	port := "8080"
	//start processing the arguments
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v":
			verbose = true
		case "-p":
			i++
			if i == len(args) {
				panic("Out of Bounds!")
			}
			port = args[i]
		case "-d":
			i++
			if i == len(args) {
				panic("Out of Bounds!")
			}
			path = args[i]
		}
	}
	//all values set
	fmt.Println("Running on :" + port)
	portNumber, _ := strconv.Atoi(port)
	udp := udp.Server("udp", portNumber, 10000)
	for {
		conn, err := udp.Handshake()
		if err != nil {
			if verbose {
				fmt.Println(err)
			}
			break
		}
		go handleConnection(conn, path, verbose)
	}
}
