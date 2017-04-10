package servers

import (
	"errors"
	//"flag"
	//"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	"sync"
	//"time"
)

type NameServer interface {
	Server
}
type namesrv struct {
	senderToWatch   core.Sender
	recverFromWatch core.Receiver
	looper          core.MessageLooper
	nameTree        core.NodeTree
	mutex           sync.Mutex
}

func NewNameServer() NameServer {
	return &namesrv{
		looper: core.NewMessageLooper(__LooperSize),
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
func (n *namesrv) SearchNodeInfo(address string) core.NodeInfo {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.nameTree == nil {
		n.nameTree = core.NewNodeTree(core.NodeInfo(address))
	}
	node, ok := n.nameTree.Search(address)
	if ok {
		return node
	} else {
		return n.nameTree.Add(address)
	}
}
func (n *namesrv) HandleMessages() {
	n.looper.AddHandler(0, core.MA_Refer, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Ask:
			addr, err := core.NodeInfo(msg.GetContent()).GetAddressInfo()
			if err != nil {
				return err
			}
			nodeinfo := n.SearchNodeInfo(addr)
			msg.SetContent([]byte(nodeinfo))

			n.senderToWatch.Disconnect(true)
			n.senderToWatch.AddNodeInfo(nodeinfo)
			err = n.senderToWatch.Connect()

			if err != nil {
				return err
			}
			msg.GetInfo().SetState(core.MS_Succeed)
			err = n.senderToWatch.Send(msg)
			if err != nil {
				return err
			}
			//fmt.Println(n.senderToWatch, msg)
			//for testing
			//msg.GetInfo().SetAcion(core.MA_Term)
			//n.looper.Push(msg)
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
	n.looper.Term()
}
