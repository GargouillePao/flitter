package core

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const (
	__anonymous string = "Anonymous"
)

type NodePath string

func NewNodePath(infos ...NodeInfo) NodePath {
	var npath NodePath
	for index, info := range infos {
		if index != 0 {
			npath += "/"
		}
		npath += NodePath(info.String())
	}
	return npath
}
func (n *NodePath) Append(infos ...NodeInfo) {
	opath := *n
	npath := NewNodePath(infos...)
	*n = opath + "/" + npath
}
func (n *NodePath) AppendPath(npath NodePath) {
	*n = *n + "/" + npath
}
func (n NodePath) GetLeaderPath() (lpath NodePath, ok bool) {
	npath := strings.Split(string(n), "/")
	if len(npath) <= 1 {
		ok = false
		return
	}
	lpath = NodePath(strings.Join(npath[:len(npath)-1], "/"))
	ok = true
	return
}
func (n NodePath) GetNodeInfo() (info NodeInfo, ok bool) {
	npath := strings.Split(string(n), "/")
	if len(npath) < 1 {
		ok = false
		return
	}
	info = NewNodeInfo()
	if info.Parse(npath[len(npath)-1]) != nil {
		ok = false
	}
	ok = true
	return
}

type NodeInfo struct {
	Name string
	Host string
	Port int
}

func NewNodeInfo() NodeInfo {
	return NodeInfo{}
}
func (n *NodeInfo) Parse(info string) (err error) {
	attrs := strings.Split(info, "@")
	var (
		name string
		addr string
	)
	switch len(attrs) {
	case 1:
		addr = attrs[0]
		name = __anonymous
	case 2:
		name = attrs[0]
		addr = attrs[1]
	default:
		err = errors.New("Invalid NodeInfo")
		return
	}
	n.Name = name
	colon := strings.LastIndexAny(addr, ":")
	if colon > 0 && colon < len(addr)-1 {
		n.Host = addr[:colon]
		n.Port, err = strconv.Atoi(addr[colon+1:])
		return
	}
	err = errors.New("Invalid NodeInfo")
	return
}

func (n *NodeInfo) RandName() {
	if n.Name == __anonymous {
		name := ""
		strpull := "abcdefghijklmnopqrstuvwxyz1234567890"
		for i := 0; i < 5; i++ {
			name += fmt.Sprintf("%c", strpull[rand.Intn(len(strpull))])
		}
		n.Name = name
	}
	return
}

func (n NodeInfo) GetEndpoint(local bool) (str string) {
	if local {
		str = fmt.Sprintf("tcp://*:%d", n.Port)
	} else {
		str = fmt.Sprintf("tcp://%s:%d", n.Host, n.Port)
	}
	return
}
func (n NodeInfo) GetAddress() (str string) {
	str = fmt.Sprintf("%s:%d", n.Host, n.Port)
	return
}

func (n NodeInfo) String() string {
	return fmt.Sprintf("%s@%s", n.Name, n.GetAddress())
}

const (
	__TreeWidth int = 5
)

type Node struct {
	Children []*Node
	Info     NodeInfo
	Weight   int
}

