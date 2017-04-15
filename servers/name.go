package servers

import (
	"errors"
	"strings"
	//"flag"
	"fmt"
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

func (n *namesrv) Config(ca ConfigAction, st ServerType, addr string) error {
	var err error
	switch st {
	case ST_Watch:
		if ca == CA_Recv {
			nameaddr, err := transAddress(addr, ST_Watch, ST_Name)
			if err != nil {
				return err
			}
			n.recverFromWatch, err = core.NewReceiver(nameaddr)
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
			addr, err := transAddress(string(msg.GetContent()), ST_Name, ST_Watch)
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
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "[node server]")
		}
		return nil
	})
}
func (n *namesrv) Start() {
	n.looper.Loop()
}
func (n *namesrv) Term() {
	n.senderToWatch.Close()
	n.recverFromWatch.Close()
	n.looper.Term()
}
func (n namesrv) String() string {
	str := fmt.Sprintf("Name Server:["+
		"\n\tlooper:%p"+
		"\n\treceiver:%v"+
		"\n\t,sender:%v"+
		"\n\t,tree:%v"+
		"]",
		n.looper,
		strings.Join(strings.Split(fmt.Sprintf("%v", n.recverFromWatch), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", n.senderToWatch), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", n.nameTree), "\n"), "\n\t"),
	)
	return str
}
