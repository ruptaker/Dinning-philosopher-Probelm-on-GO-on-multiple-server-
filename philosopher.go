package main

import (
	"fmt"
	"os"
	"net"
	"math/rand"
	"time"
	"strings"
	"strconv"
)

var done int
var conn *net.UDPConn

//Check for potential error in connection
func checkError(err error) {
    if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]

	udpAddr, err := net.ResolveUDPAddr("udp", service)
	checkError(err)

	conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err)

	_, err = conn.Write([]byte("Philosopher"))
	checkError(err)
	var buf [512]byte
	n, err := conn.Read(buf[0:])
	checkError(err)

	id, _ := strconv.Atoi(string(buf[0:n]))

	if id == -1 {
		fmt.Fprintf(os.Stderr, "Maximum limit reached. Rejected from server\n")
		os.Exit(1)
	}

	fmt.Println("Philosoper ", strconv.Itoa(id)+" joined.")

	fmt.Println("Waiting for Server...")

	n, err = conn.Read(buf[0:])
	checkError(err)

	fmt.Println("Starting...")

	var send = make(chan string)
	go handleMsgsent(send)
	var msg_receive = make(chan string)
	go receiveChannel(msg_receive)

	handlephilosopher(id, send, msg_receive)

	os.Exit(0)
}




//Handling philosopher here. 

func handlephilosopher(id int, send chan string, msg_receive chan string) {

	running_time := 90 
	current_status := 0
	eat_time := 0
	next_request := 0

	_ = next_request

	current_status = 1		
	send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)
	var buf [512]byte
	for running_time > 0 {

		send <- "request," + strconv.Itoa(id)

		buffer := <-msg_receive
		message := strings.Split(buffer, ",")

		if message[0] == "reject" {

			next_request = rand.Intn(running_time) + 1
			//Sleep for certain time and try again
			time.Sleep(time.Duration(next_request) * time.Second)
			continue
		}
		//Trying to get left fork
		tcpAddr, err := net.ResolveTCPAddr("tcp4", message[1])
		//Error means fork not available so reuest for releasing
		if err != nil {				
			send <- "release," + strconv.Itoa(id)

			next_request = rand.Intn(4) + 1
			send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)

			time.Sleep(time.Duration(next_request) * time.Second)
			continue
		}
		fork1_conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			send <- "release," + strconv.Itoa(id)

			next_request = rand.Intn(4) + 1
			send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)

			time.Sleep(time.Duration(next_request) * time.Second)
			continue
		}
		//trying to get right fork
		tcpAddr, err = net.ResolveTCPAddr("tcp4", message[2])
		if err != nil {
			send <- "release," + strconv.Itoa(id)

			next_request = rand.Intn(4) + 1
			send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)

			time.Sleep(time.Duration(next_request) * time.Second)
			continue
		}
		fork2_conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			send <- "release," + strconv.Itoa(id)

			next_request = rand.Intn(4) + 1
			send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)

			time.Sleep(time.Duration(next_request) * time.Second)
			continue
		}
		//No error
		current_status = 2		

		send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)

		eat_time = rand.Intn(running_time)

		if eat_time == 0 {
			eat_time = 1
		}
		i := 0
		for i < eat_time {
			//Signal the left fork process of eating
			fork1_conn.Write([]byte("Start Eating"))
			n, err := fork1_conn.Read(buf[0:])
			_ = err

			check := string(buf[0:n])
			//Signal the left fork process of eating
			fork2_conn.Write([]byte("Start Eating"))
			n, err = fork2_conn.Read(buf[0:])
			_ = err

			if check == "eating" && string(buf[0:n]) == "eating" {
				i = i + 1
			}
		}
		//Finish eating
		fork1_conn.Write([]byte("Stop Eating"))
		fork2_conn.Write([]byte("Stop Eating"))

		running_time = running_time - eat_time

		send <- "release," + strconv.Itoa(id)

		current_status = 1 //Waiting again
		send <- "print," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)

		if running_time >= 1 {

			next_request = rand.Intn(running_time) + 1
			time.Sleep(time.Duration(next_request) * time.Second) //Thinking state
		}

	}
	current_status = 3
	send <- "done," + strconv.Itoa(id) + "," + strconv.Itoa(current_status)
	close(msg_receive)
	for done == 0 {
		next_request = rand.Intn(4) + 1
		time.Sleep(time.Duration(next_request) * time.Second)
	}
	close(send)
}

//Handle message sent to Main server 
func handleMsgsent(send chan string) {
	for {
		buffer := <-send

		_, err := conn.Write([]byte(buffer))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
			os.Exit(1)
		}

		if strings.Split(buffer, ",")[0] == "done" {
			done = 1
		}
	}

}

//Handle message receive from Main Server
func receiveChannel(msg_receive chan string) {
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}

		msg_receive <- string(buf[0:n])
	}

}