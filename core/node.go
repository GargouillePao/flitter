package core

import (
	"bytes"
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"net"
	"path"
	"strconv"
	"strings"
)

/*(NULL)*/
type Node interface {
	/** handle the message of sender */
	SendToLeader(message Message) error
	/** handle the message of publisher */
	BroadcastToChildren(message Message) error
	/** handle the message og subscriber */
	ReceiveFromLeader() (Message, error)
	/** handle the message of the receiver */
	ReceiveFromChilren() (Message, error)
	ReplyToChild(childname string, msg Message) error
	/** bind  receiver and subscriber */
	SetChildren(children []NodeInfo)
	AddChild(child NodeInfo) error
	RemoveChild(child NodeInfo) error
	SetLeader(leader NodeInfo) error
	Info() (info NodeInfo, leader NodeInfo, chilren []NodeInfo)
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
	sender             *zmq.Socket
	publisher          *zmq.Socket
	subscriber         *zmq.Socket
	receiver           *zmq.Socket
}

func NewNode(info NodeInfo) (Node, error) {
	sender, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		return nil, err
	}
	receiver, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		return nil, err
	}
	publisher, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}
	subscriber, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}
	node := &node{
		info:               info,
		children:           make([]NodeInfo, 0),
		oldLeaderEndpoints: make([]string, 0),
		oldChildEndpoints:  make([]string, 0),
		sender:             sender,
		receiver:           receiver,
		publisher:          publisher,
		subscriber:         subscriber,
	}
	err = node.Bind()
	return node, err
}
func (n *node) Info() (info NodeInfo, leader NodeInfo, chilren []NodeInfo) {
	return n.info, n.leader, n.children
}

/** bind  receiver and publisher the rport = 1+pport */
func (n *node) Bind() error {
	err := n.subscriber.SetSubscribe("")
	if err != nil {
		return err
	}
	endpoint, err := n.info.GetEndpoint(false)
	if err != nil {
		return err
	}
	pend := getPublisherEndpoint(endpoint)

	rend, err := getReceiverEndpoint(endpoint)
	if err != nil {
		return err
	}

	err = n.publisher.Bind(pend)
	if err != nil {
		return err
	}
	fmt.Println("Bind Publish", pend)
	err = n.receiver.Bind(rend)
	if err != nil {
		return err
	}
	fmt.Println("Bind Receive", rend)
	return nil
}

func (n *node) Close() {
	n.sender.Close()
	n.publisher.Close()
	n.subscriber.Close()
	n.receiver.Close()
}

func (n *node) DisconnectChildren(all bool) {
	for _, oldEnds := range n.oldChildEndpoints {
		n.sender.Disconnect(oldEnds)
	}
	if all {
		for _, newChild := range n.children {
			newEnd, err := newChild.GetEndpoint(true)
			if err != nil {
				continue
			}
			err = n.sender.Disconnect(newEnd)
			if err != nil {
				continue
			}
		}
	}
}

/**
disconnect old children
connect children
*/
func (n *node) ConnectChild(childname string) error {
	errout := errors.New("No Child")
	n.DisconnectChildren(false)
	for _, newChild := range n.children {
		newEnd, err := newChild.GetEndpoint(true)
		if err != nil {
			continue
		}
		if childname == newChild.GetName() {
			return n.sender.Connect(newEnd)
		}
	}
	return errout
}

//return err but never mind,it will no make any trouble
func (n *node) DisconnectLeader(isPublish bool, all bool) error {
	for _, oldEnds := range n.oldLeaderEndpoints {
		if isPublish {
			pubend := getPublisherEndpoint(oldEnds)
			n.subscriber.Disconnect(pubend)
		} else {
			recvend, err := getReceiverEndpoint(oldEnds)
			if err == nil {
				n.sender.Disconnect(recvend)
			}
		}
	}
	if n.leader == nil {
		return errors.New("no leader when disconnect leader")
	}
	leaderEnds, err := n.leader.GetEndpoint(true)
	if err != nil {
		return err
	}
	if all && n.leader != nil {
		if isPublish {
			n.publisher.Disconnect(leaderEnds)
		} else {
			n.sender.Disconnect(leaderEnds)
		}
	}
	return nil
}

