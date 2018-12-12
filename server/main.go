package main

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter/core"
	msgids "github.com/gargous/flitter/share/proto"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/evio"
)

var accounts map[uint64]msgids.Account

func initFakeAccount() {
	accounts = make(map[uint64]msgids.Account)
}

func main() {
	var events evio.Events
	processer := core.NewMsgProcesser()
	processer.Rejister("PID_LOGIN_REQ", func() proto.Message {
		return &msgids.LoginReq{}
	}, func(msg interface{}) error {
		pack, ok := msg.(*msgids.LoginReq)
		if !ok {
			return errors.New("Assetion Error")
		}
		fmt.Println("Client Say", pack)
		return nil
	})
	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		fmt.Println(c)
		err := processer.Process(in)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	fmt.Println("Server Start")
	if err := evio.Serve(events, "udp://localhost:8080", "tcp://localhost:8081"); err != nil {
		fmt.Println(err)
		return
	}
}
