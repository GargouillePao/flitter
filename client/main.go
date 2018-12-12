package main

import (
	"bufio"
	"fmt"
	"github.com/gargous/flitter/core"
	msgids "github.com/gargous/flitter/share/proto"
	"github.com/golang/protobuf/proto"
	"net"
)

func doClient(errChan chan error) {
	conn, err := net.Dial("udp", "localhost:8080")
	if err != nil {
		errChan <- err
		return
	}
	processer := core.NewMsgProcesser()
	processer.Rejister("PID_LOGIN_REQ", func() proto.Message {
		return &msgids.LoginReq{}
	}, nil)
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	var buf [1024]byte
	go func() {
		for i := 0; i < 10; i++ {
			msg := &msgids.LoginReq{
				Udid:    uint64(i),
				Account: "xxxx",
			}
			id, _ := processer.GetHeadId("PID_LOGIN_REQ")
			data, _ := processer.Encode(id, msg)
			_, err = w.Write(data)
			if err != nil {
				errChan <- err
				return
			}
			w.Flush()
			fmt.Println("Send", msg)
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
	go doClient(errChan)
	err := <-errChan
	fmt.Println(err)
}
