package servers

import (
	//"errors"
	//"flag"
	//"fmt"
	utils "github.com/GargouillePao/flitter/utils"
	"testing"
	"time"
)

func Test_Watch(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Watch"))
	var err error
	server := NewWatchServer()
	err = server.Config(CA_Recv, ST_Name, "*:8000")
	if err != nil {
		t.Fatal(utils.Errf("recv name %v", err))
	}
	err = server.Config(CA_Send, ST_Name, "127.0.0.1:7000")
	if err != nil {
		t.Fatal(utils.Errf("send name %v", err))
	}
	err = server.Config(CA_Recv, ST_HeartBeat, "*:8001")
	if err != nil {
		t.Fatal(utils.Errf("recv heartbeat %v", err))
	}
	err = server.Config(CA_Send, ST_HeartBeat, "127.0.0.1:8081")
	if err != nil {
		t.Fatal(utils.Errf("send heartbeat %v", err))
	}
	err = server.Init()
	if err != nil {
		t.Fatal(utils.Errf("init %v", err))
	}
	time.AfterFunc(time.Second*3, func() {
		server.Term()
	})
	server.Start()
	t.Log(utils.Norf("End Watch"))
}
