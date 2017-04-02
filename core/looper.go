package core

import (
	"fmt"
	utils "github.com/GargouillePao/flitter/utils"
	"time"
)

type TimeTricker struct {
	timer    *time.Timer
	duration time.Duration
}

/*hand over the messages*/
type MessageLooper interface {
	AddHandler(maxHandleTime time.Duration, action MessageAction, handler MessageHandler)
	RemoveHandler(action MessageAction)
	SetInterval(timestamp time.Duration, handler func(t time.Time) error)
	Loop(wait bool)
	Push(msg Message)
	/** wait until the loop ends	 */
	Wait()
}

func NewMessageLooper(bufferSize int) MessageLooper {
	return &messageLooper{
		msgs:          make(chan Message, bufferSize),
		handlers:      make(map[MessageAction]MessageHandler),
		handleTricker: make(map[MessageAction]TimeTricker),
		waiting:       make(chan bool, 0),
	}
}

/*hand over the messages*/
type messageLooper struct {
	msgs          chan Message
	handlers      map[MessageAction]MessageHandler
	handleTricker map[MessageAction]TimeTricker
	waiting       chan bool
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
	m.msgs <- msg
}
func (m *messageLooper) AddHandler(maxHandleTime time.Duration, action MessageAction, handler MessageHandler) {
	m.handlers[action] = handler
	if maxHandleTime > 0 {
		m.handleTricker[action] = TimeTricker{
			timer:    nil,
			duration: maxHandleTime,
		}
	}
}
func (m *messageLooper) RemoveHandler(action MessageAction) {
	delete(m.handlers, action)
}
func (m *messageLooper) SetInterval(timestamp time.Duration, handler func(t time.Time) error) {
	timer := time.Tick(timestamp)
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
func (m *messageLooper) Loop(wait bool) {
	go func() {
	loop_handle:
		for {
			select {
			case msg := <-m.msgs:
				action, state, _ := msg.GetInfo().Info()
				if action == MA_Terminal || len(m.handlers) <= 0 {
					break loop_handle
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
										fmt.Println(utils.Warningf("handler time out and retrying"))
										msg.GetInfo().SetState(MS_Failed)
										tricker.timer.Stop()
										tricker.timer = nil
										m.handleTricker[action] = tricker
										m.Push(msg)
									},
								)
							}
							m.handleTricker[action] = tricker
						case MS_Succeed:
							tricker.timer.Stop()
							tricker.timer = nil
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
		m.waiting <- wait
	}()
	if wait {
		m.Wait()
	}
}

/** wait until the loop ends	 */
func (m *messageLooper) Wait() {
	<-m.waiting
}

type MessageHandler func(msg Message) error
