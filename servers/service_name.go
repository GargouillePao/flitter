package servers

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/data"
	"github.com/gargous/flitter/utils"
	socketio "github.com/googollee/go-socket.io"
	"strings"
	"sync"
)

type NameService interface {
	Service
}
type namesrv struct {
	referee       Referee
	nameTree      core.NodeTree
	nameTreeSaver data.NodeTreeSaver
	mutex         sync.Mutex
	baseService
}

func NewNameService() NameService {
	srv := &namesrv{
		nameTree: core.NewNodeTree(),
	}
	srv.looper = core.NewMessageLooper(__LooperSize)
	srv.nameTreeSaver = data.NewNodeTreeSaver(srv.nameTree)
	return srv
}

func (n *namesrv) Init(srv interface{}) error {
	n.referee = srv.(Referee)
	n.HandleMessages()
	n.HandleClients()
	return nil
}
func (n *namesrv) SearchNodeInfo(address string) core.NodeInfo {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.nameTree == nil {
		n.nameTree = core.NewNodeTree()
		n.nameTreeSaver = data.NewNodeTreeSaver(n.nameTree)
		n.nameTreeSaver.Load()
	}
	node, ok := n.nameTree.Search(address)
	if !ok {
		node = n.nameTree.Add(address)
		//n.nameTreeSaver.Save(data.SA_Add, address)
	}
	utils.Logf(utils.Norf, "Search\n%v", n.nameTree)
	return node
}
func (n *namesrv) SearchNodeInfoAtHeight(index int, height int) core.NodeInfo {
	if n.nameTree == nil {
		return ""
	}
	_index := 0
	targetAddress := n.nameTree.FLoop(0, func(_height int, node core.NodeInfo) bool {
		if height == _height {
			if _index == index {
				return true
			}
			_index++
		}
		return false
	})
	return targetAddress
}
func (n *namesrv) HandleClients() {
	n.referee.AddClientHandler(func(so socketio.Socket) {
		so.On("refer address", func(index int, height int) {
			addr := n.SearchNodeInfoAtHeight(index, height)
			so.Emit("refer address", addr)
		})
	})
}
func (n *namesrv) HandleMessages() {
	n.looper.AddHandler(0, core.MA_Refer, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Ask:
			nodeinfo := n.SearchNodeInfo(string(msg.GetContent()))
			msg.SetContent([]byte(nodeinfo))
			msg.GetInfo().SetState(core.MS_Succeed)
			err := n.referee.SendToWroker(msg, nodeinfo)
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
	n.looper.Term()
}
func (n namesrv) String() string {
	str := fmt.Sprintf("Name Server:["+
		"\n\tlooper:%p"+
		"\n\ttree:%v"+
		"\n]",
		n.looper,
		strings.Join(strings.Split(fmt.Sprintf("%v", n.nameTree), "\n"), "\n\t"),
	)
	return str
}
