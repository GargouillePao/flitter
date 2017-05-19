package servers

import (
	"errors"
	"fmt"
	common "github.com/gargous/flitter/common"
	core "github.com/gargous/flitter/core"
	socketio "github.com/googollee/go-socket.io"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScenceService interface {
	Service
	LockClientData(name string, key string, count int) (err error)
	UnlockClientData(name string, key string, count int) (err error)
	UpdateClientData(name string, key string, value common.DataItem, count int) (err error)
	GetClientData(name string, key string) (values map[string]common.DataItem)
	OnClientUpdate(cdkey string, cb func(cname string, cdvalue common.DataItem) error)
	OnClientLock(cdkey string, cb func(cname string) error)
	IsAccess() bool
}

type scencesrvice struct {
	clientsMutex sync.Mutex
	worker       Worker
	accessable   bool
	baseService
	publicHandlers map[string]func(cname string, cdvalue common.DataItem, hostpath core.NodePath) error
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
	names := strings.Split(name, "|")
	if len(names) == 3 {
		targename = name
	} else {
		if hostname == workername {
			targename = name
		} else {
			targename = fmt.Sprintf("%s|%d|%s", workername, time.Now().UnixNano(), name)
		}
	}

	ok = true
	return
}
func (s *scencesrvice) deparseClientName(targename string) (hostname string, name string, ok bool) {
	names := strings.Split(targename, "|")
	if len(names) == 3 {
		hostname = names[0]
		name = names[2]
		ok = true
	} else {
		ok = false
	}
	return
}
func (s *scencesrvice) parseClientData(msg core.Message) (name string, key string, value common.DataItem, hostpath core.NodePath, clientCount int, err error) {
	contLen := len(msg.GetContents())
	switch {
	case contLen < 5:
		err = errors.New("Invalid Content")
	case contLen >= 5:
		name = string(msg.GetContents()[0])
		key = string(msg.GetContents()[1])
		err = value.Parse(msg.GetContents()[2])
		if err != nil {
			return
		}
		clientCount, err = strconv.Atoi(string(msg.GetContents()[3]))
		if err != nil {
			return
		}
		hostpath = core.NodePath(msg.GetContents()[4])
	}
	return
}
func (s *scencesrvice) setClientData(name string, key string, value common.DataItem) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if s.clients == nil {
		s.clients = make(map[string]common.DataSet)
	}
	targetclient, ok_ := s.clients[name]
	if !ok_ || targetclient == nil {
		targetclient = common.NewDataSet()
	}
	targetclient.Set(key, value)
	s.clients[name] = targetclient

	return
}
func (s *scencesrvice) lockClientsData(name string, key string, lock bool) (ok bool) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	if lock {
		lockedClients := make([]common.DataSet, 0)
		notAllowLock := false
		for _, client := range s.clients {
			ok = client.Lock(name, key)
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
			client.Unlock(key)
		}
		ok = false
		return
	} else {
		for _, client := range s.clients {
			client.Unlock(key)
		}
		fmt.Println("----UnLock----\n", s.clients)
	}

	ok = true
	return
}
func (s *scencesrvice) lockClientData(name string, key string, lock bool) (ok bool) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	data, ok := s.clients[name]
	if !ok {
		data = common.NewDataSet()
	}
	if lock {
		ok = data.Lock(name, key)
	} else {
		ok = true
		data.Unlock(key)
	}
	s.clients[name] = data
	return
}

//outside involk only
func (s *scencesrvice) LockClientData(name string, key string, count int) (err error) {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Lock)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	if name != "" {
		msg.AppendContent([]byte(name))
	}
	msg.AppendContent([]byte(key))
	msg.AppendContent([]byte("lock"))
	msg.AppendContent([]byte(strconv.Itoa(count)))
	ok := s.worker.putDataToMessage(msg, "path")
	if !ok {
		err = errors.New("Path Not Exsit")
		return
	}
	s.looper.Push(msg)
	return
}

