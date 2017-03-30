package core

import (
	"bytes"
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"net"
	"path"
	"strings"
)

/*(NULL)*/
type Node interface {
	/** handle the message of upSender */
	SendToLeader(message Message) error
	/** handle the message of downSender */
	SendToChildren(message Message) error
	/** handle the message og upReceiver */
	ReceiveFromLeader() (Message, error)
	/** handle the message of the downReceiver */
	ReceiveFromChilren() (Message, error)
	/** bind  downReceiver and upReceiver */
	SetChildren(children []NodeInfo)
	ConnectChildren() error
	ConnectLeader() error
	SetLeader(leader NodeInfo) error
	Bind() error
	Close()
	String() string
}

/*(NULL)*/
type node struct {
	info               NodeInfo
	leader             NodeInfo
	children           []NodeInfo
	oldLeaderEndpoints []string
	oldChildEndpoints  []string
	upSender           *zmq.Socket
	downSender         *zmq.Socket
	upReceiver         *zmq.Socket
	downReceiver       *zmq.Socket
}

func NewNode(info NodeInfo) (Node, error) {
	upSender, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		return nil, err
	}
	downSender, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}
	upReceiver, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}
	downReceiver, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		return nil, err
	}
	node := &node{
		info:               info,
		children:           make([]NodeInfo, 0),
		oldLeaderEndpoints: make([]string, 0),
		oldChildEndpoints:  make([]string, 0),
		upSender:           upSender,
		downSender:         downSender,
		upReceiver:         upReceiver,
		downReceiver:       downReceiver,
	}
	return node, nil
}

/** bind  downReceiver and upReceiver */
func (n *node) Bind() error {
	endpoint, err := n.info.GetEndpoint(false)
	if err != nil {
		return err
	}

	err = n.downSender.Bind(endpoint)
	if err != nil {
		return err
	}
	n.upReceiver.SetSubscribe("")
	n.downReceiver.Bind(endpoint)
	return nil
}

func (n *node) Close() {
	n.upSender.Close()
	n.downSender.Close()
	n.upReceiver.Close()
	n.downReceiver.Close()
}

/** handle the message of upSender */
func (n *node) SendToLeader(message Message) error {
	return errors.New("Err")
}

/**
disconnect old children
connect children
*/
func (n *node) ConnectChildren() error {
	errout := errors.New("No Child")
	for _, oldEnds := range n.oldChildEndpoints {
		n.downSender.Disconnect(oldEnds)
	}
	for _, newChild := range n.children {
		newEnd, err := newChild.GetEndpoint(true)
		if err != nil {
			continue
		}
		err = n.downSender.Connect(newEnd)
		if err != nil {
			continue
		}
		errout = nil
	}
	return errout
}

/**
disconnect old leader
connect leader
*/
func (n *node) ConnectLeader() error {
	errout := errors.New("No Leader")
	for _, oldEnds := range n.oldLeaderEndpoints {
		n.upReceiver.Disconnect(oldEnds)
	}
	if n.leader != nil {
		var newEnd string
		newEnd, errout = n.leader.GetEndpoint(true)
		if errout != nil {
			return errout
		}
		errout = n.upReceiver.Connect(newEnd)
	}
	return errout
}

/** handle the message of downSender*/
func (n *node) SendToChildren(msg Message) error {
	//err := n.ConnectChildren()
	err := sendMsg(n.downSender, msg)
	return err
}

/** handle the message og upReceiver */
func (n *node) ReceiveFromLeader() (Message, error) {
	err := n.ConnectLeader()
	if err != nil {
		return nil, err
	}
	msg, err := receiveMsg(n.upReceiver)
	return msg, err
}

/** handle the message of the downReceiver */
func (n *node) ReceiveFromChilren() (Message, error) {
	return nil, errors.New("Err")
}

/**
Find the children where in the oldChildren and not in the newChildren
to become the new oldChildren
*/
func (n *node) SetChildren(children []NodeInfo) {
	newChildrenEndpoint := make([]string, len(children))
	set := NewStringSet()
	for i := 0; i < len(children); i++ {
		newend, err := children[i].GetEndpoint(true)
		if err != nil {
			continue
		}
		newChildrenEndpoint[i] = newend
	}
	oldChildrenEndpoint := make([]string, len(n.children))
	for i := 0; i < len(n.children); i++ {
		oldend, err := n.children[i].GetEndpoint(true)
		if err != nil {
			continue
		}
		oldChildrenEndpoint[i] = oldend
	}
	differ := set.Minus(oldChildrenEndpoint, newChildrenEndpoint)
	n.oldChildEndpoints = differ
	n.children = children
	return
}

