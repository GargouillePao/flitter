package core

import (
	"errors"
	"fmt"
	"strconv"
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
	addr, err := n.GetAddress()
	if err != nil {
		return "", err
	}
	colonpos := strings.LastIndexAny(addr, ":")
	if colonpos > 0 && colonpos <= len(addr)-1 {
		if local {
			port := addr[colonpos+1:]
			return "tcp://*:" + port, nil
		} else {
			host := addr[:colonpos]
			port := addr[colonpos+1:]
			return "tcp://" + host + ":" + port, nil
		}
	} else {
		if local {
			return "inproc://flitter" + addr, nil
		}
	}
	return "", errors.New("Invalid node info")
}

func (n NodeInfo) GetAddress() (string, error) {
	nodes := strings.Split(string(n), "/")
	if len(nodes) > 0 {
		return nodes[len(nodes)-1], nil
	}
	return "", errors.New("Invalid node info")
}
func (n NodeInfo) GetPort() (int, error) {
	addr, err := n.GetAddress()
	if err != nil {
		return 0, err
	}
	colonpos := strings.LastIndexAny(addr, ":")
	if colonpos < 0 || colonpos >= len(addr)-1 {
		err = errors.New("Invalid Address When Get Port")
		return 0, err
	}
	port := addr[colonpos+1:]
	return strconv.Atoi(port)
}
func (n NodeInfo) GetHost() (string, error) {
	addr, err := n.GetAddress()
	if err != nil {
		return "", err
	}
	colonpos := strings.LastIndexAny(addr, ":")
	if colonpos < 0 || colonpos >= len(addr)-1 {
		err = errors.New("Invalid Address When Get Host")
		return "", err
	}
	host := addr[:colonpos]
	return host, err
}

const (
	__TreeWidth int = 5
)

type Node struct {
	Children []*Node
	Info     NodeInfo
	Weight   int
}

func NewNode(addr string) *Node {
	return &Node{
		Info:     NodeInfo(addr),
		Weight:   0,
		Children: make([]*Node, 0),
	}
}
func (n *Node) Add(address string) NodeInfo {
	n.Weight++
	if len(n.Children) < __TreeWidth {
		n.Children = append(n.Children, NewNode(address))
		return NodeInfo(string(n.Info) + "/" + address)
	} else {
		nextnode := n.Children[0]
		minweight := nextnode.Weight
		for i := 1; i < __TreeWidth; i++ {
			if minweight > n.Children[i].Weight {
				minweight = n.Children[i].Weight
				nextnode = n.Children[i]
			}
		}
		return NodeInfo(n.Info + "/" + nextnode.Add(address))
	}
}
func (n *Node) Remove(address string) NodeInfo {
	// node := n.FLoop(0, func(height int, node Node) bool {
	// 	if address == string(node) {
	// 		//root
	// 		return true
	// 	}
	// 	for i := 0; i < len(node.Children); i++ {
	// 		if string(node.Children[i].Info) == address {
	// 			//exsit
	// 			//if the node to be remove has children
	// 			if len(node.Children[i].Children) > 0 {
	// 				for j := 0; len(node.Children[i].Children) > 0; j = (j + 1) % len(node.Children) {
	// 					if i != j {
	// 						node.Children[j].Children = append(node.Children[j].Children, node.Children[i].Children[0])
	// 						if len(node.Children[i].Children) > 1 {
	// 							node.Children[i].Children = node.Children[i].Children[1:]
	// 						} else {
	// 							node.Children[i].Children = make([]*Node, 0)
	// 						}
	// 					}
	// 				}
	// 			} else {
	// 				if i >= len(node.Children)-1 {
	// 					node.Children = node.Children[:i]
	// 				} else if i <= 0 {
	// 					node.Children = node.Children[i+1:]
	// 				} else {
	// 					node.Children = append(node.Children[:i], node.Children[i+1:]...)
	// 				}
	// 			}
	// 			return true
	// 		}
	// 	}
	// 	return false
	// })
	return ""
}
func (n *Node) FLoop(height int, cb func(height int, node NodeInfo) bool) (breakoutNode NodeInfo) {
	height++
	if cb(height, n.Info) {
		breakoutNode = n.Info
		return
	}
	for _, childNode := range (*n).Children {
		bnode := childNode.FLoop(height, cb)
		if bnode != "" {
			breakoutNode = n.Info + "/" + bnode
			return
		}
	}
	return
}
func (n Node) String() string {
	return fmt.Sprintf("info:%v,weight:%d,child:%p", n.Info, n.Weight, n.Children)
}

type NodeTree interface {
	SetNode(node *Node)
	GetNode() *Node
	Search(address string) (newInfo NodeInfo, ok bool)
	Add(address string) NodeInfo
	Remove(address string)
	FLoop(height int, cb func(height int, node NodeInfo) bool) (breakoutNode NodeInfo)
	String() string
}
type nodeTree struct {
	node *Node
}

func NewNodeTree() NodeTree {
	return &nodeTree{}
}
func (n *nodeTree) SetNode(node *Node) {
	n.node = node
}
func (n *nodeTree) GetNode() *Node {
	return n.node
}
func (n *nodeTree) Remove(address string) {

}
func (n *nodeTree) FLoop(height int, cb func(height int, node NodeInfo) bool) (breakoutNode NodeInfo) {
	if n.node == nil {
		return ""
	}
	breakoutNode = n.node.FLoop(height, cb)
	return
}
func (n *nodeTree) String() string {
	info := ""
	if n.node == nil {
		return fmt.Sprintf("%p", n.node)
	}
	n.FLoop(0, func(height int, node NodeInfo) bool {
		info += "\n"
		info += strings.Repeat("\t", height)
		info += fmt.Sprintf("%v", node)
		return false
	})
	return info
}
func (n *nodeTree) Add(address string) NodeInfo {
	if n.node == nil {
		n.node = NewNode(address)
		return n.node.Info
	} else {
		return n.node.Add(address)
	}
}

func (n *nodeTree) Search(address string) (newInfo NodeInfo, ok bool) {
	newInfo = NodeInfo(address)
	ok = false
	if n.node == nil {
		return
	}
	searchedNode := n.FLoop(0, func(height int, node NodeInfo) bool {
		addr, err := node.GetAddress()
		if err == nil && addr == address {
			return true
		} else {
			return false
		}
	})
	if searchedNode != "" {
		newInfo = searchedNode
		ok = true
		return
	}
	return
}
