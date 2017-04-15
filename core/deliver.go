package core

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strings"
)

type Subscriber interface {
	SetSubscribe(filter string) error
	String() string
	Connect() error
	Disconnect(all bool)
	GetConnNodeInfo() []string
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	Close()
	Recv() (Message, error)
}

func NewSubscriber() (Subscriber, error) {
	subscriber, err := NewDeliverer("", zmq.SUB)
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
	GetBindNodeInfo() string
	Close()
	Send(msg Message) error
}

func NewPublisher(addr string) (Publisher, error) {
	return NewDeliverer(addr, zmq.PUB)
}

type Sender interface {
	String() string
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	GetConnNodeInfo() []string
	Connect() error
	Disconnect(all bool)
	Close()
	Send(msg Message) error
}

func NewSender() (Sender, error) {
	return NewDeliverer("", zmq.DEALER)
}

type Receiver interface {
	String() string
	Bind() error
	GetBindNodeInfo() string
	Close()
	Recv() (Message, error)
}

func NewReceiver(addr string) (Receiver, error) {
	return NewDeliverer(addr, zmq.DEALER)
}

type Deliverer interface {
	String() string
	SetSubscribe(filter string) error
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	GetConnNodeInfo() []string
	GetBindNodeInfo() string
	Bind() error
	Connect() error
	Disconnect(all bool)
	Close()
	Send(msg Message) error
	Recv() (Message, error)
}

func NewDeliverer(addr string, t zmq.Type) (Deliverer, error) {
	socket, err := zmq.NewSocket(t)
	if err != nil {
		return nil, err
	}
	return &deliverer{
		bindnode:    addr,
		nowconnodes: make([]string, 0),
		oldconnodes: make([]string, 0),
		socket:      socket,
	}, err
}

type deliverer struct {
	bindnode    string   //NodeInfo
	nowconnodes []string //NodeInfo
	oldconnodes []string //NodeInfo
	socket      *zmq.Socket
}

func (s *deliverer) SetSubscribe(filter string) error {
	return s.socket.SetSubscribe(filter)
}

func (d *deliverer) AddNodeInfo(info NodeInfo) {
	innow := false
	infostr := string(info)
	for _, nownode := range d.nowconnodes {
		if nownode == infostr {
			innow = true
		}
	}
	if !innow {
		d.nowconnodes = append(d.nowconnodes, infostr)
	}

	inold := false
	oldindex := 0
	for index, oldnode := range d.oldconnodes {
		if oldnode == infostr {
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
	infostr := string(info)
	innow := false
	nowindex := 0
	for index, nownode := range d.nowconnodes {
		if nownode == infostr {
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
		if oldnode == infostr {
			inold = true
		}
	}
	if !inold {
		d.oldconnodes = append(d.oldconnodes, infostr)
	}
}

func (d *deliverer) GetConnNodeInfo() []string {
	return d.nowconnodes
}

func (d *deliverer) Bind() error {
	endpoint, err := NodeInfo(d.bindnode).GetEndpoint(true)
	if err != nil {
		return err
	}
	err = d.socket.Bind(endpoint)
	return err
}
func (d *deliverer) GetBindNodeInfo() string {
	return d.bindnode
}
func (d *deliverer) Disconnect(all bool) {
	var oldconnodes []string = make([]string, 0)
	for _, oldnode := range d.oldconnodes {
		oldend, err := NodeInfo(oldnode).GetEndpoint(false)
		if err == nil {
			d.socket.Disconnect(oldend)
		} else {
			oldconnodes = append(oldconnodes, oldnode)
		}
	}
	d.oldconnodes = oldconnodes
	if all {
		var nowconnodes []string = make([]string, 0)
		for _, nownode := range d.nowconnodes {
			newend, err := NodeInfo(nownode).GetEndpoint(false)
			if err == nil {
				d.socket.Disconnect(newend)
			}
		}
		d.nowconnodes = nowconnodes
	}
}

func (d *deliverer) Connect() error {
	d.Disconnect(false)
	var errout error
	for _, nownode := range d.nowconnodes {
		nowend, err := NodeInfo(nownode).GetEndpoint(false)
		if err == nil {
			errout = d.socket.Connect(nowend)
		} else {
			errout = err
		}
	}
	return errout
}

func (d *deliverer) Close() {
	if d.socket != nil {
		d.socket.Close()
	}
}

func (d *deliverer) Send(msg Message) error {
	var err error
	_, buf, err := NewSerializer().Encode(msg.GetInfo())
	if err != nil {
		return err
	}
	buf = append(buf, msg.GetContent()...)
	_, err = d.socket.SendBytes(buf, 0)
	if err != nil {
		return err
	}
	return err
}
func (d *deliverer) Recv() (Message, error) {
	var err error
	buf, err := d.socket.RecvBytes(0)
	if err != nil {
		return nil, err
	}
	msgInfo := NewMessageInfo()
	size, err := NewSerializer().Decode(msgInfo, buf)
	if err != nil {
		return nil, err
	}
	msg := NewMessage(msgInfo, buf[size:])
	return msg, err
}
func (d deliverer) String() string {
	nowcstr := strings.Join(d.nowconnodes, "\n\t\t")
	oldcstr := strings.Join(d.oldconnodes, "\n\t\t")
	return fmt.Sprintf("deliverer[\n\tbind:%v\n\tconnect:\n\t%s\n\twaist:\n\t%s\n]", d.bindnode, nowcstr, oldcstr)
}
