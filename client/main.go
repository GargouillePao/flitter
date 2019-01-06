package main

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter"
	msgids "github.com/gargous/flitter/share/proto"
	"time"
)

func main() {
	d := core.NewDealer()
	err := d.Connect("localhost:8091")
	if err != nil {
		fmt.Println(err)
	}
	prcser := core.NewMsgProcesser()
	prcser.Rejister(msgids.PID_LOGIN_ACK, msgids.MsgCreators[msgids.PID_LOGIN_ACK],
		func(d core.Dealer, msg interface{}) error {
			pack, ok := msg.(*msgids.LoginAck)
			if !ok {
				return errors.New("Assetion Error")
			}
			fmt.Println("Server Say", pack)
			return nil
		})
	go func() {
		for i := 0; i < 10; i++ {
			msg := &msgids.LoginReq{
				Udid:    uint64(i + 1),
				Account: "xxxx",
			}
			err = d.Send(msgids.PID_LOGIN_REQ, msg)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Send", msg)
			time.Sleep(time.Second)
		}
	}()
	err = d.Process(prcser)
	if err != nil {
		fmt.Println(err)
	}
}
