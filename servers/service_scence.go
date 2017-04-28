package servers

import (
	"errors"
	"fmt"
	core "github.com/gargous/flitter/core"
	utils "github.com/gargous/flitter/utils"
	socketio "github.com/googollee/go-socket.io"
	"strings"
	"time"
)

type ScenceService interface {
	Service
	SetClientData(name string, key string, value []byte) (err error)
	GetClientData(name string, key string) (values map[string][]byte)
	IsAccess() bool
}

type scencesrvice struct {
	worker     Worker
	accessable bool
	baseService
	clients map[string]DataSet
}

func NewScenceService() ScenceService {
	service := &scencesrvice{}
	service.accessable = false
	service.clients = make(map[string]DataSet)
	service.looper = core.NewMessageLooper(__LooperSize)
	return service
}
func (s *scencesrvice) IsAccess() bool {
	return s.accessable
}
func (s *scencesrvice) Init(srv interface{}) error {
	s.worker = srv.(Worker)
	s.HandleMessages()
	s.HandleClients()
	return nil
}
func (s *scencesrvice) parseClientName(hostname string, name string) (targename string, ok bool) {
	workernameb, ok := s.worker.Get("name")
	if !ok {
		return
	}
	workername := string(workernameb)
	if hostname == workername {
		targename = name
	} else {
		targename = fmt.Sprintf("%s|%d|%s", workername, time.Now().UnixNano(), name)
	}
	ok = true
	return
}
func (s *scencesrvice) parseClientData(msg core.Message) (name string, key string, value []byte, hostpath core.NodePath, err error) {
	hostpathb, ok := msg.GetContent(3)
	if !ok {
		err = errors.New("Invalid Content")
		return
	}
	hostpath = core.NodePath(hostpathb)
	name = string(msg.GetContents()[0])
	key = string(msg.GetContents()[1])
	value = msg.GetContents()[2]
	return
}
func (s *scencesrvice) setClientData(name string, key string, value []byte) {
	if s.clients == nil {
		s.clients = make(map[string]DataSet)
	}
	targetclient, ok_ := s.clients[name]
	if !ok_ || targetclient == nil {
		targetclient = &dataSet{}
	}
	targetclient.Set(key, value)
	s.clients[name] = targetclient
	return
}

func (s *scencesrvice) updateClientData(name string, key string, value []byte) (err error) {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Update)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	msg.AppendContent([]byte(name))
	msg.AppendContent([]byte(key))
	msg.AppendContent(value)
	mypathi, ok := s.worker.Get("path")
	if !ok {
		err = errors.New("Get Path")
		return
	}
	mypath := core.NodePath(mypathi)
	lpath, ok := mypath.GetLeaderPath()
	if !ok || lpath == "" {
		return
	}
	msg.AppendContent([]byte(mypath))
	s.looper.Push(msg)
	return
}
func (s *scencesrvice) SetClientData(name string, key string, value []byte) (err error) {
	s.setClientData(name, key, value)
	err = s.updateClientData(name, key, value)
	return
}
func (s *scencesrvice) GetClientDatas(name string) map[string]DataSet {
	if name == "" {
		return s.clients
	} else {
		return map[string]DataSet{name: s.clients[name]}
	}
}
func (s *scencesrvice) GetClientData(name string, key string) (values map[string][]byte) {
	sdataset := s.GetClientDatas(name)
	values = make(map[string][]byte)
	for name, data := range sdataset {
		value, ok := data.Get(key)
		if ok {
			values[name] = value
		}
	}
	return
}
func (s *scencesrvice) SetServerData(key string, value []byte) {
	s.worker.Set(key, value)
}
func (s *scencesrvice) HandleClients() {
	s.worker.TrickClient("flitter enter", func(so socketio.Socket) interface{} {
		return func(name string, hostname string) {
			targename, ok := s.parseClientName(hostname, name)
			if ok {
				utils.Logf(utils.Norf, "Client %v want to enter", targename)
				if s.accessable {
					so.Emit("flitter enter", targename)
					utils.Logf(utils.Infof, "Client %v entered ", targename)
					soidbyte := []byte(so.Id())
					s.setClientData(targename, "socket", soidbyte)
					clientsbyte, err := utils.GobEcode(s.clients)
					if err != nil {
						utils.ErrIn(err)
						return
					}
					s.SetServerData("clients", clientsbyte)
				} else {
					so.Emit("flitter enter", __Client_Reply_bussy)
					utils.Logf(utils.Warningf, "Client %v should wait", targename)
				}
			}
		}
	})
}
func (s *scencesrvice) HandleMessages() {
	s.looper.AddHandler(0, core.MA_Init, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			s.accessable = true
		}
		return
	})
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
	s.looper.AddHandler(3000, core.MA_Update, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			npathi, ok := s.worker.Get("path")
			if !ok {
				err = errors.New("No Path")
				return
			}
			npath := core.NodePath(npathi)
			if npath == "" {
				err = errors.New("Path Is Not A String")
				return
			}
			lpath, ok := core.NodePath(npath).GetLeaderPath()
			if !ok || lpath == "" {
				err = errors.New("No Leader In Path")
				return
			}
			msg.GetInfo().SetState(core.MS_Ask)
			err = s.worker.SendToWroker(msg, lpath)
		case core.MS_Ask:
			cname, cdkey, cdvalue, hostpath, err := s.parseClientData(msg)
			if err != nil {
				return err
			}
			mypathi, ok := s.worker.Get("path")
			if ok {
				mypath := core.NodePath(mypathi)
				if mypath == hostpath {
					return err
				}
			}
			msg.GetInfo().SetState(core.MS_Succeed)
			err = s.worker.PublishToWorker(msg)
			if err != nil {
				return err
			}
			err = s.SetClientData(cname, cdkey, cdvalue)
			if err != nil {
				return err
			}
			utils.Logf(utils.Infof, "My Clients:\n%v", s.ClientsString())
		case core.MS_Succeed:
			cname, cdkey, cdvalue, hostpath, err := s.parseClientData(msg)
			if err != nil {
				return err
			}
			mypathi, ok := s.worker.Get("path")
			if ok {
				mypath := core.NodePath(mypathi)
				if mypath == hostpath {
					return err
				}
			}
			s.setClientData(cname, cdkey, cdvalue)
			utils.Logf(utils.Infof, "OK My Clients:\n%v", s.clients)
		case core.MS_Failed:
			utils.Logf(utils.Warningf, "Update Faild And Now %v", msg)
			msg.GetInfo().SetTime(time.Now())
			msg.GetInfo().SetState(core.MS_Probe)
			s.looper.Push(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.String()))
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
func (s scencesrvice) ClientsString() (str string) {
	str = ""
	for name, data := range s.clients {
		ndatastr := strings.Join(strings.Split(fmt.Sprintf("%v", data), "\n"), "\n\t")
		str += fmt.Sprintf("name:%s\ndata:[\n\t%s\n]\n", name, ndatastr)
	}
	return
}
func (s scencesrvice) String() string {
	str := fmt.Sprintf("Heartbeat Service:\n\tlooper:%p\n\tclients:%s", s.looper, s.ClientsString())
	return str
}