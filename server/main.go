package main

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter/core"
	msgids "github.com/gargous/flitter/share/proto"
	"time"
)

var accounts map[uint64]msgids.Account

func initFakeAccount() {
	accounts = make(map[uint64]msgids.Account)
}

func main() {
	prcser := core.NewMsgProcesser()
	prcser.Rejister(msgids.PID_LOGIN_REQ, msgids.MsgCreators[msgids.PID_LOGIN_REQ],
		func(d core.Dealer, msg interface{}) error {
			pack, ok := msg.(*msgids.LoginReq)
			if !ok {
				return errors.New("Assetion Error")
			}
			fmt.Println("Client Say", pack)
			roles := make([]*(msgids.Role), 5)
			for i := 0; i < 5; i++ {
				roles[i] = &msgids.Role{
					Id:   uint64(i),
					Name: "YYY",
				}
			}
			ack := &msgids.LoginAck{
				Udid:  0,
				Roles: roles,
			}
			d.Send(msgids.PID_LOGIN_ACK, ack)
			return nil
		})
	srv := core.NewServer(prcser, 100000, time.Minute*5)
	srv.Serve("0.0.0.0:8091")
}
