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
	path           core.NodePath
	srvices        map[ServiceType]Service
	clientSrv      *socketio.Server
	clientHandlers []func(so socketio.Socket)
}
type Server interface {
	Start() error
	InitClientHandler() error
	AddClientHandler(handler func(so socketio.Socket))
	ConfigService(st ServiceType, srvice Service)
	SendService(st ServiceType, msg core.Message) error
	GetPath() core.NodePath
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

	info, err := _ParseAddress(b.path, SRT_Undefine, SRT_Client)
	if err != nil {
		return
	}
	err = http.ListenAndServe(info.GetAddress(), nil)
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
func (b *baseServer) GetPath() core.NodePath {
	return b.path
}

func _ParseAddress(npath core.NodePath, fromSRT ServerType, toSRT ServerType) (info core.NodeInfo, err error) {
	info, ok := npath.GetNodeInfo()
	if !ok {
		err = errors.New("Invalid NodePath")
		return
	}
	var addr string
	switch {
	case fromSRT == SRT_Referee && toSRT == SRT_Worker:
		addr = fmt.Sprintf("%s:%d", info.Host, info.Port)
	case fromSRT == SRT_Worker && toSRT == SRT_Referee:
		addr = fmt.Sprintf("%s:%d", info.Host, info.Port)
	case fromSRT == SRT_Worker && toSRT == SRT_Worker:
		addr = fmt.Sprintf("%s:%d", info.Host, info.Port+1)
	case fromSRT == SRT_Workers && toSRT == SRT_Workers:
		addr = fmt.Sprintf("%s:%d", info.Host, info.Port+2)
	case fromSRT == SRT_Undefine && toSRT == SRT_Client:
		addr = fmt.Sprintf("%s:%d", info.Host, info.Port+3)
	}
	info = core.NewNodeInfo()
	err = info.Parse(addr)
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
