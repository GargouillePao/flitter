package servers

import (
	"fmt"
	core "github.com/gargous/flitter/core"
	utils "github.com/gargous/flitter/utils"
	socketio "github.com/googollee/go-socket.io"
)

type ScenceService interface {
	Service
}

type scencesrvice struct {
	worker     Worker
	accessable bool
	baseService
	clients map[string]socketio.Socket
}

func NewScenceService() ScenceService {
	service := &scencesrvice{}
	service.accessable = false
	service.clients = make(map[string]socketio.Socket)
	service.looper = core.NewMessageLooper(__LooperSize)
	return service
}
func (s *scencesrvice) Init(srv interface{}) error {
	s.worker = srv.(Worker)
	s.HandleMessages()
	s.HandleClients()
	return nil
}
func (s *scencesrvice) HandleClients() {
	s.worker.AddClientHandler(func(so socketio.Socket) {
		utils.Logf(utils.Norf, "Client Conneted")
		so.On("data", func(data string) {
			utils.Logf(utils.Norf, "Client say:%v", data)
			if s.accessable {
				so.Emit("data", "Hi~ [reply your %v]", data)
			}
		})
		so.On("enter", func(name string) {
			utils.Logf(utils.Norf, "Client %v want to enter", name)
			so.Emit("enter", s.accessable)
			if s.accessable {
				utils.Logf(utils.Infof, "Client %v entered ", name)
				s.clients[name] = so
			} else {
				utils.Logf(utils.Warningf, "Client %v should wait", name)
			}
		})
	})
}
func (s *scencesrvice) HandleMessages() {
	s.looper.AddHandler(0, core.MA_Heartbeat, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Succeed:
			s.accessable = true
		case core.MS_Failed:
			s.accessable = false
		}
		return
	})
}
func (s *scencesrvice) Start() {
	s.looper.Loop()
}
func (s *scencesrvice) Term() {
	s.Term()
}
func (s scencesrvice) String() string {
	str := fmt.Sprintf("Heartbeat Service:["+"looper:%p"+"]", s.looper)
	return str
}
