package core

import (
	"bytes"
	utils "github.com/gargous/flitter/utils"
	"github.com/kr/pretty"
	"testing"
	"time"
)

func Test_MessageInfo(t *testing.T) {

	t.Log(utils.Norf("Start MessageInfo"))
	info := NewMessageInfo()
	info.SetAcion(MA_Refer)
	info.SetState(MS_Probe)
	nowTime := time.Now()
	info.SetTime(nowTime)
	action, state, _time := info.Info()
	if action == MA_Refer && state == MS_Probe && info.Size() == 10 && _time.Equal(nowTime) {
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
	info.SetAcion(MA_Refer)
	info.SetState(MS_Probe)
	nowTime := time.Now()
	info.SetTime(nowTime)
	serializer := NewSerializer()
	_, buf, err := serializer.Encode(info)
	if err != nil {
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}
	if len(buf) < 2 || buf[0] != byte(MA_Refer) || buf[1] != byte(MS_Probe) {
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
	if action != MA_Refer || state != MS_Probe || !nowTime.Equal(_time) {
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
	info.SetAcion(MA_Refer)
	info.SetState(MS_Probe)
	content := []byte("Hello")
	msg := NewMessage(info)
	msg.AppendContent(content)
	serializer := NewSerializer()
	_, buf, _ := serializer.Encode(info)
	buf = append(buf, bytes.Join(msg.GetContents(), []byte("\n"))...)
	ninfo := NewMessageInfo()
	nmsg := NewMessage(ninfo)
	n, err := serializer.Decode(ninfo, buf)
	if err != nil {
		t.Log(utils.Errf("%v", err))
		t.Fail()
	}
	ncontent := buf[n:]
	nmsg.SetContents(bytes.Split(ncontent, []byte("\n")))
	diff := pretty.Diff(nmsg, msg)
	if len(diff) > 0 {
		t.Logf(utils.Errf("Err and now info is:%v.%v.%v", nmsg, msg, diff))
		t.Fail()
	} else {
		t.Logf(utils.Infof("msg now is:%v", nmsg))
	}
	t.Log(utils.Norf("End MsgInfo Serialize"))
}
