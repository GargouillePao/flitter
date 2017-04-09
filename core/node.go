package core

import (
	"errors"
	"fmt"
	"strings"
)

type NodeInfo string

func (n NodeInfo) GetLeaderInfo() (NodeInfo, error) {
	nodes := strings.Split(string(n), "/")
	if len(nodes) > 0 {
		nodes = nodes[:len(nodes)-1]
		return NodeInfo(strings.Join(nodes, "/")), nil
	}
	return "", errors.New("Invalid node info")
}

func (n NodeInfo) GetEndpoint(local bool) (string, error) {
	nodes := strings.Split(string(n), "/")
	if len(nodes) > 0 {
		if local {
			addr := strings.Split(nodes[len(nodes)-1], ":")
			if len(addr) >= 1 {
				port := addr[len(addr)-1]
				return "tcp://*:" + port, nil
			}
		} else {
			return "tcp://" + nodes[len(nodes)-1], nil
		}
	}
	return "", errors.New("Invalid node info")
}

func (n NodeInfo) GetAddressInfo() (string, error) {
	nodes := strings.Split(string(n), "/")
	if len(nodes) > 0 {
		return nodes[len(nodes)-1], nil
	}
	return "", errors.New("Invalid node info")
}

const (
	__TreeWidth int = 5
)

type NodeTree interface {
	Search(address string) (newInfo NodeInfo, ok bool)
	Add(address string) NodeInfo
	Remove(address string)
	GetWeight() int
	GetInfo() NodeInfo
	SetWeight(weight int)
	FLoop(height int, cb func(height int, node NodeTree) bool) NodeTree
	String() string
}

type nodeTree struct {
	childs []NodeTree
	info   NodeInfo
	weight int
}

func NewNodeTree(info NodeInfo) NodeTree {
	return &nodeTree{
		info:   info,
		childs: make([]NodeTree, 0),
		weight: 0,
	}
}
func (n *nodeTree) Remove(address string) {

}
func (n *nodeTree) SetWeight(weight int) {
	n.weight = weight
}
func (n *nodeTree) GetWeight() int {
	return n.weight
}
func (n *nodeTree) GetInfo() NodeInfo {
	return n.info
}
func (n *nodeTree) FLoop(height int, cb func(height int, node NodeTree) bool) (breakoutNode NodeTree) {
	height++
	if cb(height, n) {
		breakoutNode = n
	}
	for _, childNode := range n.childs {
		_node := childNode.FLoop(height, cb)
		if _node != nil {
			breakoutNode = _node
		}
	}
	return
}
func (n *nodeTree) String() string {
	info := ""
	n.FLoop(0, func(height int, node NodeTree) bool {
		info += "\n"
		info += strings.Repeat("\t", height)
		info += fmt.Sprintf("info:%v,weight:%v", node.GetInfo(), node.GetWeight())
		return false
	})
	return info
}
func (n *nodeTree) Add(address string) (newInfo NodeInfo) {
	n.SetWeight(n.GetWeight() + 1)
	if len(n.childs) > 0 {
		nextnode := n.childs[0]
		minweight := nextnode.GetWeight()
		for i := 1; i < len(n.childs); i++ {
			if minweight > n.childs[i].GetWeight() {
				minweight = n.childs[i].GetWeight()
				nextnode = n.childs[i]
			}
		}
		if len(n.childs) >= __TreeWidth {
			newInfo = nextnode.Add(address)
		} else {
			newInfo = NodeInfo(string(n.info) + "/" + address)
			n.childs = append(n.childs, NewNodeTree(newInfo))
		}
	} else {
		newInfo = NodeInfo(string(n.info) + "/" + address)
		n.childs = append(n.childs, NewNodeTree(newInfo))
	}
	return
}
func (n *nodeTree) Search(address string) (newInfo NodeInfo, ok bool) {
	searchedNode := n.FLoop(0, func(height int, node NodeTree) bool {
		addr, err := node.GetInfo().GetAddressInfo()
		if err == nil && addr == address {
			return true
		} else {
			return false
		}
	})
	if searchedNode != nil {
		return searchedNode.GetInfo(), true
	}
	return "", false
}
