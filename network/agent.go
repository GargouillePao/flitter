package network

import (
	"errors"
	"log"
	"net"

	"github.com/gogo/protobuf/proto"
)

var (
	ErrAgentClosed      error = errors.New("Error Agent Closed")
	ErrEventNotRegister error = errors.New("Error Event Not Register")
	ErrCantReceive      error = errors.New("Error Message Can Not Recv")
)

type Agent struct {
	conn      net.Conn
	conncb    func()
	disconncb func()
	errcb     func(error)
	mp        *Processor
}

func NewAgent(conn net.Conn) *Agent {
	a := &Agent{
		conn:      conn,
		mp:        NewProcessor(conn),
		conncb:    func() {},
		disconncb: func() {},
		errcb: func(e error) {
			log.Println(e)
		},
	}
	return a
}

func (a *Agent) Close() {
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
		a.disconncb()
	}
}

func (a *Agent) MP() *Processor {
	return a.mp
}

func (a *Agent) handle() (err error) {
	if a.mp == nil {
		err = ErrAgentClosed
		return
	}
	err = a.mp.handle()
	if err != nil {
		return
	}
	return
}

func (a *Agent) Send(msg proto.Message) (err error) {
	if a.mp == nil {
		err = ErrAgentClosed
		return
	}
	err = a.mp.send(msg)
	return
}

func (a *Agent) OnConnect(cb func()) {
	a.conncb = cb
}
func (a *Agent) OnDisconnect(cb func()) {
	a.disconncb = cb
}
