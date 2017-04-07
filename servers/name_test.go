package servers

import (
	//"errors"
	//"flag"
	//"fmt"
	"time"
	//core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	"testing"
)

func Test_Name(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Name"))
	var err error
	server := NewNameServer()
	err = server.Config(CA_Recv, ST_Watch, "*:7000")
	if err != nil {
		t.Fatal(utils.Errf("%v", err))
	}
	err = server.Init()
	if err != nil {
		t.Fatal(utils.Errf("%v", err))
	}
	time.AfterFunc(time.Second*3, func() {
		server.Term()
	})
	server.Start()
	t.Log(utils.Norf("End Name"))
}
