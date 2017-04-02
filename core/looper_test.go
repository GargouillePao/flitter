package core

import (
	"errors"
	utils "github.com/GargouillePao/flitter/utils"
	"testing"
	"time"
)

func TestMessageLooper(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Msg Looper"))
	looper := NewMessageLooper(10)
	looper.Loop(false)
	looper.AddHandler(MA_JoinGlobal, func(msg Message) error {
		_, state, _ := msg.GetInfo().Info()
		var err error
		switch state {
		case MS_Error:
			msg.GetInfo().SetAcion(MA_Terminal)
			looper.Push(msg)
		default:
			t.Log(utils.Infof("%v", msg))
			err = errors.New("MLooper Error")
		}
		return err
	})
	times := 0
	looper.AddHandler(MA_Heartbeat, func(msg Message) error {
		_, state, _ := msg.GetInfo().Info()
		var err error
		switch state {
		case MS_Probe:
			times++
			t.Log(utils.Infof("%v", msg))
			if times == 3 {
				msg := NewMessage(NewMessageInfo(), []byte("Hello"))
				msg.GetInfo().SetAcion(MA_JoinGlobal)
				looper.Push(msg)
			}
		}
		return err
	})

	looper.SetInterval(1000, func(t time.Time) error {
		msg := NewMessage(NewMessageInfo(), []byte(t.String()))
		msg.GetInfo().SetAcion(MA_Heartbeat)
		msg.GetInfo().SetState(MS_Probe)
		looper.Push(msg)
		return nil
	})
	looper.Wait()
	t.Log(utils.Norf("End Msg Looper"))
}
