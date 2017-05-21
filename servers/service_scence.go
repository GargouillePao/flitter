package servers

import (
	"errors"
	"fmt"
	common "github.com/gargous/flitter/common"
	core "github.com/gargous/flitter/core"
	socketio "github.com/googollee/go-socket.io"
	"strings"
	"sync"
	"time"
)

type ScenceService interface {
	Service
	LockClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (err error)
	UnlockClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (err error)
	UpdateClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (err error)
	GetClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (values map[string]common.DataItem)
	OnClientUpdate(dInfo core.DataInfo, cb func(cInfo core.ClientInfo, dInfo core.DataInfo) error)
	OnClientLock(dInfo core.DataInfo, cb func(cInfo core.ClientInfo) error)
	IsAccess() bool
}

type scencesrvice struct {
	clientsMutex sync.Mutex
	worker       Worker
	accessable   bool
	baseService
	publicHandlers map[string]func(cInfo core.ClientInfo, dInfo core.DataInfo) error
	clients        map[string]common.DataSet
}

func NewScenceService() ScenceService {
	service := &scencesrvice{}
	service.accessable = false
	service.clients = make(map[string]common.DataSet)
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
func (s *scencesrvice) setClientData(cInfo core.ClientInfo, dInfo core.DataInfo) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	if s.clients == nil {
		s.clients = make(map[string]common.DataSet)
	}
	targetclient, ok_ := s.clients[cInfo.GetName()]
	if !ok_ || targetclient == nil {
		targetclient = common.NewDataSet()
	}
	targetclient.Set(dInfo.Key, dInfo.Value)
	s.clients[cInfo.GetName()] = targetclient

	return
}
func (s *scencesrvice) lockClientsData(cInfo core.ClientInfo, dInfo core.DataInfo, lock bool) (ok bool) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	if lock {
		lockedClients := make([]common.DataSet, 0)
		notAllowLock := false
		for _, client := range s.clients {
			ok = client.Lock(cInfo.GetName(), dInfo.Key)
			if !ok {
				notAllowLock = true
				break
			}
			lockedClients = append(lockedClients, client)
		}
		fmt.Println("----Lock----\n", lockedClients)
		if !notAllowLock {
			ok = true
			return
		}
		for _, client := range lockedClients {
			client.Unlock(dInfo.Key)
		}
		ok = false
		return
	} else {
		for _, client := range s.clients {
			client.Unlock(dInfo.Key)
		}
		fmt.Println("----UnLock----\n", s.clients)
	}

	ok = true
	return
}
func (s *scencesrvice) lockClientData(cInfo core.ClientInfo, dInfo core.DataInfo, lock bool) (ok bool) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	data, ok := s.clients[cInfo.GetName()]
	if !ok {
		data = common.NewDataSet()
	}
	if lock {
		ok = data.Lock(cInfo.GetName(), dInfo.Key)
	} else {
		ok = true
		data.Unlock(dInfo.Key)
	}
	s.clients[cInfo.GetName()] = data
	return
}

//outside involk only
func (s *scencesrvice) LockClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Lock)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	cInfo.AppendToMsg(msg)
	err = dInfo.AppendToMsg(msg)
	if err != nil {
		return
	}
	s.looper.Push(msg)
	return
}

//outside involk only
func (s *scencesrvice) UnlockClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Unlock)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	cInfo.AppendToMsg(msg)
	err = dInfo.AppendToMsg(msg)
	if err != nil {
		return
	}
	s.looper.Push(msg)
	return
}

//outside involk only
func (s *scencesrvice) UpdateClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
	data, ok := s.clients[cInfo.GetName()]
	if ok {
		dInfo.Value = data.Grant(dInfo.Key, dInfo.Value)
	}
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Update)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	cInfo.AppendToMsg(msg)
	err = dInfo.AppendToMsg(msg)
	if err != nil {
		return
	}
	s.looper.Push(msg)
	return
}

