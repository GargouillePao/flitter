package flitter

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"time"
)

type AgentConn interface {
	Reply(head uint32, body proto.Message) error
	Close() error
	Token() uint64
	SetID(uint64)
	ID() uint64
}

type agentConn struct {
	d     *dealer
	id    uint64
	token uint64
}

func (ac *agentConn) Reply(head uint32, body proto.Message) error {
	return ac.d.Send(head, body, true)
}

func (ac *agentConn) Close() error {
	return ac.d.Close()
}

func (ac *agentConn) Token() uint64 {
	return ac.token
}

func (ac *agentConn) SetID(id uint64) {
	ac.id = id
}

func (ac *agentConn) ID() uint64 {
	return ac.id
}

type Agent interface {
	Start()
	To(peer string, head uint32, body proto.Message) error
	Send(tokenID uint64, head uint32, body proto.Message) error
	Register(headId uint32, act func(AgentConn, interface{}) error)
	AddPeer(name, addr string, mp MsgProcesser)
}

type agent struct {
	addr    string
	ds      map[string]*dealer
	clients map[uint64]*agentConn
	srv     *server
}

func NewAgent(addr string, c map[uint32]func() proto.Message, dealCnt int32, readWait time.Duration) Agent {
	prcser := NewMsgProcesser(c, true)
	return &agent{
		addr:    addr,
		ds:      make(map[string]*dealer),
		clients: make(map[uint64]*agentConn),
		srv:     newServer(prcser, dealCnt, readWait),
	}
}

func (a *agent) AddPeer(name, addr string, mp MsgProcesser) {
	d := newDealer(nil)
	d.addr = addr
	go func() {
		err := d.Process(mp)
		if err != nil {
			a.srv.mp.handleErr(d, err)
		}
	}()
	a.ds[name] = d
	return
}

func (a *agent) Register(headId uint32, act func(AgentConn, interface{}) error) {
	a.srv.mp.Register(headId, func(d *dealer, msg interface{}) error {
		conn, ok := a.clients[d.id]
		if !ok {
			return errors.New(fmt.Sprintf("Dealer %d Not Linked With AgentConn", d.id))
		}
		return act(conn, msg)
	})
}

func (a *agent) Start() {
	a.srv.onconnect = func(d *dealer) {
		for i := 0; i < 20; i++ {
			tokenID := rand.Uint64()
			if _, ok := a.clients[tokenID]; !ok {
				d.id = tokenID
				break
			}
		}
		a.clients[d.id] = &agentConn{
			d:     d,
			token: d.id,
		}
	}
	a.srv.ondisconnect = func(d *dealer) {
		a.clients[d.id] = nil
	}
	a.srv.Serve(a.addr)
	return
}

func (a *agent) To(peer string, head uint32, body proto.Message) (err error) {
	d, ok := a.ds[peer]
	if !ok {
		err = errors.New(fmt.Sprintf("Peer %s Not Found", peer))
		return
	}
	err = d.Connect()
	if err != nil {
		return
	}
	err = d.Send(head, body, false)
	return
}

func (a *agent) Send(tokenID uint64, head uint32, body proto.Message) (err error) {
	c, ok := a.clients[tokenID]
	if !ok {
		err = errors.New(fmt.Sprintf("Client %d Not Find ", tokenID))
		return
	}
	err = c.d.Send(head, body, true)
	return
}