//outside involk only
func (s *scencesrvice) UnlockClientData(name string, key string, count int) (err error) {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Unlock)
	info.SetState(core.MS_Probe)
	msg := core.NewMessage(info)
	if name != "" {
		msg.AppendContent([]byte(name))
	}
	msg.AppendContent([]byte(key))
	msg.AppendContent([]byte("unlock"))
	msg.AppendContent([]byte(strconv.Itoa(count)))
	ok := s.worker.putDataToMessage(msg, "path")
	if !ok {
		err = errors.New("Path Not Exsit")
		return
	}
	s.looper.Push(msg)
	return
}

//outside involk only
func (s *scencesrvice) UpdateClientData(name string, key string, value common.DataItem, count int) (err error) {
	data, ok := s.clients[name]
	if ok {
		value = data.Grant(key, value)
	}
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
	msg.AppendContent([]byte(strconv.Itoa(count)))
	ok = s.worker.putDataToMessage(msg, "path")
	if !ok {
		err = errors.New("Path Not Exsit")
		return
	}
	s.looper.Push(msg)
	return
}

func (s *scencesrvice) GetClientDatas(name string) map[string]common.DataSet {
	if name == "" {
		return s.clients
	} else {
		return map[string]common.DataSet{name: s.clients[name]}
	}
}
func (s *scencesrvice) GetClientData(name string, key string) (values map[string]common.DataItem) {
	sdataset := s.GetClientDatas(name)
	values = make(map[string]common.DataItem)
	for name, data := range sdataset {
		value, ok := data.Get(key)
		if ok {
			values[name] = value
		}
	}
	return
}
func (s *scencesrvice) SetServerData(key string, value common.DataItem) {
	s.worker.Set(key, value)
}
func (s *scencesrvice) OnClientUpdate(cdkey string, cb func(cname string, cdvalue common.DataItem) error) {
	s.On(core.MA_Update, cdkey, func(cname string, cdvalue common.DataItem, hostpath core.NodePath) error {
		return cb(cname, cdvalue)
	})
}
func (s *scencesrvice) OnClientLock(cdkey string, cb func(cname string) error) {
	s.On(core.MA_Lock, cdkey, func(cname string, cdvalue common.DataItem, hostpath core.NodePath) error {
		return cb(cname)
	})
}
func (s *scencesrvice) On(action core.MessageAction, cdkey string, cb func(cname string, cdvalue common.DataItem, hostpath core.NodePath) error) {
	if s.publicHandlers == nil {
		s.publicHandlers = make(map[string]func(cname string, cdvalue common.DataItem, hostpath core.NodePath) error)
	}
	s.publicHandlers[action.String()+cdkey] = cb
}
func (s *scencesrvice) HandleClients() {
	s.worker.OnClient("flitter enter", func(so socketio.Socket) interface{} {
		return func(name string, hostname string) {
			var soidData common.DataItem
			err := soidData.Parse(so.Id())
			if err != nil {
				common.ErrIn(err)
				return
			}
			targename, ok := s.parseClientName(hostname, name)
			if !ok {
				common.ErrIn(errors.New("Parse Client Error"))
				return
			}
			common.Logf(common.Norf, "Client %v want to enter", targename)

			if !s.accessable {
				so.Emit("flitter enter", __Client_Reply_bussy)
				common.Logf(common.Warningf, "Client %v should wait", targename)
				return
			}
			targclient, ok := s.clients[targename]
			if ok {
				_, ok = targclient.Get("socket")
				if ok {
					common.Logf(common.Warningf, "Client %v has entered", targename)
					return
				}
			}
			so.Emit("flitter enter", targename)
			common.Logf(common.Infof, "Client %v entered ", targename)

			s.setClientData(targename, "socket", soidData)

			var clientsData common.DataItem
			err = clientsData.Parse(s.clients)
			if err != nil {
				common.ErrIn(err)
				return
			}

			s.SetServerData("clients", clientsData)
		}
	})
}

var (
	__err_Not_Catch_Lock error = errors.New("Not_Catch_Lock")
)

