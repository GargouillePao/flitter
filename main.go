package main

import (
	"bufio"
	"fmt"
	"github.com/tidwall/evio"
	"net"
)

func doServer(errChan chan error) {
	var events evio.Events
	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		fmt.Println("Client Say", string(in))
		out = in
		return
	}
	if err := evio.Serve(events, "udp://localhost:8080"); err != nil {
		errChan <- err
		return
	}
}

func doClient(errChan chan error) {
	conn, err := net.Dial("udp", "localhost:8080")
	if err != nil {
		errChan <- err
		return
	}
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	var buf [1024]byte
	go func() {
		for i := 0; i < 10; i++ {
			send := "Hello"
			_, err = w.Write([]byte(send))
			if err != nil {
				errChan <- err
				return
			}
			fmt.Println("Send", send)
		}
	}()
	go func() {
		for i := 0; i < 10; i++ {
			n, err := r.Read(buf[:])
			if err != nil {
				errChan <- err
				return
			}
			fmt.Println("Recv", string(buf[:n]))
		}
	}()
}

func main() {
	errChan := make(chan error, 0)
	go doServer(errChan)
	go doClient(errChan)
	err := <-errChan
	fmt.Println(err)
}
