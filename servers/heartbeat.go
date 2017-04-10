package servers

import (
	"errors"
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	//zmq "github.com/pebbe/zmq4"
	//"os"
	//"time"
)

type HeartbeatServer interface {
	Server
}

type heartbeatsrv struct {
	notifierForWatch   core.Deliverer
	notifierForService core.Deliverer
	senderToQuorum     core.Sender
	recverFromQuorum   core.Receiver
	looper             core.MessageLooper
}

func NewHeartbeatServer() HeartbeatServer {
	return &heartbeatsrv{
		looper: core.NewMessageLooper(__LooperSize),
	}
}

func (h *heartbeatsrv) Config(ca ConfigAction, st ServerType, info core.NodeInfo) error {
	switch st {
	case ST_Watch:
		switch ca {
		case CA_Send:
		case CA_Recv:
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
	return errors.New("ERR")
}

func (h *heartbeatsrv) Init() error {
	h.HandleRecive()
	h.HandleMessages()
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
		if h.notifierForWatch != nil {
			for {
				msg, err := h.notifierForWatch.Recv()
				if err != nil {
					utils.ErrIn(err, "[receive at heartbeat from watch]")
					continue
				}
				h.looper.Push(msg)
			}
		}
	}()
}
func (h *heartbeatsrv) HandleMessages() {
	h.looper.AddHandler(0, core.MA_Init, func(msg core.Message) error {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Ask:
			fmt.Println(utils.Infof("Succed and I am binding wait for heartbeat"))
		case core.MS_Error:
			utils.ErrIn(errors.New(msg.GetInfo().String()), "[heartbeat server]")
		}
		return nil
	})
}
func (h *heartbeatsrv) Start() {
	lauchServer()
	h.looper.Loop()
}
func (h *heartbeatsrv) Term() {
	h.Term()
}
