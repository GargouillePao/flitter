package core

import (
	"bytes"
	"errors"
	"fmt"
	common "github.com/gargous/flitter/common"
	"strings"
	"time"
)

/*(NULL)*/
type MessageAction uint8

const (
	_ MessageAction = iota
	MA_Undefine
	MA_Init
	MA_Refer
	MA_Checkin
	MA_Checkout
	MA_Join
	MA_Invite
	MA_Heartbeat
	MA_Crash
	MA_Vote
	MA_Upgrade
	MA_Update
	MA_Term
	MA_Lock
	MA_Unlock
	MA_User_Request
)

func (m MessageAction) Normalize() MessageAction {
	if m > 16 {
		m = MA_Undefine
	}
	return m
}

func (m MessageAction) String() (str string) {
	switch m {
	case MA_Init:
		return "MA_Init"
	case MA_Undefine:
		return "MA_Undefine"
	case MA_Refer:
		return "MA_Refer"
	case MA_Checkin:
		return "MA_Checkin"
	case MA_Checkout:
		return "MA_Checkout"
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
	case MA_Update:
		return "MA_Update"
	case MA_Upgrade:
		return "MA_Upgrade"
	case MA_User_Request:
		return "MA_User_Request"
	case MA_Lock:
		return "MA_Lock"
	case MA_Unlock:
		return "MA_Unlock"
	case MA_Term:
		return "MA_Term"
	}
	return ""
}

/*(NULL)*/
type MessageState uint8

const (
	_ MessageState = iota
	MS_Probe
	MS_Ask
	MS_Succeed
	MS_Failed
	MS_Error
	MS_Local
)

func (m MessageState) Normalize() MessageState {
	if m > 6 {
		m = MS_Probe
	}
	return m
}

func (m MessageState) String() (str string) {
	switch m {
	case MS_Probe:
		return "MS_Probe"
	case MS_Ask:
		return "MS_Ask"
	case MS_Succeed:
		return "MS_Succeed"
	case MS_Failed:
		return "MS_Failed"
	case MS_Error:
		return "MS_Error"
	case MS_Local:
		return "MS_Local"
	}
	return ""
}

/*(NULL)*/
type Message interface {
	GetInfo() (info MessageInfo)
	GetVisitTimes() int
	Visit()
	ClearContent()
	AppendContent(buf []byte)
	GetContent(index int) (buf []byte, ok bool)
	SetContents(buf [][]byte)
	GetContents() (buf [][]byte)
	Copy() Message
	String() string
}

func NewMessage(info MessageInfo) Message {
	msg := &message{
		info:     info,
		contents: make([][]byte, 0),
		visit:    0}
	return msg
}

/*(NULL)*/
type message struct {
	info     MessageInfo
	visit    int
	contents [][]byte
}

func (m *message) Copy() Message {
	msg := &message{
		info:     m.info.Copy(),
		contents: m.GetContents(),
		visit:    m.visit,
	}
	return msg
}
func (m *message) Visit() {
	m.visit += 1
}
func (m *message) GetVisitTimes() int {
	return m.visit
}
func (m *message) GetInfo() (info MessageInfo) {
	info = m.info
	return
}
func (m *message) AppendContent(buf []byte) {
	if m.contents == nil {
		m.contents = make([][]byte, 0)
	}
	m.contents = append(m.contents, buf)
	return
}
func (m *message) ClearContent() {
	if m.contents == nil {
		m.contents = make([][]byte, 0)
	}
	m.contents = m.contents[:0]
	return
}
func (m *message) GetContent(index int) (buf []byte, ok bool) {
	if index >= len(m.contents) {
		ok = false
		return
	}
	ok = true
	buf = m.contents[index]
	return
}
func (m *message) SetContents(bufs [][]byte) {
	m.contents = bufs
	return
}
func (m *message) GetContents() (bufs [][]byte) {
	bufs = m.contents
	return
}
func (m *message) String() string {
	infostr := strings.Join(strings.Split(fmt.Sprintf("%v", m.info), "\n"), "\n\t")
	contents := bytes.Join(m.contents, []byte("\n\t"))
	str := fmt.Sprintf("Message[\n\tinfo:%s\n\tvisist:%d\n\tcontents:%s\n]", infostr, m.visit, contents[:])
	return str
}

type MessageInfo interface {
	Serializable
	String() string
	SetAcion(action MessageAction)
	SetState(state MessageState)
	SetTime(_time time.Time)
	Copy() MessageInfo
	Info() (action MessageAction, state MessageState, _time time.Time)
}

func NewMessageInfo() MessageInfo {
	return &messageInfo{
		action: MA_Undefine,
		state:  MS_Probe,
		size:   10,
	}
}

/*the data of message head*/
type messageInfo struct {
	action   MessageAction
	state    MessageState
	sendtime int64
	size     int
}

func (m *messageInfo) Copy() MessageInfo {
	info := NewMessageInfo()
	info.SetState(m.state)
	info.SetAcion(m.action)
	info.SetTime(time.Unix(0, m.sendtime))
	return info
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
	timebuf, err := common.Ecode(m.sendtime)
	if err != nil {
		return 0, err
	}
	if len(timebuf) < 8 {
		err = errors.New("Invalid message time to read")
		return 0, err
	}
	for i := 2; i < m.Size(); i++ {
		buf[i] = timebuf[i-2]
	}
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
	buf = buf[1:]
	sendtime, err := common.ByteArrayToUInt64(buf)
	m.sendtime = int64(sendtime)
	return m.Size(), err
}

func (m *messageInfo) SetAcion(action MessageAction) {
	m.action = action
	return
}

func (m *messageInfo) SetState(state MessageState) {
	m.state = state
	return
}

func (m *messageInfo) SetTime(_time time.Time) {
	m.sendtime = _time.UnixNano()
	return
}

func (m *messageInfo) Info() (action MessageAction, state MessageState, _time time.Time) {
	action = m.action
	state = m.state
	_time = time.Unix(0, m.sendtime)
	return
}

func (m messageInfo) String() string {
	timstr := time.Unix(0, m.sendtime).Format("2006.01.02 15:04:05")
	str := fmt.Sprintf("MsgInfo[\n\taction:%v\n\tstate:%v\n\ttime:%s\n]", m.action, m.state, timstr)
	return str
}
func (m messageInfo) Size() int {
	return m.size
}
