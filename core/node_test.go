package core

import (
	"fmt"
	utils "github.com/GargouillePao/flitter/utils"
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
	tree := NewNodeTree("root")
	_, ok := tree.Search("root")
	if !ok {
		t.Fatal(utils.Errf("Failed Search root %v", tree))
	}
	node1 := tree.Add("garg1")
	_, ok = tree.Search("garg1")
	if !ok {
		t.Fatal(utils.Errf("Failed Search garg1 %v", tree))
	}
	tree.Add("garg2")
	tree.Add("garg3")
	tree.Add("garg4")
	tree.Add("garg5")
	node6 := tree.Add("garg6")
	node7 := tree.Add("garg7")

	if string(node1) == "root/garg1" {
		t.Log(utils.Infof("OK garg1 %v,%v", tree, node1))
	} else {
		t.Fatal(utils.Errf("Failed garg1 %v", tree))
	}
	if string(node6) == "root/garg1/garg6" {
		t.Log(utils.Infof("OK garg6 %v,%v", tree, node6))
	} else {
		t.Fatal(utils.Errf("Failed garg6 %v", tree))
	}
	if string(node7) == "root/garg2/garg7" {
		t.Log(utils.Infof("OK garg7 %v,%v", tree, node7))
	} else {
		t.Fatal(utils.Errf("Failed garg7 %v", tree))
	}
	node, ok := tree.Search("garg7")
	if !ok {
		t.Fatal(utils.Errf("Search %v", ok))
	}
	for i := 0; i < 30; i++ {
		tree.Add(fmt.Sprintf("garg1%v", i))
	}
	if string(node) == "root/garg2/garg7" {
		t.Log(utils.Infof("OK Search garg7 %v,%v", tree, node))
	} else {
		t.Fatal(utils.Errf("Failed Search garg7 %v", tree))
	}
	t.Log(utils.Norf("End Node Tree"))
}
