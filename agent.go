package flitter

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"sync"
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
	SendPeer(peer string, head uint32, body proto.Message) error
	Send(tokenID uint64, head uint32, body proto.Message) error
	Register(headId uint32, act func(AgentConn, interface{}) error)
	AddPeer(name, addr string, mp MsgProcesser) error
}

type agent struct {
	addr    string
	clients sync.Map
	server
}

func NewAgent(addr string, c map[uint32]func() proto.Message, dealCnt int32, clientWait time.Duration, serverWait time.Duration) Agent {
	prcser := NewMsgProcesser(c, true)
	ag := &agent{
		addr:   addr,
		server: newServer(prcser, dealCnt, clientWait, serverWait),
	}
	return ag
}

func (a *agent) Register(headId uint32, act func(AgentConn, interface{}) error) {
	a.mp.Register(headId, func(d *dealer, msg interface{}) error {
		conn, ok := a.clients.Load(d.id)
		if !ok {
			return errors.New(fmt.Sprintf("Dealer %d Not Linked With AgentConn", d.id))
		}
		return act(conn.(AgentConn), msg)
	})
}

func (a *agent) Start() {
	a.onconnect = func(d *dealer) {
		for i := 0; i < 20; i++ {
			tokenID := rand.Uint64()
			if _, ok := a.clients.Load(tokenID); !ok {
				d.id = tokenID
				break
			}
		}
		a.clients.Store(d.id, AgentConn(&agentConn{
			d:     d,
			token: d.id,
		}))
	}
	a.ondisconnect = func(d *dealer) {
		a.clients.Delete(d.id)
	}
	a.serve(a.addr)
	return
}

func (a *agent) Send(tokenID uint64, head uint32, body proto.Message) (err error) {
	c, ok := a.clients.Load(tokenID)
	if !ok {
		err = errors.New(fmt.Sprintf("Client %d Not Find ", tokenID))
		return
	}
	err = c.(AgentConn).Reply(head, body)
	return
}
