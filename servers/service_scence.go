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
	SetClientData(name string, key string, value DataItem) (err error)
	GetClientData(name string, key string) (values map[string]DataItem)
	OnScenceDataUpdate(cdkey string, cb func(cname string, cdkey string, cdvalue DataItem, hostpath core.NodePath))
	IsAccess() bool
}

type scencesrvice struct {
	worker     Worker
	accessable bool
	baseService
	clientsDataUpdateHandlers map[string]func(cname string, cdkey string, cdvalue DataItem, hostpath core.NodePath)
	clients                   map[string]DataSet
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
	workerpathi, ok := s.worker.Get("path")
	if !ok {
		return
	}
	info, ok := core.NodePath(workerpathi.Data).GetNodeInfo()
	if !ok {
		return
	}
	workername := info.Name
	if hostname == workername {
		targename = name
	} else {
		targename = fmt.Sprintf("%s|%d|%s", workername, time.Now().UnixNano(), name)
	}
	ok = true
	return
}
func (s *scencesrvice) parseClientData(msg core.Message) (name string, key string, value DataItem, hostpath core.NodePath, err error) {
	hostpathb, ok := msg.GetContent(3)
	if !ok {
		err = errors.New("Invalid Content")
		return
	}
	hostpath = core.NodePath(hostpathb)
	name = string(msg.GetContents()[0])
	key = string(msg.GetContents()[1])
	err = value.Parse(msg.GetContents()[2])
	return
}
func (s *scencesrvice) setClientData(name string, key string, value DataItem) {
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
func (s *scencesrvice) updateClientData(name string, key string, value DataItem) (err error) {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Update)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	msg.AppendContent([]byte(name))
	msg.AppendContent([]byte(key))
	b, err := value.Bytes()
	if err != nil {
		return
	}
	msg.AppendContent(b)
	mypathi, ok := s.worker.Get("path")
	if !ok {
		err = errors.New("Get Path")
		return
	}
	mypath := core.NodePath(mypathi.Data)
	lpath, ok := mypath.GetLeaderPath()
	if !ok || lpath == "" {
		return
	}
	msg.AppendContent([]byte(mypath))
	s.looper.Push(msg)
	return
}
func (s *scencesrvice) updateClientDataVersion(name string, key string, value *DataItem) {
	data, ok := s.clients[name]
	if ok {
		oldval, ok := data.Get(key)
		if ok {
			value.Version = oldval.Version + 1
		}
	}

	return
}
func (s *scencesrvice) SetClientData(name string, key string, value DataItem) (err error) {
	s.updateClientDataVersion(name, key, &value)
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
func (s *scencesrvice) GetClientData(name string, key string) (values map[string]DataItem) {
	sdataset := s.GetClientDatas(name)
	values = make(map[string]DataItem)
	for name, data := range sdataset {
		value, ok := data.Get(key)
		if ok {
			values[name] = value
		}
	}
	return
}
func (s *scencesrvice) SetServerData(key string, value DataItem) {
	s.worker.Set(key, value)
}
func (s *scencesrvice) OnScenceDataUpdate(cdkey string, cb func(cname string, cdkey string, cdvalue DataItem, hostpath core.NodePath)) {
	if s.clientsDataUpdateHandlers == nil {
		s.clientsDataUpdateHandlers = make(map[string]func(cname string, cdkey string, cdvalue DataItem, hostpath core.NodePath))
	}
	s.clientsDataUpdateHandlers[cdkey] = cb
}
func (s *scencesrvice) HandleClients() {
	s.worker.TrickClient("flitter enter", func(so socketio.Socket) interface{} {
		return func(name string, hostname string) {
			var soidData DataItem
			err := soidData.Parse(so.Id())
			if err != nil {
				utils.ErrIn(err)
				return
			}
			targename, ok := s.parseClientName(hostname, name)
			if !ok {
				utils.ErrIn(errors.New("Parse Client Error"))
				return
			}
			utils.Logf(utils.Norf, "Client %v want to enter", targename)

			if !s.accessable {
				so.Emit("flitter enter", __Client_Reply_bussy)
				utils.Logf(utils.Warningf, "Client %v should wait", targename)
				return
			}
			targclient, ok := s.clients[targename]
			if ok {
				_, ok = targclient.Get("socket")
				if ok {
					utils.Logf(utils.Norf, "Client %v has entered", targename)
					return
				}
			}

			so.Emit("flitter enter", targename)
			utils.Logf(utils.Infof, "Client %v entered ", targename)

			s.setClientData(targename, "socket", soidData)

			var clientsData DataItem
			err = clientsData.Parse(s.clients)
			if err != nil {
				utils.ErrIn(err)
				return
			}

			s.SetServerData("clients", clientsData)
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
			npath := core.NodePath(npathi.Data)
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
				mypath := core.NodePath(mypathi.Data)
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
		case core.MS_Succeed:
			cname, cdkey, cdvalue, hostpath, err := s.parseClientData(msg)
			if err != nil {
				return err
			}
			mypathi, ok := s.worker.Get("path")
			if ok {
				mypath := core.NodePath(mypathi.Data)
				if mypath == hostpath {
					return err
				}
			}
			s.setClientData(cname, cdkey, cdvalue)
			handler, ok := s.clientsDataUpdateHandlers[cdkey]
			if ok {
				handler(cname, cdkey, cdvalue, hostpath)
			}
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
