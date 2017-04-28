package servers

import (
	"errors"
	"fmt"
	"github.com/gargous/flitter/core"
	saver "github.com/gargous/flitter/save"
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
	nameTreeSaver saver.NodeTreeSaver
	mutex         sync.Mutex
	bussyness     bool
	baseService
}

func NewNameService() NameService {
	srv := &namesrv{
		nameTree:  core.NewNodeTree(),
		bussyness: false,
	}
	srv.looper = core.NewMessageLooper(__LooperSize)
	srv.nameTreeSaver = saver.NewNodeTreeSaver(srv.nameTree)
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
		n.nameTreeSaver = saver.NewNodeTreeSaver(n.nameTree)
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
		//n.nameTreeSaver.SaveLastItem(saver.SA_Add)
	}
	utils.Logf(utils.Norf, "Search\n%v,%v", opath, n.nameTree)
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
	n.referee.TrickClient("flitter refer address", func(so socketio.Socket) interface{} {
		return func(name string, index int) {
			if !n.bussyness {
				addr := n.SearchNodeInfoWithGroupName(name, index)
				so.Emit("flitter refer address", addr)
			} else {
				so.Emit("flitter refer address", __Client_Reply_bussy)
			}
		}
	})
}
func (n *namesrv) HandleMessages() {
	n.looper.AddHandler(0, core.MA_Refer, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Ask:
			if !n.bussyness {
				content, ok := msg.GetContent(0)
				if !ok {
					return
				}
				nodeinfo, err := n.SearchNodeInfo(core.NodePath(content))
				if err != nil {
					return err
				}
				msg.ClearContent()
				msg.AppendContent([]byte(nodeinfo))
				msg.GetInfo().SetState(core.MS_Succeed)
				err = n.referee.SendToWroker(msg, nodeinfo)
				if err != nil {
					return err
				}
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