func (s scencesrvice) ThatIsMe(cname string) (ok bool, err error) {
	npathd, ok := s.worker.Get("path")
	if !ok {
		err = errors.New("Path Not Exsit")
		return
	}
	ninfo, ok := core.NodePath(npathd.Data).GetNodeInfo()
	if !ok {
		err = errors.New("Node Not Exsit")
		return
	}
	workername, _, ok := s.deparseClientName(cname)
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
		cbAsk func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) error,
		cbSucceed func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) error,
	) (err error) {
		action, state, _ := msg.GetInfo().Info()
		cname, cdkey, cdvalue, hostpath, count, err := s.parseClientData(msg)
		if err != nil {
			return err
		}
		host, ok := hostpath.GetNodeInfo()
		if !ok {
			return errors.New("Invalid NodePath")
		}

		switch state {
		case core.MS_Probe:
			//assert lock
			cname, ok = s.parseClientName(host.Name, cname)
			if !ok {
				return errors.New("ParseFailed")
			}
			if s.clients != nil {
				switch count {
				case -1:
					for _, data := range s.clients {
						if data.IsLocked(cname, cdkey) {
							return errors.New("Lock Failed When Probe")
						}
					}
				case 0:
					if s.worker.IsLocked(cname, cdkey) {
						return errors.New("Lock Failed When Probe")
					}
				case 1:
					clientData, ok := s.clients[cname]
					if ok {
						if clientData.IsLocked(cname, cdkey) {
							return errors.New("Lock Failed When Probe")
						}
					}
				default:
					errors.New("Client Count")
				}

			}
			//send to leader
			dleader, ok := s.worker.Get("path_leader")
			if !ok {
				msg.GetInfo().SetState(core.MS_Succeed)
				s.looper.Push(msg)
			} else {
				nleader := core.NodePath(dleader.Data)
				msg.GetInfo().SetState(core.MS_Ask)
				err = s.worker.SendToWroker(msg, nleader)
			}
		case core.MS_Ask:
			mypathi, ok := s.worker.Get("path")
			if ok {
				mypath := core.NodePath(mypathi.Data)
				if mypath == hostpath {
					return nil
				}
			}
			err = cbAsk(cname, cdkey, cdvalue, hostpath, count)
			if err != nil {
				return err
			}
		case core.MS_Succeed:
			err = s.worker.PublishToWorker(msg)
			if err != nil {
				return err
			}
			err = cbSucceed(cname, cdkey, cdvalue, hostpath, count)
			if err != nil {
				if err == __err_Not_Catch_Lock {
					return nil
				}
				return err
			}
			if s.publicHandlers != nil {
				handler, ok := s.publicHandlers[action.String()+cdkey]
				if ok {
					err = handler(cname, cdvalue, hostpath)
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
			func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) error {
				return s.LockClientData(name, key, count)
			},
			func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) (err error) {

				switch count {
				case -1:
					if !s.lockClientsData(name, key, true) {
						msg.GetInfo().SetState(core.MS_Failed)
						s.looper.Push(msg)
						err = errors.New("Lock Failed When Succeed")
					}
				case 0:
					if !s.worker.Lock(name, key) {
						msg.GetInfo().SetState(core.MS_Failed)
						s.looper.Push(msg)
						err = errors.New("Lock Failed When Succeed")
					}
				case 1:
					if !s.lockClientData(name, key, true) {
						msg.GetInfo().SetState(core.MS_Failed)
						s.looper.Push(msg)
						err = errors.New("Lock Failed When Succeed")
					}
				}

				ok, err := s.ThatIsMe(name)
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
			func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) error {
				return s.UnlockClientData(name, key, count)
			},
			func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) (err error) {
				switch count {
				case -1:
					s.lockClientsData(name, key, false)
				case 0:
					s.worker.Unlock(key)
				case 1:
					s.lockClientData(name, key, false)
				}
				ok, err := s.ThatIsMe(name)
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
			func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) error {
				return s.UpdateClientData(name, key, value, count)
			},
			func(name string, key string, value common.DataItem, hostpath core.NodePath, count int) error {
				switch count {
				case -1:
					for _, client := range s.clients {
						client.Set(key, value)
					}
				case 0:
					s.worker.Set(key, value)
				case 1:
					s.setClientData(name, key, value)
				}
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