func NewNode(info NodeInfo) *Node {
	return &Node{
		Info:     info,
		Weight:   0,
		Children: make([]*Node, 0),
	}
}
func (n *Node) appendChild(info NodeInfo) NodePath {
	if info.Name == __anonymous {
		info.RandName()
		n.Children = append(n.Children, NewNode(info))
		return NewNodePath(n.Info, info)
	}
	n.Children = append(n.Children, NewNode(info))
	return NewNodePath(n.Info, info)
}
func (n *Node) AddAt(info NodeInfo, parentName string) (npath NodePath, ok bool) {
	ok = false
	tnode := n.FLoopForNode(0, func(height int, tinfo NodeInfo) bool {
		if parentName == tinfo.Name {
			return true
		}
		return false
	})
	if tnode != nil {
		ok = true
		npath = NewNodePath(n.Info)
		outpath := tnode.Add(info)
		if n == tnode {
			npath = outpath
		} else {
			npath.AppendPath(outpath)
		}
		return
	}
	return
}
func (n *Node) Add(info NodeInfo) NodePath {
	n.Weight++
	if len(n.Children) < __TreeWidth {
		return n.appendChild(info)
	} else {
		nextnode := n.Children[0]
		minweight := nextnode.Weight
		for i := 1; i < len(n.Children); i++ {
			if minweight > n.Children[i].Weight {
				minweight = n.Children[i].Weight
				nextnode = n.Children[i]
			}
		}
		cpath := nextnode.Add(info)
		npath := NewNodePath(n.Info)
		npath.AppendPath(cpath)
		return npath
	}
}
func (n *Node) Remove(info NodeInfo) NodeInfo {
	return NewNodeInfo()
}
func (n *Node) FLoop(height int, cb func(height int, node NodeInfo) bool) (breakoutPath NodePath) {
	height++
	if cb(height, n.Info) {
		breakoutPath = NewNodePath(n.Info)
		return
	}
	for _, childNode := range (*n).Children {
		bnode := childNode.FLoop(height, cb)
		if bnode != "" {
			breakoutPath = NewNodePath(n.Info)
			breakoutPath.AppendPath(bnode)
			return
		}
	}
	return
}
func (n *Node) FLoopForNode(height int, cb func(height int, node NodeInfo) bool) (breakoutNode *Node) {
	height++
	if cb(height, n.Info) {
		breakoutNode = n
		return
	}
	for _, childNode := range (*n).Children {
		pnode := childNode.FLoopForNode(height, cb)
		if pnode != nil {
			breakoutNode = pnode
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
	GetLastAdd() NodePath
	GetLastRemove() NodePath
	Search(info NodeInfo) (npath NodePath, ok bool)
	SearchWithName(name string) (npath NodePath, ok bool)
	SearchWithAddr(addr string) (npath NodePath, ok bool)
	Add(ipath NodePath) (npath NodePath, err error)
	Remove(ipath NodePath) (err error)
	FLoop(height int, cb func(height int, node NodeInfo) bool) (npath NodePath)
	FLoopGroup(groupname string, cb func(height int, node NodeInfo) bool) (breakoutPath NodePath)
	String() string
}
type nodeTree struct {
	node       *Node
	lastAdd    NodePath
	lastRemove NodePath
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
func (n *nodeTree) GetLastAdd() NodePath {
	return n.lastAdd
}
func (n *nodeTree) GetLastRemove() NodePath {
	return n.lastRemove
}
func (n *nodeTree) Remove(ipath NodePath) (err error) {
	n.lastRemove = ipath
	return
}
func (n *nodeTree) FLoop(height int, cb func(height int, node NodeInfo) bool) (breakoutPath NodePath) {
	if n.node == nil {
		return ""
	}
	breakoutPath = n.node.FLoop(height, cb)
	return
}
func (n *nodeTree) FLoopGroup(groupname string, cb func(height int, node NodeInfo) bool) (breakoutPath NodePath) {
	if n.node == nil {
		return ""
	}
	target := n.node.FLoopForNode(0, func(height int, node NodeInfo) bool {
		if node.Name == groupname {
			return true
		}
		return false
	})
	if target != nil {
		breakoutPath = target.FLoop(0, cb)
	}
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
func (n *nodeTree) Add(ipath NodePath) (npath NodePath, err error) {
	n.lastAdd = ipath
	info, ok := ipath.GetNodeInfo()
	if !ok {
		err = errors.New("Invalid NodePath")
		return
	}
	if n.node == nil {
		n.node = NewNode(info)
		return NewNodePath(n.node.Info), nil
	}
	lpath, ok := ipath.GetLeaderPath()
	if ok {
		linfo, ok := lpath.GetNodeInfo()
		if ok {
			npath, ok = n.node.AddAt(info, linfo.Name)
			if !ok {
				err = errors.New("Your Leader Is Not Exsit")
			}
		}
		return
	}
	npath = n.node.Add(info)
	return
}

func (n *nodeTree) searchWith(withWhatCall func(NodeInfo) bool) (newPath NodePath, ok bool) {
	ok = false
	if n.node == nil {
		return
	}
	searchedPath := n.FLoop(0, func(height int, node NodeInfo) bool {
		return withWhatCall(node)
	})
	if searchedPath != "" {
		newPath = searchedPath
		ok = true
		return
	}
	return
}

func (n *nodeTree) Search(info NodeInfo) (newPath NodePath, ok bool) {
	newPath, ok = n.searchWith(func(selected NodeInfo) bool {
		if selected == info {
			return true
		} else {
			return false
		}
	})
	return
}
func (n *nodeTree) SearchWithName(name string) (newPath NodePath, ok bool) {
	newPath, ok = n.searchWith(func(selected NodeInfo) bool {
		if selected.Name == name {
			return true
		} else {
			return false
		}
	})
	return
}
func (n *nodeTree) SearchWithAddr(addr string) (newPath NodePath, ok bool) {
	newPath, ok = n.searchWith(func(selected NodeInfo) bool {
		if selected.GetAddress() == addr {
			return true
		} else {
			return false
		}
	})
	return
}
