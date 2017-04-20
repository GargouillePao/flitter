package core

import (
	"errors"
	"fmt"
	utils "github.com/gargous/flitter/utils"
	"testing"
	"time"
)

func TestMessageLooper(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Msg Looper"))
	looper := NewMessageLooper(10)
	failTimes := 0
	looper.AddHandler(1000, MA_Refer, func(msg Message) error {
		_, state, _ := msg.GetInfo().Info()
		var err error
		switch state {
		case MS_Probe:
			fmt.Println(utils.Norf("%v", msg))
			time.Sleep(time.Second * 3)
			msg.GetInfo().SetState(MS_Succeed)
			looper.Push(msg)
		case MS_Failed:
			failTimes++
			msg.GetInfo().SetTime(time.Now())
			if failTimes >= 3 {
				err = errors.New("MLooper Error")
			} else {
				fmt.Println(utils.Warningf("%v", msg))
				msg.GetInfo().SetState(MS_Probe)
				looper.Push(msg)
			}
		case MS_Succeed:
			fmt.Println(utils.Infof("%v", msg))
		case MS_Error:
			fmt.Println(utils.Errf("%v", msg))
			msg.GetInfo().SetAcion(MA_Term)
			looper.Push(msg)
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
			if times == 3 {
				msg := NewMessage(NewMessageInfo(), []byte("Hello"))
				msg.GetInfo().SetAcion(MA_Refer)
				msg.GetInfo().SetState(MS_Probe)
				looper.Push(msg)
			}
			if times == 6 {
				msg := NewMessage(NewMessageInfo(), []byte("World"))
				msg.GetInfo().SetAcion(MA_Refer)
				msg.GetInfo().SetState(MS_Probe)
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
