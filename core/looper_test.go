package core

import (
	"errors"
	//"fmt"
	utils "github.com/GargouillePao/flitter/utils"
	"testing"
	"time"
)

func TestMessageLooper(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Msg Looper"))
	looper := NewMessageLooper(10)

	looper.AddHandler(1000, MA_Refer, func(msg Message) error {
		_, state, _ := msg.GetInfo().Info()
		var err error
		switch state {
		case MS_Failed:
			t.Log(utils.Infof("%v", msg))
		case MS_Error:
			msg.GetInfo().SetAcion(MA_Term)
			looper.Push(msg)
		default:
			//t.Log(utils.Infof("%v", msg))
			err = errors.New("MLooper Error")
		}
		return err
	})
	times := 0
	looper.AddHandler(0, MA_Heartbeat, func(msg Message) error {
		_, state, _ := msg.GetInfo().Info()
		var err error
		switch state {
		case MS_Probe:
			times++
			//fmt.Println(utils.Infof("%v", msg))
			if times == 5 {
				msg := NewMessage(NewMessageInfo(), []byte("Hello"))
				msg.GetInfo().SetAcion(MA_Refer)
				looper.Push(msg)
			}
		}
		return err
	})
	looper.SetInterval(1000, func(_t time.Time) error {
		msg := NewMessage(NewMessageInfo(), []byte(_t.String()))
		msg.GetInfo().SetAcion(MA_Heartbeat)
		msg.GetInfo().SetState(MS_Probe)
		looper.Push(msg)
		//fmt.Println(utils.Norf("%v", msg))
		return nil
	})
	looper.Loop()
	t.Log(utils.Norf("End Msg Looper"))
}
