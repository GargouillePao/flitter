package servers

import (
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/utils"
	"sync"
)

type Worker interface {
	SendToReferee(msg core.Message, npath core.NodePath) error
	SendToWroker(msg core.Message, npath core.NodePath) error
	PublishToWorker(msg core.Message) error
	SubscribeWorker(npath core.NodePath) error
	Server
}

type workersrv struct {
	recverR2W  core.Receiver
	senderW2R  core.Sender
	recverW2W  core.Receiver
	senderW2W  core.Sender
	subscriber core.Subscriber
	publisher  core.Publisher
	wg         sync.WaitGroup
	baseServer
}

func NewWorker(npath core.NodePath) (worker Worker, err error) {
	senderW2R, err := core.NewSender()
	if err != nil {
		return
	}
	senderW2W, err := core.NewSender()
	if err != nil {
		return
	}
	subscriber, err := core.NewSubscriber()
	if err != nil {
		return
	}
	_worker := &workersrv{
		senderW2R:  senderW2R,
		senderW2W:  senderW2W,
		subscriber: subscriber,
	}
	_worker.path = npath
	_worker.srvices = make(map[ServiceType]Service)

	recverR2WAddr, err := _ParseAddress(npath, SRT_Referee, SRT_Worker)
	if err != nil {
		return
	}
	recverR2W, err := core.NewReceiver(recverR2WAddr)
	if err != nil {
		return
	}

	recverW2WAddr, err := _ParseAddress(npath, SRT_Worker, SRT_Worker)
	if err != nil {
		return
	}
	recverW2W, err := core.NewReceiver(recverW2WAddr)
	if err != nil {
		return
	}

	publisherAddr, err := _ParseAddress(npath, SRT_Workers, SRT_Workers)
	if err != nil {
		return
	}
	publisher, err := core.NewPublisher(publisherAddr)
	if err != nil {
		return
	}

	_worker.recverR2W = recverR2W
	_worker.recverW2W = recverW2W
	_worker.publisher = publisher
	worker = _worker
	return
}

func (w *workersrv) SendToReferee(msg core.Message, npath core.NodePath) (err error) {
	info, err := _ParseAddress(npath, SRT_Worker, SRT_Referee)
	if err != nil {
		return
	}
	w.senderW2R.Disconnect(true)
	w.senderW2R.AddNodeInfo(info)
	err = w.senderW2R.Connect()
	err = w.senderW2R.Send(msg)
	return
}
func (w *workersrv) SendToWroker(msg core.Message, npath core.NodePath) (err error) {
	info, err := _ParseAddress(npath, SRT_Worker, SRT_Worker)
	if err != nil {
		return
	}
	w.senderW2W.Disconnect(true)
	w.senderW2W.AddNodeInfo(info)
	err = w.senderW2W.Connect()
	err = w.senderW2W.Send(msg)
	return
}
func (w *workersrv) PublishToWorker(msg core.Message) error {
	return w.publisher.Send(msg)
}
func (w *workersrv) SubscribeWorker(npath core.NodePath) (err error) {
	info, err := _ParseAddress(npath, SRT_Workers, SRT_Workers)
	if err != nil {
		return
	}
	w.subscriber.Disconnect(true)
	w.subscriber.AddNodeInfo(info)
	err = w.subscriber.Connect()
	return
}
func (w *workersrv) Start() (err error) {
	err = w.recverR2W.Bind()
	err = w.recverW2W.Bind()
	err = w.publisher.Bind()
	if err != nil {
		return
	}
	go func() {
		for {
			msg, err := w.recverR2W.Recv()
			if err != nil {
				utils.ErrIn(err, "Receive From Referee", "At Wroker_"+string(w.path))
				continue
			}
			if msg != nil {
				for _, srvice := range w.srvices {
					srvice.Push(msg)
				}
			}
		}
	}()
	go func() {
		for {
			msg, err := w.recverW2W.Recv()
			if err != nil {
				utils.ErrIn(err, "Receive From Worker", "At Wroker_"+string(w.path))
				continue
			}
			if msg != nil {
				for _, srvice := range w.srvices {
					srvice.Push(msg)
				}
			}
		}
	}()
	go func() {
		for {
			msg, err := w.subscriber.Recv()
			if err != nil {
				utils.ErrIn(err, "Subscribe From Worker", "At Wroker_"+string(w.path))
				continue
			}
			if msg != nil {
				for _, srvice := range w.srvices {
					srvice.Push(msg)
				}
			}
		}
	}()
	w.wg.Add(len(w.srvices))
	for _, srvice := range w.srvices {
		func(srvice Service) {
			go func() {
				srvice.Init(w)
				srvice.Start()
				w.wg.Done()
			}()
		}(srvice)
	}
	w.subscriber.SetSubscribe("")
	err = w.InitClientHandler()
	if err != nil {
		return
	}
	w.wg.Wait()
	return
}
