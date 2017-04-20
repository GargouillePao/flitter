package servers

import (
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/utils"
	"sync"
)

type Referee interface {
	SendToWroker(msg core.Message, node core.NodeInfo) error
	Server
}
type refereesrv struct {
	recverW2R core.Receiver
	senderR2W core.Sender
	wg        sync.WaitGroup
	baseServer
}

func NewReferee(name string, addr string) (referee Referee, err error) {
	senderR2W, err := core.NewSender()
	if err != nil {
		return
	}
	_referee := &refereesrv{
		senderR2W: senderR2W,
	}
	_referee.addr = addr
	_referee.name = name
	_referee.srvices = make(map[ServiceType]Service)
	_addr, err := _ParseAddress(core.NodeInfo(addr), SRT_Worker, SRT_Referee)
	if err != nil {
		return
	}
	recverW2R, err := core.NewReceiver(_addr)
	if err != nil {
		return
	}
	_referee.recverW2R = recverW2R
	referee = _referee
	return
}

func (r *refereesrv) SendToWroker(msg core.Message, node core.NodeInfo) (err error) {
	addr, err := _ParseAddress(node, SRT_Referee, SRT_Worker)
	if err != nil {
		return
	}
	r.senderR2W.Disconnect(true)
	r.senderR2W.AddNodeInfo(core.NodeInfo(addr))
	err = r.senderR2W.Connect()
	err = r.senderR2W.Send(msg)
	return
}

func (r *refereesrv) Start() (err error) {
	err = r.recverW2R.Bind()
	if err != nil {
		return
	}
	go func() {
		for {
			msg, err := r.recverW2R.Recv()
			if err != nil {
				utils.ErrIn(err, "Receive From Worker", "At Referee_"+r.name)
				continue
			}
			if msg != nil {
				for _, srvice := range r.srvices {
					srvice.Push(msg)
				}
			}
		}
	}()
	r.wg.Add(len(r.srvices))
	for _, srvice := range r.srvices {
		utils.Logf(utils.Norf, "%v", srvice)
		func(srvice Service) {
			go func() {
				defer r.wg.Done()
				srvice.Init(r)
				srvice.Start()
				utils.Logf(utils.Norf, "Start")
			}()
		}(srvice)
	}
	err = r.InitClientHandler()
	if err != nil {
		return
	}
	r.wg.Wait()
	return
}
