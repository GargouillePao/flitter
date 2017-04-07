package servers

import (
	"errors"
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	zmq "github.com/pebbe/zmq4"
	//"os"
	"time"
)

type WatchServer interface {
	Server
}
type watchsrv struct {
	senderToName         core.Sender
	recverFromName       core.Receiver
	notifierForHeartBeat core.Deliverer
	looper               core.MessageLooper
	nameservers          []string
	nameserverIndex      int
}

func NewWatchServer() WatchServer {
	return &watchsrv{
		looper:          core.NewMessageLooper(__looperSize),
		nameservers:     make([]string, 0),
		nameserverIndex: 0,
	}
}
func (w *watchsrv) Config(ca ConfigAction, st ServerType, info core.NodeInfo) error {
	var err error
	switch st {
	case ST_Name:
		if ca == CA_Send {
			if w.senderToName == nil {
				sender, err := core.NewSender()
				if err != nil {
					return err
				}
				w.senderToName = sender
			}
			w.nameservers = append(w.nameservers, string(info))
		} else {
			w.recverFromName, err = core.NewReceiver(info)
			if err != nil {
				return err
			}
		}
	case ST_HeartBeat:
		if ca == CA_Send {
			if w.notifierForHeartBeat == nil {
				return errors.New("Heartbeat notifier need config recv first")
			}
			w.notifierForHeartBeat.AddNodeInfo(info)
		} else {
			w.notifierForHeartBeat, err = core.NewDeliverer(info, zmq.DEALER)
			if err != nil {
				return err
			}
		}
	}
	return err
}
func (w *watchsrv) getNameServer() core.NodeInfo {
	info := w.nameservers[w.nameserverIndex]
	w.nameserverIndex = (w.nameserverIndex + 1) % len(w.nameservers)
	return core.NodeInfo(info)
}
func (w *watchsrv) Init() error {
	var err error
	if w.senderToName == nil || w.recverFromName == nil || w.notifierForHeartBeat == nil {
		err = errors.New("Config not enough")
		return err
	}
	err = w.recverFromName.Bind()
	if err != nil {
		return err
	}
	err = w.notifierForHeartBeat.Bind()
	if err != nil {
		return err
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
					utils.ErrIn(err, "[receive at watch from name]")
					continue
				}
				w.looper.Push(msg)
			}
		}
	}()
	go func() {
		if w.notifierForHeartBeat != nil {
			for {
				msg, err := w.notifierForHeartBeat.Recv()
				if err != nil {
					utils.ErrIn(err, "[receive at watch from heartbeat]")
					continue
				}
				w.looper.Push(msg)
			}
		}
	}()
}
func (w *watchsrv) HandleMessages() {
	w.looper.AddHandler(1000, core.MA_Refer, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			w.senderToName.AddNodeInfo(w.getNameServer())
			err := w.senderToName.Connect()
			if err != nil {
				return err
			}
			msg.GetInfo().SetState(core.MS_Ask)
			err = w.senderToName.Send(msg)
			if err != nil {
				return err
			}
		case core.MS_Succeed:
		case core.MS_Failed:
			for _, nameserver := range w.nameservers {
				w.senderToName.RemoveNodeInfo(core.NodeInfo(nameserver))
			}
			if __verb {
				fmt.Println(utils.Warningf("%v[watch server failed]", msg.GetInfo()))
			}
			msg.GetInfo().SetTime(time.Now())
			msg.GetInfo().SetState(core.MS_Probe)
			w.looper.Push(msg)

		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "[watch server]")
		}
		return nil
	})
}
func (w *watchsrv) Start() {
	lauchServer()
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Refer)
	info.SetState(core.MS_Probe)
	info.SetTime(time.Now())
	w.looper.Push(core.NewMessage(info, []byte("Hello")))
	w.looper.Loop()
}
func (w *watchsrv) Term() {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Term)
	w.looper.Push(core.NewMessage(info, []byte("")))
}
