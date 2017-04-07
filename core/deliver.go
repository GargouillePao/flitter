package core

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strings"
)

type Subscriber interface {
	String() string
	Connect() error
	Disconnect(all bool)
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	Recv() (Message, error)
}

func NewSubscriber() (Subscriber, error) {
	return NewDeliverer("", zmq.SUB)
}

type Publisher interface {
	String() string
	Bind() error
	Send(msg Message) error
}

func NewPublisher(info NodeInfo) (Publisher, error) {
	return NewDeliverer(info, zmq.PUB)
}

type Sender interface {
	String() string
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	Connect() error
	Disconnect(all bool)
	Send(msg Message) error
}

func NewSender() (Sender, error) {
	return NewDeliverer("", zmq.DEALER)
}

type Receiver interface {
	String() string
	Bind() error
	Recv() (Message, error)
}

func NewReceiver(info NodeInfo) (Receiver, error) {
	return NewDeliverer(info, zmq.DEALER)
}

type Deliverer interface {
	String() string
	AddNodeInfo(info NodeInfo)
	RemoveNodeInfo(info NodeInfo)
	GetConnNodeInfo() []string
	Bind() error
	Connect() error
	Disconnect(all bool)
	Send(msg Message) error
	Recv() (Message, error)
}

func NewDeliverer(info NodeInfo, t zmq.Type) (Deliverer, error) {
	socket, err := zmq.NewSocket(t)
	if err != nil {
		return nil, err
	}
	endpoint, _ := info.GetEndpoint()
	return &deliverer{
		bindend:     endpoint,
		nowconnodes: make([]string, 0),
		oldconnodes: make([]string, 0),
		socket:      socket,
	}, err
}

type deliverer struct {
	bindend     string
	nowconnodes []string //NodeInfo
	oldconnodes []string //NodeInfo
	socket      *zmq.Socket
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
	return d.socket.Bind(d.bindend)
}

func (d *deliverer) Disconnect(all bool) {
	for _, oldnode := range d.oldconnodes {
		oldend, err := NodeInfo(oldnode).GetEndpoint()
		if err == nil {
			d.socket.Disconnect(oldend)
		}
	}
	if all {
		for _, nownode := range d.nowconnodes {
			newend, err := NodeInfo(nownode).GetEndpoint()
			if err == nil {
				d.socket.Disconnect(newend)
			}
		}
	}
}

func (d *deliverer) Connect() error {
	d.Disconnect(false)
	var errout error
	for _, nownode := range d.nowconnodes {
		nowend, err := NodeInfo(nownode).GetEndpoint()
		if err == nil {
			errout = d.socket.Connect(nowend)
		} else {
			errout = err
		}
	}
	return errout
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
	return fmt.Sprintf("deliverer[\n\tbind:%v\n\tconnect:\n\t%s\n\twaist:\n\t%s\n]", d.bindend, nowcstr, oldcstr)
}
