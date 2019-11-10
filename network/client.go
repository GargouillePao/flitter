package network

import (
	"net"
	"reflect"
	"sync"

	"github.com/gogo/protobuf/proto"
)

type Client struct {
	network string
	addr    string
	Agent
	mutex sync.Mutex
}

func NewClient(network, addr string) *Client {
	c := &Client{
		network: network,
		addr:    addr,
	}
	return c
}

func (c *Client) Req(msg proto.Message, cb func(proto.Message)) error {
	cbType := reflect.TypeOf(cb)
	mp := c.Agent.mp
	c.mutex.Lock()
	ack := mp.GetIdByName(cbType.In(0).Name())
	ok := mp.onNotify(ack, cb)
	if !ok {
		return ErrMsgNotRegistered
	}
	c.mutex.Unlock()
	err := c.Agent.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Start() (err error) {
	c.conn, err = net.Dial(c.network, c.addr)
	if err != nil {
		return
	}
	defer c.Close()
	c.mp = NewProcessor(c.conn)
	c.conncb()
	for {
		err = c.handle()
		if err != nil {
			return
		}
	}
}
