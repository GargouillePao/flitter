package core

import (
	utils "github.com/gargous/flitter/utils"
	"testing"
)

func Test_NodeInfo(t *testing.T) {
	t.Log(utils.Norf("Start NodeInfo"))
	info := NodeInfo("/127.0.0.1:8000/127.0.0.2:7000/127.0.0.3:8080")
	ends, err := info.GetEndpoint(false)
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
func Test_NodeTree(t *testing.T) {
	t.Log(utils.Norf("Start Node Tree"))
	tree := NewNodeTree()
	node := tree.Add("root")
	_, ok := tree.Search("root")
	if !ok && node != "root" {
		t.Fatal(utils.Errf("Failed Search root %v,%v", tree, node))
	} else {
		t.Log(utils.Infof("add root\n%v,%v", tree, node))
	}
	tree.Add("garg1")
	_, ok = tree.Search("garg1")
	if !ok {
		t.Fatal(utils.Errf("Failed Search garg1 %v", tree))
	} else {
		t.Log(utils.Infof("add garg1\n%v", tree))
	}
	tree.Add("garg2")
	tree.Add("garg3")
	tree.Add("garg4")
	tree.Add("garg5")
	node = tree.Add("garg6")
	tree.Add("garg7")
	if node != "root/garg1/garg6" {
		t.Fatal(utils.Errf("Failed Add garg6 %v", node))
	} else {
		t.Log(utils.Infof("add garg6 %v", node))
	}
	t.Log(utils.Infof("OK \n%v", tree))
	t.Log(utils.Norf("End Node Tree"))
}
