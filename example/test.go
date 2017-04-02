package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"time"
)

func main() {
	fmt.Println("Start")
	server1, _ := zmq.NewSocket(zmq.ROUTER)
	//server2, _ := zmq.NewSocket(zmq.ROUTER)
	client1, _ := zmq.NewSocket(zmq.DEALER)
	client2, _ := zmq.NewSocket(zmq.DEALER)
	server1.Bind("tcp://*:8080")
	//server2.Bind("tcp://*:8082")
	client1.Connect("tcp://127.0.0.1:8080")
	client2.Connect("tcp://127.0.0.1:8080")
	go func() {
		for {
			msg, _ := client1.Recv(0)
			if msg != "" {
				fmt.Println("111111111", msg)
			}
		}
	}()
	go func() {
		for {
			msg, _ := client2.Recv(0)
			if msg != "" {
				fmt.Println("2222222222", msg)
			}
		}
	}()

	go func() {
		times := 0
		for {
			time.Sleep(time.Second)
			times++
			server1.Send(fmt.Sprintf("Hello [%d] ", times), 0)
			// if times == 3 {
			// 	server1.Connect("tcp://127.0.0.1:8082")
			// }
			// if times == 6 {
			// 	server1.Disconnect("tcp://127.0.0.1:8082")
			// }
		}
	}()
	// go func() {
	// 	msg, _ := server2.Recv(0)
	// 	if msg != "" {
	// 		fmt.Println("Server1:", msg)
	// 	}
	// 	//server2.Send(msg, 0)
	// }()
	for {
		msg, _ := server1.Recv(0)
		if msg != "" {
			fmt.Println("Client:", msg)
			server1.Send(msg, 0)
		}
	}
}
