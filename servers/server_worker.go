package servers

import (
	"errors"
	"github.com/gargous/flitter/common"
	"github.com/gargous/flitter/core"
	"sync"
)

type Worker interface {
	SendToReferee(msg core.Message, npath core.NodePath) error
	SendToWroker(msg core.Message, npath core.NodePath) error
	PublishToWorker(msg core.Message) error
	SubscribeWorker(npath core.NodePath) error
	Server
	common.DataSet
}

type workersrv struct {
	recverR2W  core.Receiver
	senderW2R  core.Sender
	recverW2W  core.Receiver
	senderW2W  core.Sender
	subscriber core.Subscriber
	publisher  core.Publisher
	wg         sync.WaitGroup
	mutex      sync.Mutex
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
	_worker.Set("path", common.DataItem{Data: []byte(npath)})
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
	w.mutex.Lock()
	defer w.mutex.Unlock()
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
	w.mutex.Lock()
	defer w.mutex.Unlock()
	info, err := _ParseAddress(npath, SRT_Worker, SRT_Worker)
	if err != nil {
		return
	}
	w.senderW2W.Disconnect(true)
	w.senderW2W.AddNodeInfo(info)
	err = w.senderW2W.Connect()
	if err != nil {
		return
	}
	err = w.senderW2W.Send(msg)
	return
}
func (w *workersrv) PublishToWorker(msg core.Message) error {
	return w.publisher.Send(msg)
}
func (w *workersrv) SubscribeWorker(npath core.NodePath) (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
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
	if err != nil {
		return
	}
	err = w.recverW2W.Bind()
	if err != nil {
		return
	}
	err = w.publisher.Bind()
	if err != nil {
		return
	}
	go func() {
		for {
			msg, err := w.recverR2W.Recv()
			if err != nil {
				common.ErrIn(err, "Receive From Referee")
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
				common.ErrIn(err, "Receive From Worker")
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
				common.ErrIn(err, "Subscribe From Worker")
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
	index := 0
	for _, srvice := range w.srvices {
		func(srvice Service) {
			go func() {
				defer w.wg.Done()
				err = srvice.Init(w)
				if err != nil {
					return
				}
				common.Logf(common.Norf, "Initiate %v", srvice)
				index++
				if index >= len(w.srvices) {
					dpath, ok := w.Get("path")
					if !ok {
						err = errors.New("Path Not Config")
						return
					}
					common.Logf(common.Norf, "Worker Started At %v\n%v", core.NodePath(dpath.Data), w)
				}
				srvice.Start()
			}()
		}(srvice)
	}
	w.subscriber.SetSubscribe("")
	err = w.InitClientHandler(nil)
	if err != nil {
		common.ErrIn(err)
		return
	}
	w.wg.Wait()
	return
}
