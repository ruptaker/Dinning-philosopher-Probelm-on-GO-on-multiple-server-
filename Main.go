package main

import (
	"fmt"
	"os"
	"time"
	"net"
	"strings"
	"strconv"
)

//Variable declaration
var conn *net.UDPConn
var process_address [10]*net.UDPAddr
var num_philosophers int
var fork_locking [5]int
var fork_address [5]string
var done int
var num_forks int

//Check for potential error in connection
func checkError(err error) {
    if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

//Starting program 
func main() {
	fmt.Println("\n\nDining Philosophers Problem")
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s port\n", os.Args[0])
		os.Exit(1)
	}
	service := ":" + os.Args[1]
	//Establish UDP connection
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	checkError(err)

	conn, err = net.ListenUDP("udp", udpAddr)
	checkError(err)

	fmt.Println("Server is running on the address: " + service) //server is ready 

	num_forks = 0
	num_philosophers = 0
	
	for {
			
			var buf [512]byte

			n, addr, err := conn.ReadFromUDP(buf[0:])
			if err != nil {
				return
			}
			//Create array of size =5philosopher and fork process addresses
			if string(buf[0:n]) == string("Philosopher") && (num_philosophers < 5) {
				process_address[num_philosophers] = addr
				//Send philosopher id to client process
				conn.WriteToUDP([]byte(strconv.Itoa(num_philosophers)), addr)
				num_philosophers += 1

			} else if strings.Split(string(buf[0:n]), ",")[0] == "Fork" && num_forks < 5 {
				process_address[num_forks+5] = addr
				fork_address[num_forks] = strings.Split(string(buf[0:n]), ",")[1]
				//Send fork id to client process
				conn.WriteToUDP([]byte(strconv.Itoa(num_forks)), addr)
				num_forks += 1

			} else {
				conn.WriteToUDP([]byte("-1"), addr)
			}
			//Stop creating fork and philosopher when each number is 5	
			if num_philosophers == 5 && num_forks == 5 {
					break
			}
		}

	var i int

	for i = 0; i < 10; i++ {
		conn.WriteToUDP([]byte("Start"), process_address[i])
	}

	for i = 0; i < 5; i++ {
	 fork_locking[i] = 0
	}
	fmt.Println("Current time\t\tphilo #0\t\tphilo #1\t\tphilo #2\t\tphilo #3\t\tphilo #4\t\t")
	go display("0", "0")




	var msg_receive = make(chan string)
	var msg_send = make(chan string)
	done = 0
	



	go UDPMsgReceive(msg_receive, msg_send)
	go UDPMsgSend(msg_send, conn)

	var buf [512]byte





	for {

		n, _, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			continue
		}
		msg_receive <- string(buf[0:n])
	}
}




func forkRequest(identifier string, msg_send chan string) {         //Check for the neighbouring fork i.e handle fork request
	id, err := strconv.Atoi(identifier)
	_ = err

	fork_left := id
	fork_right := (id + 1) % 5

	if fork_locking[fork_left] == 0 && fork_locking[fork_right] == 0 {
		 fork_locking[fork_left] = 1
		 fork_locking[fork_right] = 1
		 msg_send <- identifier + "-accept," + fork_address[fork_left] + "," + fork_address[fork_right]

	} else {
		msg_send <- identifier + "-reject"
	}
}

//Release fork function i.e. handle fork release
func forkRelease(identifier string, send chan string) {
	id, err := strconv.Atoi(identifier)
	_ = err

	fork_left := id
	fork_right := (id + 1) % 5

	fork_locking[fork_left] = 0
	fork_locking[fork_right] = 0
}


func UDPMsgSend(send chan string, conn *net.UDPConn) {                      //Handle message sent to philosopher process
	for {
		buffer := <-send
		id, err := strconv.Atoi(strings.Split(buffer, "-")[0])
		message := strings.Split(buffer, "-")[1]

		_ = err

		conn.WriteToUDP([]byte(message), process_address[id])
	}
}



func UDPMsgReceive(msg_receive chan string, send chan string) {          
	for {
		buffer := <-msg_receive
		message := strings.Split(buffer, ",")

		msg:= message[0] 
		if strings.Compare("print",msg) == 0{
			display(message[1], message[2])
		}else if strings.Compare("request",msg) == 0{
			forkRequest(message[1], send)
		}else if strings.Compare("release",msg) == 0{
			forkRelease(message[1], send)
		}else if strings.Compare("done",msg) == 0{
			done = done + 1
			display(message[1], message[2])
			if done == 5 {
				fmt.Println("\n\n------Dining Philosophers Ends Here-----")
				conn.Close()
			}
		}		
		
	}
}

  






// Status of philosopher
func display(identifier string, status_string string) {
	currentTime := time.Now().Format("Jan 01 15:04:05")
	print_string := currentTime + "\t\t"

	id, err := strconv.Atoi(identifier)
	checkError(err)
	flag, err := strconv.Atoi(status_string)
	checkError(err)

	if flag != 0 {
		for i := 0; i < 5; i++ {
			if id == i {
				switch flag {
				case 0:
					print_string += "thinking\t\t"
				case 1:
					print_string += "hungry\t\t"
				case 2:
					print_string += "eating \t\t"
				case 3:
					print_string += "done\t\t"
				}
			} else {
				print_string += "....\t\t"
			}
		}
	} else {
		for i := 0; i < 5; i++ {
			print_string += "thinking\t\t"
		}
	}

	fmt.Println(print_string)
}