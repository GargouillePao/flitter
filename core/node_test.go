package core

import (
	utils "github.com/GargouillePao/flitter/utils"
	"net"
	"testing"
)

func Test_NetHostPort(t *testing.T) {
	hostport := "127.0.0.5:8080"
	host, port, err := net.SplitHostPort(hostport)
	if err == nil {
		t.Log(utils.Infof("Succeed host:%s,port:%s", host, port))
	} else {
		t.Log(utils.Errf("Node Create Faild:,%v", err))
		t.Fail()
	}
}

func Test_NodeInfo(t *testing.T) {
	t.Log(utils.Norf("Start NodeInfo"))
	info := NewNodeInfo()
	info.SetAddr("127.0.0.1", "8080")
	info.SetPath("/rout/1/1")
	path, addr := info.Info()
	if addr == "127.0.0.1:8080" && path == "/rout/1/1" {
		t.Log(utils.Infof("NodeInfo Made Succeed: %v", info))
	} else {
		t.Log(utils.Errf("NodeInfo Made Faild: %v", info))
		t.Fail()
	}

	name := info.GetName()
	if name == "1" {
		t.Log(utils.Infof("NodeInfo GetName Succeed: %v", name))
	} else {
		t.Log(utils.Errf("NodeInfo GetName Faild: %v", name))
		t.Fail()
	}
	endpointr, err := info.GetEndpoint(true)
	endpointl, err := info.GetEndpoint(false)
	if err == nil && endpointr == "tcp://127.0.0.1:8080" && endpointl == "tcp://*:8080" {
		t.Log(utils.Infof("NodeInfo GetEndpoint Succeed: [remote:%v,local:%v]", endpointr, endpointl))
	} else {
		t.Log(utils.Errf("NodeInfo GetEndpoint Faild: [remote:%v,local:%v]", endpointr, endpointl))
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}

	t.Log(utils.Norf("End NodeInfo"))
}

func Test_NodeInfo_Serialize(t *testing.T) {
	t.Log(utils.Norf("Start NodeInfo Serilize"))
	serializer := NewSerializer()
	info := NewNodeInfo()
	host := "196.0.5.120"
	port := "8080"
	addr := "196.0.5.120:8080"
	path := "root/1/2"

	info.SetAddr(host, port)
	info.SetPath(path)
	_, buf, err := serializer.Encode(info)
	if err == nil {
		t.Log(utils.Infof("NodeInfo Encode Succeed: [buf:%v,info:%v]", buf, info))
	} else {
		t.Log(utils.Errf("NodeInfo Encode Faild: [buf:%v,info:%v]", buf, info))
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}

	info = NewNodeInfo()
	_, err = serializer.Decode(info, buf)
	pathtd, addrtd := info.Info()
	if err == nil && addrtd == addr && pathtd == path {
		t.Log(utils.Infof("NodeInfo Decode Succeed: [buf:%v,info:%v]", buf, info))
	} else {
		t.Log(utils.Errf("NodeInfo Decode Faild: [buf:%v,info:%v]", buf, info))
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}
	t.Log(utils.Norf("End NodeInfo Serilize"))
}

func Test_Node(t *testing.T) {
	t.Log(utils.Norf("Start Node"))
	info := NewNodeInfo()
	info.SetAddr("127.0.0.1", "8090")
	node, err := NewNode(info)
	if node != nil && err == nil {
		t.Log(utils.Infof("Node Create Succeed"))
	} else {
		t.Log(utils.Errf("Node Create Faild:,%v", err))
		t.Fail()
	}
	t.Log(utils.Norf("End Node"))
}

func Test_Node_Children(t *testing.T) {
	t.Log(utils.Norf("Start Node Children"))
	node, _ := NewNode(NewNodeInfo())
	child1 := NewNodeInfo()
	child2 := NewNodeInfo()
	child3 := NewNodeInfo()
	child4 := NewNodeInfo()
	child5 := NewNodeInfo()

	child1.SetAddr("167.0.0.1", "8090")
	child2.SetAddr("167.1.0.2", "7090")
	child3.SetAddr("167.2.0.3", "6090")
	node.SetChildren([]NodeInfo{
		child1,
		child2,
		child3,
	})
	t.Log(utils.Infof("Node Children Succeed1,%v", node))

	child4.SetAddr("167.3.0.4", "5090")
	child5.SetAddr("167.4.0.5", "4090")
	node.SetChildren([]NodeInfo{
		child1,
		child4,
		child5,
	})
	t.Log(utils.Infof("Node Children Succeed2,%v", node))
	t.Log(utils.Norf("End Node Children"))
}

func Test_Node_Leader(t *testing.T) {
	t.Log(utils.Norf("Start Node Leader"))
	info := NewNodeInfo()
	node, _ := NewNode(info)
	leader1 := NewNodeInfo()
	leader2 := NewNodeInfo()
	leader3 := NewNodeInfo()

	leader1.SetAddr("167.0.0.1", "8090")
	leader2.SetAddr("167.1.0.2", "7090")
	leader3.SetAddr("167.1.0.2", "7090")
	node.SetLeader(leader1)
	t.Log(utils.Infof("Node Leader Succeed1,%v", node))
	node.SetLeader(leader2)
	t.Log(utils.Infof("Node Leader Succeed2,%v", node))
	node.SetLeader(leader3)
	t.Log(utils.Infof("Node Leader Succeed3,%v", node))
	t.Log(utils.Norf("End Node Leader"))
}