func (s *scencesrvice) GetClientDatas(cInfo core.ClientInfo) map[string]common.DataSet {
	if cInfo.GetName() == "" {
		return s.clients
	} else {
		return map[string]common.DataSet{cInfo.GetName(): s.clients[cInfo.GetName()]}
	}
}
func (s *scencesrvice) GetClientData(cInfo core.ClientInfo, dInfo core.DataInfo) (values map[string]common.DataItem) {
	sdataset := s.GetClientDatas(cInfo)
	values = make(map[string]common.DataItem)
	for name, data := range sdataset {
		if data != nil {
			value, ok := data.Get(dInfo.Key)
			if ok {
				values[name] = value
			}
		}
	}
	return
}
func (s *scencesrvice) SetServerData(dInfo core.DataInfo) {
	s.worker.Set(dInfo.Key, dInfo.Value)
}
func (s *scencesrvice) OnClientUpdate(dInfo core.DataInfo, cb func(cInfo core.ClientInfo, dInfo core.DataInfo) error) {
	s.On(core.MA_Update, dInfo, func(cInfo core.ClientInfo, dInfo core.DataInfo) error {
		return cb(cInfo, dInfo)
	})
}
func (s *scencesrvice) OnClientLock(dInfo core.DataInfo, cb func(cInfo core.ClientInfo) error) {
	s.On(core.MA_Lock, dInfo, func(cInfo core.ClientInfo, dInfo core.DataInfo) error {
		return cb(cInfo)
	})
}
func (s *scencesrvice) On(action core.MessageAction, dInfo core.DataInfo, cb func(cInfo core.ClientInfo, dInfo core.DataInfo) error) {
	if s.publicHandlers == nil {
		s.publicHandlers = make(map[string]func(cInfo core.ClientInfo, dInfo core.DataInfo) error)
	}
	s.publicHandlers[action.String()+dInfo.Key] = cb
}
func (s *scencesrvice) HandleClients() {
	s.worker.OnClient("flitter enter", func(so socketio.Socket) interface{} {
		return func(name string) {
			var soidData common.DataItem
			err := soidData.Parse(so.Id())
			if err != nil {
				common.ErrIn(err)
				return
			}
			cInfo := core.NewClientInfo(name, s.worker.GetPath())
			common.Logf(common.Norf, "Client %v want to enter", cInfo.GetName())
			if !s.accessable {
				so.Emit("flitter enter", __Client_Reply_bussy)
				common.Logf(common.Warningf, "Client %v should wait", cInfo.GetName())
				return
			}
			targclient, ok := s.clients[cInfo.GetName()]
			if ok {
				_, ok = targclient.Get("socket")
				if ok {
					common.Logf(common.Warningf, "Client %v has entered", cInfo.GetName())
					return
				}
			}
			so.Emit("flitter enter", cInfo.GetName())
			common.Logf(common.Infof, "Client %v entered ", cInfo.GetName())
			dInfo := core.NewDataInfo("socket")
			dInfo.Value = soidData
			s.setClientData(cInfo, dInfo)
			var clientsData common.DataItem
			err = clientsData.Parse(s.clients)
			if err != nil {
				common.ErrIn(err)
				return
			}
			serverInfo := core.NewDataInfo("clients")
			serverInfo.Value = clientsData
			s.SetServerData(serverInfo)
		}
	})
}

var (
	__err_Not_Catch_Lock error = errors.New("Not_Catch_Lock")
)

