package core

import (
	utils "github.com/gargous/flitter/utils"
	"testing"
)

func Test_Sender(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Sender"))
	sender, err := NewSender()
	if err != nil {
		t.Fatal(utils.Errf("sender err:%v", err))
	}
	sender.AddNodeInfo("/127.0.0.1:7000/127.0.0.1:8000")
	err = sender.Connect()
	if err != nil {
		t.Fatal(utils.Errf("connect err:%v", err))
	}
	msg := NewMessage(NewMessageInfo(), []byte("Hello"))
	err = sender.Send(msg)
	if err != nil {
		t.Fatal(utils.Errf("send err:%v", err))
	} else {
		t.Log(utils.Infof("send ok\nmsg:%v\nsender:%v", msg, sender))
	}
	t.Log(utils.Norf("End Sender"))
}
func Test_Receiver(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Receiver"))
	receiver, err := NewReceiver("*:8000")
	if err != nil {
		t.Fatal(utils.Errf("receiver err:%v,%v", err, receiver))
	}
	err = receiver.Bind()
	if err != nil {
		t.Fatal(utils.Errf("bind err:%v,%v", err, receiver))
	}
	msg, err := receiver.Recv()
	if err != nil {
		t.Fatal(utils.Errf("receive err:%v", err))
	} else {
		t.Log(utils.Infof("receive ok\nmsg:%v\nreceiver:%v", msg, receiver))
	}
	t.Log(utils.Norf("End Receiver"))
}
