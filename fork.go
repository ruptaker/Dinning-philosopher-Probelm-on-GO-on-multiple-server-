package main

import (
	"fmt"
	"net"
	"os"
	"time"
)


func checkError(err error) {                    //Check for potential error in connection
    if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr,"Enter server address and fork local address",os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s server_host:server_port fork_host:fork_port", os.Args[0])
		os.Exit(1)
	}
	mainHost := os.Args[1]
	forkServer := os.Args[2]

	udpAddr, err := net.ResolveUDPAddr("udp", mainHost)
	checkError(err)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err)

	_, err = conn.Write([]byte("Fork," + forkServer))
	checkError(err)
	var buf [512]byte
	n, err := conn.Read(buf[0:])
	checkError(err)

	id := string(buf[0:n])

	if id == "-1" {
		fmt.Fprintf(os.Stderr, "Maximum limit reached. Rejected from server\n")
		os.Exit(1)
	}

	fmt.Println("Fork: ", id+" joined.")

	fmt.Println("Waiting for Server...")

	n, err = conn.Read(buf[0:])
	checkError(err)

	fmt.Println("Running..")

	tcpAddr, err := net.ResolveTCPAddr("tcp", forkServer)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	fmt.Println("Fork listening on... ", id)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		send := make(chan string)
		//handlePhilosopher request
		go handlePhilosopher(send, conn)
	}

	os.Exit(0)
}


//Reply status to philosopher process
func ReplyTo(send chan string, conn net.Conn) {
	for {
			a := <-send
			if a == "0" {
				break
			}
			time.Sleep(time.Second)
			conn.Write([]byte("eating"))
	}
}

//Check if philosopher got fork 
func handlePhilosopher(send chan string, conn net.Conn) {
	go ReplyTo(send, conn)
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		_ = err
		fmt.Println(string(buf[0:n]))
		if string(buf[0:n])  == "Start Eating"{
			send <- "1"
		}else if string(buf[0:n]) ==  "Stop Eating"{
			send <- "0"
		}
	}
}

