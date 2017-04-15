package servers

import (
	"errors"
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	//"os"
	"strings"
	"time"
)

type WatchServer interface {
	Server
}
type watchsrv struct {
	senderToName    core.Sender
	recverFromName  core.Receiver
	senderToHeart   core.Sender
	recverFromHeart core.Receiver
	looper          core.MessageLooper
	nameservers     []string
	nameserverIndex int
}

func NewWatchServer() WatchServer {
	return &watchsrv{
		looper:          core.NewMessageLooper(__LooperSize),
		nameservers:     make([]string, 0),
		nameserverIndex: 0,
	}
}
func (w *watchsrv) Config(ca ConfigAction, st ServerType, addr string) (err error) {
	switch st {
	case ST_Name:
		var nameaddr string
		if ca == CA_Send {
			if w.senderToName == nil {
				w.senderToName, err = core.NewSender()
				if err != nil {
					return
				}
			}
			nameaddr, err = transAddress(addr, ST_Watch, ST_Name)
			if err != nil {
				return
			}
			w.nameservers = append(w.nameservers, nameaddr)
		} else {
			nameaddr, err = transAddress(addr, ST_Name, ST_Watch)
			if err != nil {
				return
			}
			w.recverFromName, err = core.NewReceiver(nameaddr)
			if err != nil {
				return
			}
		}
	case ST_HeartBeat:
		var heartbeataddr string
		if ca == CA_Send {
			w.senderToHeart, err = core.NewSender()
			if err != nil {
				return
			}
			heartbeataddr, err = transAddress(addr, ST_Watch, ST_HeartBeat)
			if err != nil {
				return
			}
			w.senderToHeart.AddNodeInfo(core.NodeInfo(heartbeataddr))
			if err != nil {
				return
			}
		} else {
			heartbeataddr, err = transAddress(addr, ST_HeartBeat, ST_Watch)
			if err != nil {
				return
			}
			w.recverFromHeart, err = core.NewReceiver(heartbeataddr)
			if err != nil {
				return
			}
		}
	}
	return
}
func (w *watchsrv) getNameServer() core.NodeInfo {
	info := w.nameservers[w.nameserverIndex]
	w.nameserverIndex = (w.nameserverIndex + 1) % len(w.nameservers)
	return core.NodeInfo(info)
}
func (w *watchsrv) Init() error {
	var err error
	if w.senderToName == nil || w.recverFromName == nil || w.senderToHeart == nil || w.recverFromHeart == nil {
		err = errors.New("Config not enough")
		return err
	}
	err = w.recverFromName.Bind()
	if err != nil {
		return utils.ErrAppend(err, " Recver(from name) Bind")
	}
	err = w.recverFromHeart.Bind()
	if err != nil {
		return utils.ErrAppend(err, " Recver(for heartbeat) Bind")
	}
	w.HandleRecive()
	w.HandleMessages()
	return err
}
func (w *watchsrv) HandleRecive() {
	go func() {
		if w.recverFromName != nil {
			for {
				msg, err := w.recverFromName.Recv()
				if err != nil {
					utils.ErrIn(err, "receive at watch from name")
					continue
				}
				w.looper.Push(msg)
			}
		}
	}()
	go func() {
		if w.recverFromHeart != nil {
			for {
				msg, err := w.recverFromHeart.Recv()
				if err != nil {
					utils.ErrIn(err, "receive at watch from heartbeat")
					continue
				}
				w.looper.Push(msg)
			}
		}
	}()
}
func (w *watchsrv) HandleMessages() {
	w.looper.AddHandler(3000, core.MA_Refer, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			w.senderToName.Disconnect(true)
			w.senderToName.AddNodeInfo(w.getNameServer())
			err := w.senderToName.Connect()
			if err != nil {
				return err
			}
			msg.GetInfo().SetState(core.MS_Ask)
			msg.SetContent([]byte(w.recverFromName.GetBindNodeInfo()))
			err = w.senderToName.Send(msg)
			if err != nil {
				return err
			}

		case core.MS_Succeed:
			nodeinfo := core.NodeInfo(string(msg.GetContent()))
			leader, err := nodeinfo.GetLeaderInfo()
			if err != nil {
				msg.GetInfo().SetState(core.MS_Failed)
				w.looper.Push(msg)
				return err
			}

			if leader != "" {
				msg.GetInfo().SetAcion(core.MA_Init_Heartbeat)
				msg.GetInfo().SetState(core.MS_Probe)
				w.looper.Push(msg)
			}
		case core.MS_Failed:
			for _, nameserver := range w.nameservers {
				w.senderToName.RemoveNodeInfo(core.NodeInfo(nameserver))
			}
			utils.Logf(utils.Warningf, "%v[watch server failed]", msg.GetInfo())
			msg.GetInfo().SetTime(time.Now())
			msg.GetInfo().SetState(core.MS_Probe)
			w.looper.Push(msg)

		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "watch server")
		}
		return nil
	})
	w.looper.AddHandler(3000, core.MA_Init_Heartbeat, func(msg core.Message) error {
		var err error
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:

			err = w.senderToHeart.Connect()
			if err != nil {
				return err
			}
			err = w.senderToHeart.Send(msg)
			if err != nil {
				return err
			}
		case core.MS_Succeed:
			utils.Logf(utils.Norf, "I start a heartbeat")
		case core.MS_Failed:
			msg.GetInfo().SetTime(time.Now())
			msg.GetInfo().SetState(core.MS_Probe)
			w.looper.Push(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "init heartbeat server")
		}
		return err
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
	w.senderToHeart.Close()
	w.senderToName.Close()
	w.recverFromHeart.Close()
	w.recverFromName.Close()
	w.looper.Term()
}
func (w watchsrv) String() string {
	str := fmt.Sprintf("Watch Server:["+
		"\n\tlooper:%p"+
		"\n\tname receiver:%v"+
		"\n\tname sender:%v"+
		"\n\theart receiver:%v"+
		"\n\theart sender:%v"+
		"\n\tconfiged name server:%v"+
		"\n\tnow using name server:%v"+
		"\n]",
		w.looper,
		strings.Join(strings.Split(fmt.Sprintf("%v", w.recverFromName), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", w.senderToName), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", w.recverFromHeart), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", w.senderToHeart), "\n"), "\n\t"),
		w.nameservers,
		w.nameserverIndex,
	)
	return str
}
