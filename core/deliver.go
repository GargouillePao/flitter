package core

import (
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

type Subscriber interface {
	SetSubscribe(filter string) error
	String() string
	Connect() error
	Disconnect(all bool)
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	GetConnNodeInfo() []NodeInfo
	Close()
	Recv() (Message, error)
}

func NewSubscriber() (Subscriber, error) {
	subscriber, err := NewDeliverer(NewNodeInfo(), zmq.SUB)
	if err != nil {
		return nil, err
	}
	err = subscriber.SetSubscribe("")
	if err != nil {
		return nil, err
	}
	return subscriber, nil
}

type Publisher interface {
	String() string
	Bind() error
	GetBindNodeInfo() NodeInfo
	Close()
	Send(msg Message) error
}

func NewPublisher(info NodeInfo) (Publisher, error) {
	return NewDeliverer(info, zmq.PUB)
}

type Sender interface {
	String() string
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	GetConnNodeInfo() []NodeInfo
	Connect() error
	Disconnect(all bool)
	Close()
	Send(msg Message) error
}

func NewSender() (Sender, error) {
	return NewDeliverer(NewNodeInfo(), zmq.DEALER)
}

type Receiver interface {
	String() string
	Bind() error
	GetBindNodeInfo() NodeInfo
	Close()
	Recv() (Message, error)
}

func NewReceiver(info NodeInfo) (Receiver, error) {
	return NewDeliverer(info, zmq.DEALER)
}

type Deliverer interface {
	String() string
	SetSubscribe(filter string) error
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	GetConnNodeInfo() []NodeInfo
	GetBindNodeInfo() NodeInfo
	Bind() error
	Connect() error
	Disconnect(all bool)
	Close()
	Send(msg Message) error
	Recv() (Message, error)
}

func NewDeliverer(info NodeInfo, t zmq.Type) (Deliverer, error) {
	socket, err := zmq.NewSocket(t)
	if err != nil {
		return nil, err
	}
	return &deliverer{
		bindnode:    info,
		nowconnodes: make([]NodeInfo, 0),
		oldconnodes: make([]NodeInfo, 0),
		socket:      socket,
	}, err
}

type deliverer struct {
	bindnode    NodeInfo
	nowconnodes []NodeInfo
	oldconnodes []NodeInfo
	socket      *zmq.Socket
}

func (s *deliverer) SetSubscribe(filter string) error {
	return s.socket.SetSubscribe(filter)
}

func (d *deliverer) AddNodeInfo(info NodeInfo) {
	innow := false
	for _, nownode := range d.nowconnodes {
		if nownode == info {
			innow = true
		}
	}
	if !innow {
		d.nowconnodes = append(d.nowconnodes, info)
	}

	inold := false
	oldindex := 0
	for index, oldnode := range d.oldconnodes {
		if oldnode == info {
			inold = true
			oldindex = index
		}
	}
	if inold {
		if oldindex >= len(d.oldconnodes)-1 {
			d.oldconnodes = d.oldconnodes[:oldindex]
		} else {
			d.oldconnodes = append(d.oldconnodes[:oldindex], d.oldconnodes[oldindex+1:]...)
		}
	}
	return
}

func (d *deliverer) RemoveNodeInfo(info NodeInfo) {
	innow := false
	nowindex := 0
	for index, nownode := range d.nowconnodes {
		if nownode == info {
			innow = true
			nowindex = index
		}
	}
	if innow {
		if nowindex >= len(d.nowconnodes)-1 {
			d.nowconnodes = d.nowconnodes[:nowindex]
		} else {
			d.nowconnodes = append(d.nowconnodes[:nowindex], d.nowconnodes[nowindex+1:]...)
		}
	}
	inold := false
	for _, oldnode := range d.oldconnodes {
		if oldnode == info {
			inold = true
		}
	}
	if !inold {
		d.oldconnodes = append(d.oldconnodes, info)
	}
}

func (d *deliverer) GetConnNodeInfo() []NodeInfo {
	return d.nowconnodes
}

func (d *deliverer) Bind() error {
	endpoint := d.bindnode.GetEndpoint(true)
	return d.socket.Bind(endpoint)
}
func (d *deliverer) GetBindNodeInfo() NodeInfo {
	return d.bindnode
}
func (d *deliverer) Disconnect(all bool) {
	var oldconnodes []NodeInfo = make([]NodeInfo, 0)
	for _, oldnode := range d.oldconnodes {
		oldend := oldnode.GetEndpoint(false)
		d.socket.Disconnect(oldend)
	}
	d.oldconnodes = oldconnodes
	if all {
		var nowconnodes []NodeInfo = make([]NodeInfo, 0)
		for _, nownode := range d.nowconnodes {
			newend := nownode.GetEndpoint(false)
			d.socket.Disconnect(newend)
		}
		d.nowconnodes = nowconnodes
	}
}

func (d *deliverer) Connect() (err error) {
	d.Disconnect(false)
	for _, nownode := range d.nowconnodes {
		nowend := nownode.GetEndpoint(false)
		err = d.socket.Connect(nowend)
		if err != nil {
			return
		}
	}
	return
}

func (d *deliverer) Close() {
	if d.socket != nil {
		d.socket.Close()
	}
}

func (d *deliverer) Send(msg Message) (err error) {
	_, buf, err := NewSerializer().Encode(msg.GetInfo())
	if err != nil {
		return err
	}
	if msg.GetContents() == nil || len(msg.GetContents()) == 0 {
		_, err = d.socket.SendBytes(buf, 0)
	} else {
		_, err = d.socket.SendBytes(buf, zmq.SNDMORE)
		if err != nil {
			return
		}
		for index, content := range msg.GetContents() {
			if index == len(msg.GetContents())-1 {
				_, err = d.socket.SendBytes(content, 0)
			} else {
				_, err = d.socket.SendBytes(content, zmq.SNDMORE)
			}
			if err != nil {
				return
			}
		}
	}
	return
}
func (d *deliverer) Recv() (msg Message, err error) {
	bufs, err := d.socket.RecvMessageBytes(0)
	if err != nil {
		return
	}
	if len(bufs) < 1 {
		err = errors.New("No Info In This Message")
	}
	msgInfo := NewMessageInfo()
	_, err = NewSerializer().Decode(msgInfo, bufs[0])
	if err != nil {
		return
	}
	msg = NewMessage(msgInfo)
	msg.SetContents(bufs[1:])
	return
}
func (d deliverer) String() string {
	nowcstr := ""
	oldcstr := ""
	for _, nownodes := range d.nowconnodes {
		nowcstr += fmt.Sprintf("%d\n\t\t", nownodes)
	}
	for _, oldnodes := range d.nowconnodes {
		oldcstr += fmt.Sprintf("%d\n\t\t", oldnodes)
	}
	return fmt.Sprintf("deliverer[\n\tbind:%v\n\tconnect:\n\t%s\n\twaist:\n\t%s\n]", d.bindnode, nowcstr, oldcstr)
}