func (s scencesrvice) ThatIsMe(cname string) (ok bool, err error) {
	ninfo, ok := core.NodePath(s.worker.GetPath()).GetNodeInfo()
	if !ok {
		err = errors.New("Node Not Exsit")
		return
	}
	workername, _, ok := common.DeparseClientName(cname)
	if !ok {
		err = errors.New("Invalid Name")
		return
	}
	if workername != ninfo.Name {
		ok = false
		return
	}
	ok = true
	return
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
	broadcast := func(
		msg core.Message,
		cbAsk func(cInfo core.ClientInfo, dInfo core.DataInfo) error,
		cbSucceed func(cInfo core.ClientInfo, dInfo core.DataInfo) error,
	) (err error) {
		action, state, _ := msg.GetInfo().Info()
		var clientInfo core.ClientInfo
		ok := clientInfo.Parse(msg)
		if !ok {
			return errors.New("Invalid ClientInfo")
		}
		var dataInfo core.DataInfo
		ok = dataInfo.Parse(msg)
		if !ok {
			return errors.New("Invalid DataInfo")
		}
		switch state {
		case core.MS_Probe:
			//assert lock
			cname := clientInfo.GetName()
			if s.clients != nil {
				ok := dataInfo.AssertCount(
					func() bool {
						for _, data := range s.clients {
							if data.IsLocked(cname, dataInfo.Key) {
								return false
							}
						}
						return true
					},
					func() bool {
						if s.worker.IsLocked(cname, dataInfo.Key) {
							return false
						}
						return true
					},
					func() bool {
						clientData, ok := s.clients[cname]
						if ok {
							if clientData.IsLocked(cname, dataInfo.Key) {
								return false
							}
						}
						return true
					},
				)
				if !ok {
					return errors.New("Lock Failed When Probe")
				}
			}
			//send to leader
			nleader, ok := s.worker.GetPath().GetLeaderPath()
			if !ok {
				msg.GetInfo().SetState(core.MS_Succeed)
				s.looper.Push(msg)
			} else {
				msg.GetInfo().SetState(core.MS_Ask)
				err = s.worker.SendToWroker(msg, nleader)
			}
		case core.MS_Ask:
			mypath := s.worker.GetPath()
			if mypath == clientInfo.GetPath() {
				return nil
			}
			err = cbAsk(clientInfo, dataInfo)
			if err != nil {
				return err
			}
		case core.MS_Succeed:
			err = s.worker.PublishToWorker(msg)
			if err != nil {
				return err
			}
			err = cbSucceed(clientInfo, dataInfo)
			if err != nil {
				if err == __err_Not_Catch_Lock {
					return nil
				}
				return err
			}
			if s.publicHandlers != nil {
				handler, ok := s.publicHandlers[action.String()+dataInfo.Key]
				if ok {
					err = handler(clientInfo, dataInfo)
				}
			}
		case core.MS_Failed:
			common.Logf(common.Warningf, "%v Faild And Now %v", action, msg)
			time.AfterFunc(time.Second, func() {
				msg.GetInfo().SetTime(time.Now())
				msg.GetInfo().SetState(core.MS_Probe)
				s.looper.Push(msg)
			})
		case core.MS_Error:
			common.ErrIn(errors.New(msg.String()))
		}
		return
	}
	s.looper.AddHandler(3000, core.MA_Lock, func(msg core.Message) (err error) {
		return broadcast(
			msg,
			func(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
				return s.LockClientData(cInfo, dInfo)
			},
			func(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
				ok := dInfo.AssertCount(
					func() bool {
						if !s.lockClientsData(cInfo, dInfo, true) {
							msg.GetInfo().SetState(core.MS_Failed)
							s.looper.Push(msg)
							return false
						}
						return true
					},
					func() bool {
						if !s.worker.Lock(cInfo.GetName(), dInfo.Key) {
							msg.GetInfo().SetState(core.MS_Failed)
							s.looper.Push(msg)
							return false
						}
						return true
					},
					func() bool {
						if !s.lockClientData(cInfo, dInfo, true) {
							msg.GetInfo().SetState(core.MS_Failed)
							s.looper.Push(msg)
							return false
						}
						return true
					},
				)
				if !ok {
					err = errors.New("Lock Failed When Succeed")
					return
				}

				ok, err = s.ThatIsMe(cInfo.GetName())
				if err != nil {
					return
				}
				if !ok {
					err = __err_Not_Catch_Lock
				}
				return
			},
		)
	})
	s.looper.AddHandler(3000, core.MA_Unlock, func(msg core.Message) (err error) {
		return broadcast(
			msg,
			func(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
				return s.UnlockClientData(cInfo, dInfo)
			},
			func(cInfo core.ClientInfo, dInfo core.DataInfo) (err error) {
				dInfo.AssertCount(
					func() bool {
						s.lockClientsData(cInfo, dInfo, false)
						return true
					},
					func() bool {
						s.worker.Unlock(dInfo.Key)
						return true
					},
					func() bool {
						s.lockClientData(cInfo, dInfo, false)
						return true
					},
				)
				ok, err := s.ThatIsMe(cInfo.GetName())
				if err != nil {
					return
				}
				if !ok {
					err = __err_Not_Catch_Lock
				}
				return
			},
		)
	})
	s.looper.AddHandler(3000, core.MA_Update, func(msg core.Message) (err error) {
		return broadcast(
			msg,
			func(cInfo core.ClientInfo, dInfo core.DataInfo) error {
				return s.UpdateClientData(cInfo, dInfo)
			},
			func(cInfo core.ClientInfo, dInfo core.DataInfo) error {
				dInfo.AssertCount(
					func() bool {
						for _, client := range s.clients {
							client.Set(dInfo.Key, dInfo.Value)
						}
						return true
					},
					func() bool {
						s.worker.Set(dInfo.Key, dInfo.Value)
						return true
					},
					func() bool {
						s.setClientData(cInfo, dInfo)
						return true
					},
				)
				return nil
			},
		)
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
	str := fmt.Sprintf("Scene Service:[\n\tlooper:%p\n\tclients:%s\n]", s.looper, s.ClientsString())
	return str
}
