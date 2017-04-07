package servers

import (
	"errors"
	//"flag"
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	//"time"
)

type NameServer interface {
	Server
}
type namesrv struct {
	senderToWatch   core.Sender
	recverFromWatch core.Receiver
	looper          core.MessageLooper
}

func NewNameServer() NameServer {
	return &namesrv{
		looper: core.NewMessageLooper(__looperSize),
	}
}

func (n *namesrv) Config(ca ConfigAction, st ServerType, info core.NodeInfo) error {
	var err error
	switch st {
	case ST_Watch:
		if ca == CA_Recv {
			n.recverFromWatch, err = core.NewReceiver(info)
			if err != nil {
				return err
			}
			n.senderToWatch, err = core.NewSender()
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (n *namesrv) Init() error {
	if n.senderToWatch == nil || n.recverFromWatch == nil {
		return errors.New("Config not enough")
	}
	err := n.recverFromWatch.Bind()
	if err != nil {
		return err
	}
	n.HandleRecive()
	n.HandleMessages()
	return nil
}
func (n *namesrv) HandleRecive() {
	go func() {
		if n.recverFromWatch != nil {
			for {
				msg, err := n.recverFromWatch.Recv()
				if err != nil {
					utils.ErrIn(err, "[receive at name from watch]")
					continue
				}
				n.looper.Push(msg)
			}
		}
	}()
}
func (n *namesrv) HandleMessages() {
	n.looper.AddHandler(0, core.MA_Refer, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Ask:
			fmt.Println(utils.Infof("ask for %v", msg))
			msg.GetInfo().SetAcion(core.MA_Term)
			n.looper.Push(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "[node server]")
		}
		return nil
	})
}
func (n *namesrv) Start() {
	lauchServer()
	n.looper.Loop()
}
func (n *namesrv) Term() {
	info := core.NewMessageInfo()
	info.SetAcion(core.MA_Term)
	n.looper.Push(core.NewMessage(info, []byte("")))
}
