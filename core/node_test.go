package core

import (
	utils "github.com/gargous/flitter/utils"
	"testing"
)

func Test_NodeInfo(t *testing.T) {
	t.Log(utils.Norf("Start NodeInfo"))
	info := NewNodeInfo()
	err := info.Parse("127.0.0.3:8080")
	if err != nil {
		t.Fatal(utils.Errf("Parse Info %v", err))
	}
	ends := info.GetEndpoint(false)
	if ends == "tcp://127.0.0.3:8080" {
		t.Log(utils.Infof("Get Endpoint:%v", ends))
	} else {
		t.Fatalf(utils.Errf("No Endpoint:%v", ends))
	}
	if info.Name == "" {
		t.Fatalf(utils.Errf("No Name"))
	}
	if info.Host == "127.0.0.3" {
		t.Log(utils.Infof("Get Host:%v", info))
	} else {
		t.Fatalf(utils.Errf("No Host:%v", info))
	}
	if info.Port == 8080 {
		t.Log(utils.Infof("Get Port:%v", info))
	} else {
		t.Fatalf(utils.Errf("No Port:%v", info))
	}
	t.Log(utils.Norf("End NodeInfo"))
}
func Test_NodePath(t *testing.T) {
	t.Log(utils.Norf("Start NodePath"))
	npath := NodePath("/127.0.0.1:8000/127.0.0.2:7000/127.0.0.3:8080")
	info, ok := npath.GetNodeInfo()
	if ok && info.String() == "127.0.0.3:8080" {
		t.Log(utils.Infof("Get Info:%v", info))
	} else {
		t.Log(utils.Errf("No Info:%v", info))
	}
	leader, ok := npath.GetLeaderPath()
	if ok && leader == "/127.0.0.1:8000/127.0.0.2:7000" {
		t.Log(utils.Infof("Get Leader:%v", leader))
	} else {
		t.Fatalf(utils.Errf("No Leader:%v", leader))
	}
	t.Log(utils.Norf("End NodePath"))
}
func Test_NodeTree(t *testing.T) {
	t.Log(utils.Norf("Start Node Tree"))
	tree := NewNodeTree()
	node, err := tree.Add("root@127.0.0.1:8000")
	if err != nil {
		t.Fatal(utils.Errf("Add Root %v\ntree:%v", node, tree))
	} else {
		t.Log(utils.Infof("Add Root %v\ntree:%v", node, tree))
	}
	info := NewNodeInfo()
	info.Parse("root@127.0.0.1:8000")
	npath, ok := tree.Search(info)
	if ok && npath == "root@127.0.0.1:8000" {
		t.Log(utils.Infof("Search Root %v\ntree:%v", npath, tree))
	} else {
		t.Fatal(utils.Errf("Search Root %v\ntree:%v", npath, tree))
	}
	tree.Add("scene@127.0.0.1:8080")
	npath, ok = tree.SearchWithName("scene")
	if ok && npath == "root@127.0.0.1:8000/scene@127.0.0.1:8080" {
		t.Log(utils.Infof("Search WithName Scene %v\ntree:%v", npath, tree))
	} else {
		t.Fatal(utils.Errf("Search WithName Scene %v\ntree:%v", npath, tree))
	}
	tree.Add("scene@0:0/127.0.0.1:7080")
	tree.Add("scene@0:0/127.0.0.2:7080")
	tree.Add("scene@0:0/127.0.0.3:7080")
	tree.Add("scene@0:0/127.0.0.4:7080")
	tree.Add("scene@0:0/127.0.0.5:7080")
	tree.Add("scene@0:0/127.0.0.6:7080")
	tree.Add("scene@0:0/127.0.0.7:7080")
	npath, err = tree.Add("scene@0:0/scenex@127.0.0.8:7080")
	if err == nil && npath == NodePath("root@127.0.0.1:8000/scene@127.0.0.1:8080/scenex@127.0.0.8:7080") {
		t.Log(utils.Infof("Add scenex %v\ntree:%v", npath, tree))
	} else {
		t.Fatal(utils.Errf("Add scenex %v\ntree:%v\nerr:%v", npath, tree, err))

	}
	t.Log(utils.Norf("End Node Tree"))
}
