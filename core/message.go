package core

import (
	"errors"
	"fmt"
)

/*(NULL)*/
type MessageAction uint8

const (
	_ MessageAction = iota
	MA_Undefine
	MA_JoinGlobal
	MA_Join
	MA_Invite
	MA_Heartbeat
	MA_Crash
	MA_Vote
	MA_Upgrade
	MA_User_Request
)

func (m MessageAction) Normalize() MessageAction {
	if m > 9 {
		m = MA_Undefine
	}
	return m
}

func (m MessageAction) String() (str string) {
	switch m {
	case MA_Undefine:
		return "MA_Undefine"
	case MA_JoinGlobal:
		return "MA_JoinGlobal"
	case MA_Join:
		return "MA_Join"
	case MA_Invite:
		return "MA_Invite"
	case MA_Heartbeat:
		return "MA_Heartbeat"
	case MA_Crash:
		return "MA_Crash"
	case MA_Vote:
		return "MA_Vote"
	case MA_Upgrade:
		return "MA_Upgrade"
	case MA_User_Request:
		return "MA_User_Request"
	}
	return ""
}

/*(NULL)*/
type Message interface {
	GetInfo() (info MessageInfo)
	SetContent(buf []byte)
	GetContent() (buf []byte)
}

func NewMessage(info MessageInfo, content []byte) Message {
	msg := &message{info: info, content: content}
	return msg
}

/*(NULL)*/
type message struct {
	info    MessageInfo
	content []byte
}

func (m *message) GetInfo() (info MessageInfo) {
	info = m.info
	return
}

func (m *message) SetContent(buf []byte) {
	m.content = buf
	return
}
func (m *message) GetContent() (buf []byte) {
	buf = m.content
	return
}

type MessageInfo interface {
	Serializable
	String() string
	SetAcion(action MessageAction)
	SetState(state MessageState)
	Info() (action MessageAction, state MessageState)
}

func NewMessageInfo() MessageInfo {
	return &messageInfo{
		action: MA_Undefine,
		state:  MS_Undefine,
		size:   2,
	}
}

/*the data of message head*/
type messageInfo struct {
	action MessageAction
	state  MessageState
	size   int
}

/** read out to the buf */
func (m *messageInfo) Read(buf []byte) (int, error) {
	var err error
	if len(buf) < m.Size() {
		err = errors.New("Invalid message info to read")
		return 0, err
	}
	buf[0] = byte(m.action)
	buf[1] = byte(m.state)
	return m.Size(), nil
}

/** write in from the buf */
func (m *messageInfo) Write(buf []byte) (int, error) {
	var err error
	if len(buf) < m.Size() {
		err = errors.New("Invalid message info to write")
		return 0, err
	}
	m.action = MessageAction(buf[0]).Normalize()
	buf = buf[1:]
	m.state = MessageState(buf[0]).Normalize()
	return m.Size(), nil
}

func (m *messageInfo) SetAcion(action MessageAction) {
	m.action = action
	return
}

func (m *messageInfo) SetState(state MessageState) {
	m.state = state
	return
}

func (m *messageInfo) Info() (action MessageAction, state MessageState) {
	action = m.action
	state = m.state
	return
}

func (m messageInfo) String() string {
	str := fmt.Sprintf("MsgInfo:[action:%v,state:%v]", m.action, m.state)
	return str
}
func (m messageInfo) Size() int {
	return m.size
}
func (m messageInfo) Resize(size int) {
	m.size = size
}

/*(NULL)*/
type MessageState uint8

const (
	_ MessageState = iota
	MS_Undefine
	MS_Probe
	MS_Ask
	MS_Succeed
	MS_Failed
)

func (m MessageState) Normalize() MessageState {
	if m > 5 {
		m = MS_Undefine
	}
	return m
}

func (m MessageState) String() (str string) {
	switch m {
	case MS_Undefine:
		return "MS_Undefine"
	case MS_Probe:
		return "MS_Probe"
	case MS_Ask:
		return "MS_Ask"
	case MS_Succeed:
		return "MS_Succeed"
	case MS_Failed:
		return "MS_Failed"
	}
	return ""
}
