package network

import (
	"errors"
	"net"
)

var (
	ErrAgentClosed      error = errors.New("Error Agent Closed")
	ErrEventNotRegister error = errors.New("Error Event Not Register")
)

type Agent struct {
	conn      net.Conn
	handlers  map[uint32]func([]byte)
	conncb    func()
	disconncb func()
	mp        *Processor
}

func NewAgent(conn net.Conn) *Agent {
	a := &Agent{
		conn:     conn,
		mp:       NewProcessor(conn),
		handlers: make(map[uint32]func([]byte)),
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

func (a *Agent) Processor() *Processor {
	return a.mp
}

func (a *Agent) emit(head uint32, body []byte) {
	h, ok := a.handlers[head]
	if ok {
		h(body)
	}
}

func (a *Agent) handle() (err error) {
	if a.mp == nil {
		err = ErrAgentClosed
		return
	}
	head, body, err := a.mp.Read()
	if err != nil {
		return
	}
	a.emit(head, body)
	return
}

func (a *Agent) Send(head uint32, body []byte) (err error) {
	if a.mp == nil {
		err = ErrAgentClosed
		return
	}
	err = a.mp.Write(head, body)
	return
}

func (a *Agent) On(head uint32, cb func([]byte)) {
	a.handlers[head] = cb
}

func (a *Agent) OnConnect(cb func()) {
	a.conncb = cb
}
func (a *Agent) OnDisconnect(cb func()) {
	a.disconncb = cb
}
