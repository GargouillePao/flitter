package servers

import (
	"github.com/gargous/flitter/common"
	"github.com/gargous/flitter/core"
	"sync"
)

type Referee interface {
	SendToWroker(msg core.Message, npath core.NodePath) error
	Server
}
type refereesrv struct {
	recverW2R core.Receiver
	senderR2W core.Sender
	wg        sync.WaitGroup
	baseServer
}

func NewReferee(npath core.NodePath) (referee Referee, err error) {
	senderR2W, err := core.NewSender()
	if err != nil {
		return
	}
	_referee := &refereesrv{
		senderR2W: senderR2W,
	}
	_referee.SetPath(npath)
	_referee.srvices = make(map[ServiceType]Service)
	info, err := _ParseAddress(npath, SRT_Worker, SRT_Referee)
	if err != nil {
		return
	}
	recverW2R, err := core.NewReceiver(info)
	if err != nil {
		return
	}
	_referee.recverW2R = recverW2R
	referee = _referee
	return
}

func (r *refereesrv) SendToWroker(msg core.Message, npath core.NodePath) (err error) {
	info, err := _ParseAddress(npath, SRT_Referee, SRT_Worker)
	if err != nil {
		return
	}
	r.senderR2W.Disconnect(true)
	r.senderR2W.AddNodeInfo(info)
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
				common.ErrIn(err, "Receive From Worker")
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
	index := 0
	for _, srvice := range r.srvices {
		func(srvice Service) {
			go func() {
				defer r.wg.Done()
				err = srvice.Init(r)
				if err != nil {
					return
				}
				common.Logf(common.Norf, "Initiate %v", srvice)
				index++
				if index >= len(r.srvices) {
					dpath := r.GetPath()
					common.Logf(common.Norf, "Referee Started At %v\n%v", dpath, r)
				}
				srvice.Start()
			}()
		}(srvice)
	}
	if err != nil {
		return
	}

	err = r.InitClientHandler(nil)
	if err != nil {
		return
	}
	//r.wg.Wait()
	return
}
