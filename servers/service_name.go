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
func (n *namesrv) SearchNodeInfo(npath core.NodePath) (opath core.NodePath, err error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.nameTree == nil {
		n.nameTree = core.NewNodeTree()
		n.nameTreeSaver = data.NewNodeTreeSaver(n.nameTree)
		n.nameTreeSaver.Load()
	}
	info, ok := npath.GetNodeInfo()
	if !ok {
		err = errors.New("Invalid NodeInfo When SearchNodeInfo")
		return
	}
	opath, ok = n.nameTree.Search(info)
	if !ok {
		opath, err = n.nameTree.Add(npath)
		if err != nil {
			return
		}
		//n.nameTreeSaver.SaveLastItem(data.SA_Add)
	}
	utils.Logf(utils.Norf, "Search\n%v", n.nameTree)
	return
}
func (n *namesrv) SearchNodeInfoWithGroupName(groupname string, index int) core.NodePath {
	if n.nameTree == nil {
		return ""
	}
	_index := 0
	targetAddress := n.nameTree.FLoopGroup(groupname, func(height int, node core.NodeInfo) bool {
		if _index == index {
			return true
		}
		_index++
		return false
	})
	return targetAddress
}
func (n *namesrv) HandleClients() {
	n.referee.AddClientHandler(func(so socketio.Socket) {
		so.On("refer address", func(name string, index int) {
			addr := n.SearchNodeInfoWithGroupName(name, index)
			so.Emit("refer address", addr)
		})
	})
}
func (n *namesrv) HandleMessages() {
	n.looper.AddHandler(0, core.MA_Refer, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Ask:
			nodeinfo, err := n.SearchNodeInfo(core.NodePath(msg.GetContent()))
			if err != nil {
				return err
			}
			msg.SetContent([]byte(nodeinfo))
			msg.GetInfo().SetState(core.MS_Succeed)
			err = n.referee.SendToWroker(msg, nodeinfo)
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
