package core

import (
	utils "github.com/gargous/flitter/utils"
	"time"
)

type TimeTricker struct {
	timer    *time.Timer
	duration time.Duration
}

func (t *TimeTricker) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
	t.timer = nil
}

/*hand over the messages*/
type MessageLooper interface {
	GetHandler() map[MessageAction]MessageHandler
	AddHandler(maxHandleTime time.Duration, action MessageAction, handler MessageHandler)
	RemoveHandler(action MessageAction)
	SetInterval(timestamp time.Duration, handler func(t time.Time) error)
	Loop()
	Push(msg Message)
	Term()
}

func NewMessageLooper(bufferSize int) MessageLooper {
	return &messageLooper{
		msgs:          make(chan Message, bufferSize),
		handlers:      make(map[MessageAction]MessageHandler),
		handleTricker: make(map[MessageAction]TimeTricker),
	}
}

/*hand over the messages*/
type messageLooper struct {
	msgs          chan Message
	handlers      map[MessageAction]MessageHandler
	handleTricker map[MessageAction]TimeTricker
}

func (m *messageLooper) goHandle(handler MessageHandler, msg Message) {
	go func() {
		m.gatherError(handler, msg)
	}()
}
func (m *messageLooper) gatherError(handler MessageHandler, msg Message) {
	err := handler(msg)
	if err != nil {
		if msg == nil {
			msg = NewMessage(NewMessageInfo(), []byte(""))
		}
		msg.GetInfo().SetState(MS_Error)
		msg.SetContent([]byte(err.Error()))
		m.Push(msg)
	}
}
func (m *messageLooper) Push(msg Message) {
	if m.msgs != nil {
		_, state, _ := msg.GetInfo().Info()
		if state == MS_Failed {
			msg.Visit()
		}
		m.msgs <- msg
	}
}
func (m *messageLooper) AddHandler(maxHandleTime time.Duration, action MessageAction, handler MessageHandler) {
	m.handlers[action] = handler
	if maxHandleTime > 0 {
		m.handleTricker[action] = TimeTricker{
			timer:    nil,
			duration: maxHandleTime * time.Millisecond,
		}
	}
}
func (m *messageLooper) GetHandler() map[MessageAction]MessageHandler {
	return m.handlers
}
func (m *messageLooper) RemoveHandler(action MessageAction) {
	delete(m.handlers, action)
}
func (m *messageLooper) SetInterval(timestamp time.Duration, handler func(t time.Time) error) {
	timer := time.Tick(timestamp * time.Millisecond)
	go func() {
		for {
			select {
			case t := <-timer:
				m.gatherError(func(msg Message) error {
					return handler(t)
				}, nil)
			}
		}
	}()
}
func (m *messageLooper) Loop() {
	for {
		select {
		case msg, isOpen := <-m.msgs:
			if !isOpen {
				utils.Logf(utils.Errf, "Message Loop Closed")
				m.term()
				return
			} else {
				msg = msg.Copy()
				action, state, _ := msg.GetInfo().Info()
				if action == MA_Term || len(m.handlers) <= 0 {
					m.term()
					return
				}
				//for retry
				tricker, ok := m.handleTricker[action]
				if ok {
					func(tricker TimeTricker, state MessageState, action MessageAction) {
						switch state {
						case MS_Probe:
							if tricker.timer == nil {
								tricker.timer = time.AfterFunc(
									tricker.duration,
									func() {
										msg.GetInfo().SetState(MS_Failed)
										tricker.Stop()
										m.handleTricker[action] = tricker
										m.Push(msg)
									},
								)
							}
							m.handleTricker[action] = tricker
						case MS_Succeed:
							tricker.Stop()
							m.handleTricker[action] = tricker
						}
					}(tricker, state, action)
				}

				handler := m.handlers[action]
				if handler != nil {
					m.goHandle(handler, msg)
				}
			}

		}
	}
}

func (m *messageLooper) Term() {
	info := NewMessageInfo()
	info.SetAcion(MA_Term)
	m.Push(NewMessage(info, []byte("")))
}

func (m *messageLooper) term() {
	if m.msgs != nil {
		close(m.msgs)
		m.msgs = nil
	}
	for _, tricker := range m.handleTricker {
		if tricker.timer != nil {
			tricker.timer.Stop()
			tricker.timer = nil
		}
	}
	utils.CloseError()
}

type MessageHandler func(msg Message) error
