package core

import (
	utils "github.com/GargouillePao/flitter/utils"
	"testing"
	"time"
)

func Test_MessageInfo(t *testing.T) {
	t.Log(utils.Norf("Start MessageInfo"))
	info := NewMessageInfo()
	info.SetAcion(MA_JoinGlobal)
	info.SetState(MS_Probe)
	nowTime := time.Now()
	info.SetTime(nowTime)
	action, state, _time := info.Info()
	if action == MA_JoinGlobal && state == MS_Probe && info.Size() == 10 && _time.Equal(nowTime) {
		t.Logf(utils.Infof("MessageInfo create succeed and now it's:%v", info))
	} else {
		t.Logf(utils.Errf("MessageInfo create failed and now it's:%v", info))
		t.Fail()
	}
	t.Log(utils.Norf("End MessageInfo"))
}

func Test_MessageInfo_Serialize(t *testing.T) {
	t.Log(utils.Norf("Start MsgInfo Serialize"))
	info := NewMessageInfo()
	info.SetAcion(MA_JoinGlobal)
	info.SetState(MS_Probe)
	nowTime := time.Now()
	info.SetTime(nowTime)
	serializer := NewSerializer()
	_, buf, err := serializer.Encode(info)
	if err != nil {
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}
	if len(buf) < 2 || buf[0] != byte(MA_JoinGlobal) || buf[1] != byte(MS_Probe) {
		t.Logf(utils.Errf("Err and now info is:%v; buf is:%v", info, buf))
		t.Fail()
	} else {
		t.Logf(utils.Infof("buf now is:%v", buf))
		t.Logf(utils.Infof("msg now is:%v", info))
	}
	info = NewMessageInfo()
	_, err = serializer.Decode(info, buf)
	if err != nil {
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}
	action, state, _time := info.Info()
	if action != MA_JoinGlobal || state != MS_Probe || !nowTime.Equal(_time) {
		t.Logf(utils.Errf("Err and now info is:%v", info))
		t.Fail()
	} else {
		t.Logf(utils.Infof("msg now is:%v", info))
		t.Logf(utils.Infof("buf now is:%v", buf))
	}
	t.Log(utils.Norf("End MsgInfo Serialize"))
}

func Test_Message_Serialize(t *testing.T) {
	t.Log(utils.Norf("Start MsgInfo Serialize"))
	info := NewMessageInfo()
	info.SetAcion(MA_JoinGlobal)
	info.SetState(MS_Probe)
	content := []byte("Hello")
	msg := NewMessage(info, content)
	serializer := NewSerializer()
	_, buf, _ := serializer.Encode(info)
	buf = append(buf, msg.GetContent()...)

	info = NewMessageInfo()
	n, err := serializer.Decode(info, buf)
	if err != nil {
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}
	ncontent := buf[n:]
	if string(ncontent) != string(content) {
		t.Logf(utils.Errf("Err and now info is:%s.%v.%v", ncontent, n, info.Size()))
		t.Fail()
	} else {
		t.Logf(utils.Infof("msg now is:%s", ncontent))
	}
	t.Log(utils.Norf("End MsgInfo Serialize"))
}
