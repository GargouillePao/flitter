package servers

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/utils"
	socketio "github.com/googollee/go-socket.io"
	"net/http"
)

type baseServer struct {
	addr           string
	name           string
	srvices        map[ServiceType]Service
	clientSrv      *socketio.Server
	clientHandlers []func(so socketio.Socket)
}
type Server interface {
	Start() error
	GetAddress() string
	InitClientHandler() error
	AddClientHandler(handler func(so socketio.Socket))
	ConfigService(st ServiceType, srvice Service)
	SendService(st ServiceType, msg core.Message) error
}

func (b *baseServer) AddClientHandler(handler func(so socketio.Socket)) {
	b.clientHandlers = append(b.clientHandlers, handler)
}

func (b *baseServer) InitClientHandler() (err error) {
	if b.clientHandlers != nil {
		b.clientHandlers = make([]func(so socketio.Socket), 0)
	}
	b.clientSrv, err = socketio.NewServer(nil)
	if err != nil {
		return
	}
	err = b.clientSrv.On("error", func(so socketio.Socket, err error) {
		utils.ErrIn(err, so.Id(), "Client")
	})
	if err != nil {
		return
	}

	err = b.clientSrv.On("connection", func(so socketio.Socket) {
		utils.Logf(utils.Infof, "Client Connected")
		so.On("disconnection", func() {
			utils.Logf(utils.Warningf, "Client Disconnected")
		})
		for _, handler := range b.clientHandlers {
			handler(so)
		}
	})
	if err != nil {
		return
	}
	http.Handle("/socket.io/", b.clientSrv)
	addr, err := _ParseAddress(core.NodeInfo(b.GetAddress()), SRT_Undefine, SRT_Client)
	if err != nil {
		return
	}
	err = http.ListenAndServe(addr, nil)
	return
}
func (b *baseServer) ConfigService(st ServiceType, srvice Service) {
	b.srvices[st] = srvice
}
func (b *baseServer) SendService(st ServiceType, msg core.Message) (err error) {
	srvice, ok := b.srvices[st]
	if ok {
		srvice.Push(msg)
	} else {
		err = errors.New("service " + st.String() + " hasnt config")
	}
	return
}
func (b *baseServer) GetAddress() (addr string) {
	return b.addr
}

func _ParseAddress(info core.NodeInfo, fromSRT ServerType, toSRT ServerType) (addr string, err error) {
	host, err := info.GetHost()
	if err != nil {
		return
	}
	port, err := info.GetPort()
	if err != nil {
		return
	}
	switch {
	case fromSRT == SRT_Referee && toSRT == SRT_Worker:
		addr = fmt.Sprintf("%s:%d", host, port)
	case fromSRT == SRT_Worker && toSRT == SRT_Referee:
		addr = fmt.Sprintf("%s:%d", host, port)
	case fromSRT == SRT_Worker && toSRT == SRT_Worker:
		addr = fmt.Sprintf("%s:%d", host, port+1)
	case fromSRT == SRT_Workers && toSRT == SRT_Workers:
		addr = fmt.Sprintf("%s:%d", host, port+2)
	case fromSRT == SRT_Undefine && toSRT == SRT_Client:
		addr = fmt.Sprintf("%s:%d", host, port+3)
	}
	return
}

type ServerType uint8

const (
	_ ServerType = iota
	SRT_Undefine
	SRT_Referee
	SRT_Worker
	SRT_Workers
	SRT_Client
)

func (s ServerType) String() string {
	switch s {
	case SRT_Undefine:
		return "Undefine"
	case SRT_Referee:
		return "Referee"
	case SRT_Worker:
		return "Worker"
	case SRT_Workers:
		return "Workers"
	case SRT_Client:
		return "Client"
	}
	return ""
}
