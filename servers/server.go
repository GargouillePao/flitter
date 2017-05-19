package servers

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter/common"
	"github.com/gargous/flitter/core"
	socketio "github.com/googollee/go-socket.io"
	"net/http"
)

type baseServer struct {
	common.BaseDataSet
	srvices        map[ServiceType]Service
	clientSrv      *socketio.Server
	clientSessions map[string]socketio.Socket
	clientHandlers map[string](func(so socketio.Socket) interface{})
}

type Server interface {
	putDataToMessage(msg core.Message, keys ...string) (ok bool)
	Start() error
	InitClientHandler(cb func()) error
	OnClient(event string, handler func(so socketio.Socket) interface{})
	ConfigService(st ServiceType, srvice Service)
	SendService(st ServiceType, msg core.Message) error
	GetClientSocket() *socketio.Server
}

func (s *baseServer) putDataToMessage(msg core.Message, keys ...string) (ok bool) {
	for _, key := range keys {
		item, ok := s.Get(key)
		if !ok {
			return ok
		}
		msg.AppendContent(item.Data)
	}
	ok = true
	return
}

func (b *baseServer) GetClientSocket() *socketio.Server {
	return b.clientSrv
}
func (b *baseServer) OnClient(event string, handler func(so socketio.Socket) interface{}) {
	if b.clientHandlers == nil {
		b.clientHandlers = make(map[string]func(so socketio.Socket) interface{})
	}
	b.clientHandlers[event] = handler
}

func (b *baseServer) InitClientHandler(cb func()) (err error) {
	if b.clientHandlers == nil {
		b.clientHandlers = make(map[string]func(so socketio.Socket) interface{})
	}
	b.clientSessions = make(map[string]socketio.Socket)
	b.clientSrv, err = socketio.NewServer(nil)
	if err != nil {
		err = common.ErrAppend(err, "New SocketIO")
		return
	}
	err = b.clientSrv.On("error", func(so socketio.Socket, err error) {
		common.ErrIn(err, so.Id(), "Client")
	})
	if err != nil {
		err = common.ErrAppend(err, "Client On Error")
		return
	}
	err = b.clientSrv.On("connection", func(so socketio.Socket) {

		_, ok := b.clientSessions[so.Id()]
		if !ok {
			b.clientSessions[so.Id()] = so
		} else {
			return
		}
		err = so.Join("flitter")
		if err != nil {
			common.ErrIn(err, "Client When Join Flitter")
			return
		}
		common.Logf(common.Infof, "Client Connected")
		err = so.On("disconnection", func() {
			common.Logf(common.Warningf, "Client Disconnected")
		})
		if err != nil {
			common.ErrIn(err, "Client On Disconnection")
			return
		}
		for event, handler := range b.clientHandlers {
			cb := handler(so)
			err = so.On(event, cb)
			if err != nil {
				common.ErrIn(err, "Client On "+event)
				return
			}
		}
	})
	if err != nil {
		err = common.ErrAppend(err, "Client On Connect")
		return
	}
	http.Handle("/socket.io/", b.clientSrv)

	npath, ok := b.Get("path")
	if !ok {
		err = errors.New("Path Not Exist")
	}
	info, err := _ParseAddress(core.NodePath(npath.Data), SRT_Undefine, SRT_Client)
	if err != nil {
		err = common.ErrAppend(err, "Parse Address")
		return
	}
	common.Logf(common.Norf, "Initiate Clients Handler")
	if cb != nil {
		cb()
	}
	err = http.ListenAndServe(info.GetAddress(), nil)
	if err != nil {
		err = common.ErrAppend(err, "HTTP Listen And Serve")
	}
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