/**
Check whether the new leader is == nowleader and in the oldLeaders
*/
func (n *node) SetLeader(leader NodeInfo) error {
	set := NewStringSet()
	newEnd, err := leader.GetEndpoint(true)
	if err != nil {
		return err
	}
	if n.leader != nil {
		oldEnd, err := n.leader.GetEndpoint(true)
		if err != nil {
			return err
		}
		if oldEnd != newEnd {
			n.oldLeaderEndpoints = append(n.oldLeaderEndpoints, oldEnd)
		}
	}
	index := set.IndexOf(n.oldLeaderEndpoints, newEnd)
	if index >= 0 && index < len(n.oldLeaderEndpoints)-1 {
		n.oldLeaderEndpoints = append(n.oldLeaderEndpoints[:index], n.oldLeaderEndpoints[index+1:]...)
	}
	n.leader = leader
	return nil
}
func (n *node) String() string {
	childrenStr := ""
	for _, child := range n.children {
		childrenStr += fmt.Sprintf("\n\t\t%v", strings.Join(strings.Split(fmt.Sprintf("%v", child), "\n"), "\n\t\t"))
	}
	str := fmt.Sprintf("\nNode[\n\tinfo:%s,\n\tleader:%s,\n\toldLeaders:%v,\n\tchildren:%v,\n\toldChildren:%v\n]",
		strings.Join(strings.Split(fmt.Sprintf("%v", n.info), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", n.leader), "\n"), "\n\t"),
		n.oldLeaderEndpoints,
		childrenStr,
		n.oldChildEndpoints,
	)
	return str
}

func sendMsg(socket *zmq.Socket, msg Message) error {
	serializer := NewSerializer()
	_, buf, err := serializer.Encode(msg.GetInfo())
	if err != nil {
		return err
	}
	buf = append(buf, msg.GetContent()...)
	_, err = socket.SendBytes(buf, 0)
	if err != nil {
		return err
	}
	return nil
}

func receiveMsg(socket *zmq.Socket) (Message, error) {
	buf, err := socket.RecvBytes(0)
	if err != nil {
		return nil, err
	}
	msgInfo := NewMessageInfo()
	serializer := NewSerializer()
	size, err := serializer.Decode(msgInfo, buf)
	if err != nil {
		return nil, err
	}
	msg := NewMessage(msgInfo, buf[size:])
	return msg, nil
}

type NodeInfo interface {
	GetName() string
	GetEndpoint(remote bool) (string, error)
	Info() (path string, addr string)
	SetPath(path string)
	SetAddr(host, port string)
	String() string
	Serializable
}

func NewNodeInfo() NodeInfo {
	info := &nodeInfo{
		path: "/",
		addr: "0.0.0.0",
	}
	return info
}

/*the data of node*/
type nodeInfo struct {
	size int
	path string
	addr string
}

/** read out to the buf [Encode]*/
func (n *nodeInfo) Read(buf []byte) (int, error) {
	err := errors.New("Invalide node to read")
	serialize := NewSerializer()
	if len(buf) >= n.Size() {
		buf[0] = byte(n.Size())
		seris := serialize.Serialize([]byte(n.path), []byte(n.addr))
		if len(seris) == n.Size()-1 {
			for i := 1; i <= len(seris); i++ {
				buf[i] = seris[i-1]
			}
			return n.Size(), nil
		}
	}
	return 0, err
}

/** write in from the buf [Decode]*/
func (n *nodeInfo) Write(buf []byte) (int, error) {
	//size := 0
	if len(buf) > 0 {
		size := int(buf[0])
		buf = buf[1:]
		if len(buf) >= size-1 {
			bufs := bytes.Split(buf, []byte("\n"))
			if len(bufs) == 2 {
				path := string(bufs[0][:])
				addr := string(bufs[1][:])
				n.path = path
				n.addr = addr
				return 0, nil
			}
		}
	}
	err := errors.New("Invalide node to write")
	return 0, err
}

func (n nodeInfo) Size() int {
	size := 1 + len(n.addr) + 1 + len(n.path)
	return size
}

func (n nodeInfo) Resize(size int) {
	n.size = size
}

/** return the endpoint for zmq*/
func (n *nodeInfo) GetEndpoint(remote bool) (string, error) {
	host, port, err := net.SplitHostPort(n.addr)
	if err != nil || port == "" {
		return "", err
	}
	if remote {
		return "tcp://" + host + ":" + port, nil
	} else {
		return "tcp://*:" + port, nil
	}
}

/** return the name(at the end of the path) */
func (n *nodeInfo) GetName() string {
	name := path.Base(n.path)
	return name
}

func (n *nodeInfo) Info() (path string, addr string) {
	path = n.path
	addr = n.addr
	return
}
func (n *nodeInfo) SetPath(path string) {
	n.path = path
}
func (n *nodeInfo) SetAddr(host, port string) {
	n.addr = net.JoinHostPort(host, port)
}
func (n nodeInfo) String() string {
	str := fmt.Sprintf("NodeInfo[\n\taddr:%v,\n\tpath:%v\n]", n.addr, n.path)
	return str
}
