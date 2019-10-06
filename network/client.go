package network

import "net"

type Client struct {
	network string
	addr    string
	Agent
}

func NewClient(network, addr string) *Client {
	c := &Client{
		network: network,
		addr:    addr,
	}
	c.handlers = make(map[uint32]func([]byte))
	return c
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
