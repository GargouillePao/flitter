package core

import (
	//"fmt"
	utils "github.com/GargouillePao/flitter/utils"
	"testing"
)

func Test_NodeInfo(t *testing.T) {
	t.Log(utils.Norf("Start NodeInfo"))
	info := NodeInfo("/127.0.0.1:8000/127.0.0.2:7000/127.0.0.3:8080")
	ends, err := info.GetEndpoint()
	if err != nil {
		t.Fatalf(utils.Errf("get endpoint error:%v", err))
	}
	if ends == "tcp://127.0.0.3:8080" {
		t.Log(utils.Infof("get endpoint:%v", ends))
	} else {
		t.Log(utils.Errf("no endpoint"))
	}
	leader, err := info.GetLeaderInfo()
	if err != nil {
		t.Fatalf(utils.Errf("get leader error:%v", err))
	}
	if leader == "/127.0.0.1:8000/127.0.0.2:7000" {
		t.Log(utils.Infof("get leader:%v", leader))
	} else {
		t.Log(utils.Errf("no leader"))
	}
	t.Log(utils.Norf("End NodeInfo"))
}
