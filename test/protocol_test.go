package test

import (
	"bytes"
	"github.com/gargous/flitter/core"
	msgids "github.com/gargous/flitter/share/proto"
	"testing"
)

func TestEncoding(t *testing.T) {
	b := bytes.NewBuffer(nil)
	msg1 := core.NewMsg(
		msgids.PID_LOGIN_REQ,
		msgids.Role{
			Id:   123,
			Name: "xxx",
		},
	)
	err := core.EncodeMsg(b, msg1)
	if err != nil {
		t.Fatal(err)
	}
	msg2 := core.NewMsg(
		uint32(0),
		msgids.Role{},
	)
	err = core.DecodeMsg(b, msg2)
	if err != nil {
		t.Fatal(err)
	}
	if msg1.Head != msg2.Head {
		t.Fatalf("Head %d,%d", msg1.Head, msg2.Head)
	}
	if msg1.GetId() != msg2.GetId() {
		t.Fatalf("Body %v,%v", msg1, msg2)
	}
}
