package servers

import (
	"errors"
	"fmt"
	core "github.com/gargous/flitter/core"
	utils "github.com/gargous/flitter/utils"
	//"strings"
	"time"
)

type HeartbeatService interface {
	Service
}
type heartbeatsrv struct {
	worker Worker
	baseService
}

func NewHeartbeatService() HeartbeatService {
	service := &heartbeatsrv{}
	service.looper = core.NewMessageLooper(__LooperSize)
	return service
}
func (h *heartbeatsrv) Init(srv interface{}) error {
	h.worker = srv.(Worker)
	h.HandleMessages()
	return nil
}

func (h *heartbeatsrv) HandleMessages() {
	if h.worker != nil {
		h.looper.SetInterval(3000, func(t time.Time) error {
			info := core.NewMessageInfo()
			info.SetAcion(core.MA_Heartbeat)
			info.SetState(core.MS_Succeed)
			err := h.worker.PublishToWorker(core.NewMessage(info))
			if err != nil {
				utils.ErrIn(err, "[publish at heartbeat to Children]")
			}
			return nil
		})
	}
	h.looper.AddHandler(0, core.MA_Init, func(msg core.Message) (err error) {
		_, state, _ := msg.GetInfo().Info()
		switch state {
		case core.MS_Probe:
			content, ok := msg.GetContent(0)
			if !ok {
				return
			}
			leader, ok := core.NodePath(content).GetLeaderPath()
			if !ok {
				msgInfo := msg.GetInfo()
				msgInfo.SetState(core.MS_Succeed)
				h.looper.Push(core.NewMessage(msgInfo))
			} else {
				msg.GetInfo().SetState(core.MS_Ask)
				msg.AppendContent([]byte(leader))
				h.looper.Push(msg)
			}

		case core.MS_Ask:
			content, ok := msg.GetContent(0)
			if !ok {
				return
			}
			lpath, ok := core.NodePath(content).GetLeaderPath()
			if !ok {
				err = errors.New("No Leader")
				return
			}
			err = h.worker.SubscribeWorker(lpath)
			if err != nil {
				return
			}
			msgInfo := msg.GetInfo()
			msgInfo.SetState(core.MS_Succeed)
			h.looper.Push(core.NewMessage(msgInfo))

			msgInfo = core.NewMessageInfo()
			msgInfo.SetAcion(core.MA_Heartbeat)
			msgInfo.SetState(core.MS_Probe)
			h.looper.Push(core.NewMessage(msgInfo))

		case core.MS_Succeed:
			h.worker.SendService(ST_Watch, msg)
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
			//utils.Logf(utils.Infof, "Heartbeating succeed and now %v", msg)
			msg.GetInfo().SetState(core.MS_Probe)
			h.looper.Push(msg)
		case core.MS_Failed:
			msg.GetInfo().SetTime(time.Now())
			utils.Logf(utils.Warningf, "Heartbeating faild and now %v", msg)
			if msg.GetVisitTimes() < 3 {
				msg.GetInfo().SetState(core.MS_Probe)
				h.looper.Push(msg)
			} else {
				utils.Logf(utils.Errf, "Heartbeating maybe dead and now %v", msg)
			}

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
	str := fmt.Sprintf("Heartbeat Service:["+"looper:%p"+"]", h.looper)
	return str
}