/**
disconnect old leader
connect leader
if isPublish then do with subscriber or do with sender
*/
func (n *node) ConnectLeader(isPublish bool) error {
	errout := errors.New("No Leader")

	errout = n.DisconnectLeader(isPublish, false)

	if n.leader != nil {
		var newEnd string
		newEnd, errout = n.leader.GetEndpoint(true)
		if errout != nil {
			return errout
		}
		if isPublish {
			pubend := getPublisherEndpoint(newEnd)
			errout = n.subscriber.Connect(pubend)
			if errout != nil {
				return errout
			}
		} else {
			recvend, errout := getReceiverEndpoint(newEnd)
			if errout != nil {
				return errout
			}
			fmt.Println("Connected", recvend)
			errout = n.sender.Connect(recvend)
			if errout != nil {
				return errout
			}
		}

	}
	return errout
}

/** handle the message of sender */
func (n *node) SendToLeader(msg Message) error {
	n.DisconnectChildren(true)
	err := n.ConnectLeader(false)
	if err != nil {
		return err
	}
	return sendMsg(n.sender, msg)
}

/** handle the message of publisher*/
func (n *node) BroadcastToChildren(msg Message) error {
	err := sendMsg(n.publisher, msg)
	return err
}

/** handle the message og subscriber */
func (n *node) ReceiveFromLeader() (Message, error) {
	err := n.ConnectLeader(true)
	if err != nil {
		return nil, err
	}
	msg, err := receiveMsg(n.subscriber)
	return msg, err
}

/** handle the message of the receiver */
func (n *node) ReceiveFromChilren() (Message, error) {
	msg, err := receiveMsg(n.receiver)
	return msg, err
}

/** handle the message of the receiver */
func (n *node) ReplyToChild(childname string, msg Message) error {
	err := n.DisconnectLeader(false, true)
	if err != nil {
		return err
	}
	err = n.ConnectChild(childname)
	if err != nil {
		return err
	}
	return sendMsg(n.sender, msg)
}

func (n *node) AddChild(child NodeInfo) error {
	newChildEndpoint, err := child.GetEndpoint(true)
	if err != nil {
		return err
	}
	for index, oldEnd := range n.oldChildEndpoints {
		if oldEnd == newChildEndpoint {
			if index >= len(n.oldChildEndpoints)-1 {
				n.oldChildEndpoints = n.oldChildEndpoints[:index]
			} else {
				n.oldChildEndpoints = append(n.oldChildEndpoints[:index], n.oldChildEndpoints[index+1:]...)
			}
			break
		}
	}
	for _, nowEnd := range n.children {
		if nowEnd.GetName() == child.GetName() {
			return nil
		}
	}
	n.children = append(n.children, child)
	return nil
}

func (n *node) RemoveChild(child NodeInfo) error {
	waistedChildEndpoint, err := child.GetEndpoint(true)
	if err != nil {
		return err
	}
	targindex := -1
	for index, nowEnd := range n.children {
		if nowEnd.GetName() == child.GetName() {
			targindex = index
			break
		}
	}
	if targindex >= 0 {
		if targindex >= len(n.children)-1 {
			n.children = n.children[:targindex]
		} else {
			n.children = append(n.children[:targindex], n.children[targindex+1:]...)
		}
	}
	hasthesame := false
	for _, oldEnd := range n.oldChildEndpoints {
		if oldEnd == waistedChildEndpoint {
			hasthesame = true
			break
		}
	}
	if !hasthesame {
		n.oldChildEndpoints = append(n.oldChildEndpoints, waistedChildEndpoint)
	}
	return nil
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
	var newEnd string = ""
	var err error
	if leader != nil {
		newEnd, err = leader.GetEndpoint(true)
		if err != nil {
			return err
		}
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
	index := NewStringSet().IndexOf(n.oldLeaderEndpoints, newEnd)
	if index >= 0 && index < len(n.oldLeaderEndpoints)-1 {
		n.oldLeaderEndpoints = append(n.oldLeaderEndpoints[:index], n.oldLeaderEndpoints[index+1:]...)
	}
	n.leader = leader
	if n.leader == nil {
		err = errors.New("no leader after set leader")
	}
	return err
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

func getPublisherEndpoint(endpoint string) string {
	return endpoint
}

func getReceiverEndpoint(endpoint string) (string, error) {
	colon := strings.LastIndexAny(endpoint, ":")
	if colon < 0 || colon >= len(endpoint)-1 {
		return "", errors.New("Invalide endpoint")
	}
	str := endpoint[:colon+1]
	port, err := strconv.Atoi(endpoint[colon+1:])
	if err != nil {
		return "", err
	}
	return str + strconv.Itoa(port+1), nil
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
	str := fmt.Sprintf("NodeInfo[\n\taddr:%v\n\tpath:%v\n\tname:%v\n]", n.addr, n.path, n.GetName())
	return str
}
