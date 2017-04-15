package servers

import (
	"errors"
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	"strings"
	"time"
)

type HeartbeatServer interface {
	Server
}

type heartbeatsrv struct {
	senderToWatch       core.Sender
	recverFromWatch     core.Receiver
	notifierForService  core.Deliverer
	senderToQuorum      core.Sender
	recverFromQuorum    core.Receiver
	publisherHeartbeat  core.Publisher
	subscriberHeartbeat core.Subscriber
	looper              core.MessageLooper
}

func NewHeartbeatServer() HeartbeatServer {
	return &heartbeatsrv{
		looper: core.NewMessageLooper(__LooperSize),
	}
}

func (h *heartbeatsrv) Config(ca ConfigAction, st ServerType, addr string) (err error) {
	switch st {
	case ST_Watch:
		var watchaddr string
		switch ca {
		case CA_Recv:
			watchaddr, err = transAddress(addr, ST_Watch, ST_HeartBeat)
			if err != nil {
				return
			}
			h.recverFromWatch, err = core.NewReceiver(watchaddr)
			if err != nil {
				return
			}
		case CA_Send:
			h.senderToWatch, err = core.NewSender()
			if err != nil {
				return
			}
			watchaddr, err = transAddress(addr, ST_HeartBeat, ST_Watch)
			if err != nil {
				return
			}
			h.senderToWatch.AddNodeInfo(core.NodeInfo(watchaddr))
		}
	case ST_HeartBeat:
		switch ca {
		case CA_Send:
			var heartaddr string
			heartaddr, err = transAddress(addr, ST_HeartBeat, ST_HeartBeat)
			if err != nil {
				return
			}
			h.publisherHeartbeat, err = core.NewPublisher(heartaddr)
			if err != nil {
				return
			}
		case CA_Recv:
			h.subscriberHeartbeat, err = core.NewSubscriber()
			if err != nil {
				return
			}
		}
	case ST_Quorum:
		switch ca {
		case CA_Send:
		case CA_Recv:
		}
	case ST_Service:
		switch ca {
		case CA_Send:
		case CA_Recv:
		}
	}
	return
}

func (h *heartbeatsrv) Init() error {
	var err error
	if h.publisherHeartbeat == nil || h.senderToWatch == nil || h.recverFromWatch == nil || h.subscriberHeartbeat == nil {
		err = errors.New("Config not enough")
		return err
	}
	h.HandleRecive()
	h.HandleMessages()
	err = h.publisherHeartbeat.Bind()
	err = h.recverFromWatch.Bind()
	return nil
}
func (h *heartbeatsrv) HandleRecive() {
	go func() {
		if h.recverFromQuorum != nil {
			for {
				msg, err := h.recverFromQuorum.Recv()
				if err != nil {
					utils.ErrIn(err, "[receive at heartbeat from quorum]")
					continue
				}
				h.looper.Push(msg)
			}
		}
	}()
	go func() {
		if h.notifierForService != nil {
			for {
				msg, err := h.notifierForService.Recv()
				if err != nil {
					utils.ErrIn(err, "[receive at heartbeat from service]")
					continue
				}
				h.looper.Push(msg)
			}
		}
	}()
	go func() {
		if h.recverFromWatch != nil {
			for {
				msg, err := h.recverFromWatch.Recv()
				if err != nil {
					utils.ErrIn(err, "[receive at heartbeat from watch]")
					continue
				}
				h.looper.Push(msg)
			}
		}
	}()
	go func() {
		if h.subscriberHeartbeat != nil {
			for {
				msg, err := h.subscriberHeartbeat.Recv()
				if err != nil {
					utils.ErrIn(err, "[subscribe at heartbeat from Leader]")
					continue
				}
				h.looper.Push(msg)
			}
		}
	}()
}
func (h *heartbeatsrv) HandleMessages() {
	if h.publisherHeartbeat != nil {
		h.looper.SetInterval(3000, func(t time.Time) error {
			info := core.NewMessageInfo()
			info.SetAcion(core.MA_Heartbeat)
			info.SetState(core.MS_Succeed)
			err := h.publisherHeartbeat.Send(core.NewMessage(info, []byte("")))
			if err != nil {
				utils.ErrIn(err, "[publish at heartbeat to Children]")
			}
			return nil
		})
	}
	h.looper.AddHandler(0, core.MA_Init_Heartbeat, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			var leader core.NodeInfo
			leader, err = core.NodeInfo(msg.GetContent()).GetLeaderInfo()
			if err != nil {
				return
			}
			if leader == "" {
				msgInfo := msg.GetInfo()
				msgInfo.SetState(core.MS_Succeed)
				h.looper.Push(core.NewMessage(msgInfo, []byte("")))
			} else {
				msg.GetInfo().SetState(core.MS_Ask)
				msg.SetContent([]byte(leader))
				h.looper.Push(msg)
			}

		case core.MS_Ask:

			var leaderaddr string
			leaderaddr, err = transAddress(string(msg.GetContent()), ST_HeartBeat, ST_HeartBeat)
			h.subscriberHeartbeat.Disconnect(true)
			h.subscriberHeartbeat.AddNodeInfo(core.NodeInfo(leaderaddr))
			err = h.subscriberHeartbeat.Connect()
			if err != nil {
				return
			}
			msgInfo := msg.GetInfo()
			msgInfo.SetState(core.MS_Succeed)
			h.looper.Push(core.NewMessage(msgInfo, []byte("")))

			msgInfo = core.NewMessageInfo()
			msgInfo.SetAcion(core.MA_Heartbeat)
			msgInfo.SetState(core.MS_Probe)
			h.looper.Push(core.NewMessage(msgInfo, []byte("")))

		case core.MS_Succeed:
			h.senderToWatch.Connect()
			h.senderToWatch.Send(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "[heartbeat server when init]")
		}
		return
	})
	h.looper.AddHandler(5000, core.MA_Heartbeat, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Succeed:
			msg.GetInfo().SetTime(time.Now())
			utils.Logf(utils.Infof, "Heartbeating succeed and now %v", msg)
			msg.GetInfo().SetState(core.MS_Probe)
			h.looper.Push(msg)
		case core.MS_Failed:
			utils.Logf(utils.Warningf, "Heartbeating faild and now %v", msg)
			msg.GetInfo().SetState(core.MS_Probe)
			h.looper.Push(msg)
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "[heartbeat server when heartbeating]")
		}
		return
	})
}
func (h *heartbeatsrv) Start() {
	h.looper.Loop()
}
func (h *heartbeatsrv) Term() {
	h.Term()
}
func (h heartbeatsrv) String() string {

	str := fmt.Sprintf("Heartbeat Server:["+
		"\n\tlooper:%p"+
		"\n\tquorum receiver:%v"+
		"\n\tquorum sender:%v"+
		"\n\twatch receiver:%v"+
		"\n\twatch sender:%v"+
		"\n\tservice notifier:%v"+
		"\n\theart publisher:%v"+
		"\n\theart subscriber:%v"+
		"]",
		h.looper,
		strings.Join(strings.Split(fmt.Sprintf("%v", h.recverFromQuorum), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", h.senderToQuorum), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", h.recverFromWatch), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", h.senderToWatch), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", h.notifierForService), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", h.publisherHeartbeat), "\n"), "\n\t"),
		strings.Join(strings.Split(fmt.Sprintf("%v", h.subscriberHeartbeat), "\n"), "\n\t"),
	)
	return str
}
