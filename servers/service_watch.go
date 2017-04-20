package servers

import (
	"errors"
	"fmt"
	core "github.com/gargous/flitter/core"
	utils "github.com/gargous/flitter/utils"
	//"os"
	"time"
)

type WatchService interface {
	Service
	ConfigRefereeServer(addr string)
}
type watchsrv struct {
	worker             Worker
	refereeServers     []string
	refereeServerIndex int
	baseService
}

func NewWatchService() WatchService {
	_watchsrv := &watchsrv{
		refereeServers:     make([]string, 0),
		refereeServerIndex: 0,
	}
	_watchsrv.looper = core.NewMessageLooper(__LooperSize)
	return _watchsrv
}
func (w *watchsrv) ConfigRefereeServer(addr string) {
	w.refereeServers = append(w.refereeServers, addr)
}
func (w *watchsrv) getRefereeServer() core.NodeInfo {
	info := w.refereeServers[w.refereeServerIndex]
	w.refereeServerIndex = (w.refereeServerIndex + 1) % len(w.refereeServers)
	return core.NodeInfo(info)
}
func (w *watchsrv) Init(srv interface{}) error {
	w.worker = srv.(Worker)
	w.HandleMessages()
	return nil
}
func (w *watchsrv) HandleMessages() {
	w.looper.AddHandler(3000, core.MA_Refer, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			msg.GetInfo().SetState(core.MS_Ask)
			msg.SetContent([]byte(w.worker.GetAddress()))
			err = w.worker.SendToReferee(msg, w.getRefereeServer())
			if err != nil {
				return err
			}
		case core.MS_Succeed:
			nodeinfo := core.NodeInfo(string(msg.GetContent()))
			var leader core.NodeInfo
			leader, err = nodeinfo.GetLeaderInfo()
			if err != nil {
				msg.GetInfo().SetState(core.MS_Failed)
				w.looper.Push(msg)
				return
			}
			if leader != "" {
				msg.GetInfo().SetAcion(core.MA_Init)
				msg.GetInfo().SetState(core.MS_Probe)
				w.looper.Push(msg)
			}
		case core.MS_Failed:
			utils.Logf(utils.Warningf, "%v[watch server failed]", msg.GetInfo())
			msg.GetInfo().SetTime(time.Now())
			msg.GetInfo().SetState(core.MS_Probe)
			w.looper.Push(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "watch server")
		}
		return
	})
	w.looper.AddHandler(3000, core.MA_Init, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			err = w.worker.SendService(ST_HeartBeat, msg)
			if err != nil {
				return
			}
		case core.MS_Succeed:
			utils.Logf(utils.Norf, "Init")
		case core.MS_Failed:
			msg.GetInfo().SetTime(time.Now())
			msg.GetInfo().SetState(core.MS_Probe)
			w.looper.Push(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "Init")
		}
		return
	})
}
func (w *watchsrv) Start() {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Refer)
	info.SetState(core.MS_Probe)
	info.SetTime(time.Now())
	w.looper.Push(core.NewMessage(info, []byte("Hello")))
	w.looper.Loop()
}
func (w *watchsrv) Term() {
	w.looper.Term()
}
func (w watchsrv) String() string {
	str := fmt.Sprintf("Watch Server:["+
		"\n\tlooper:%p"+
		"\n\tconfiged referee server:%v"+
		"\n\tnow using referee server:%v"+
		"\n]",
		w.looper,
		w.refereeServers,
		w.refereeServerIndex,
	)
	return str
}
