package core

/*hand over the messages*/
type MessageLooper interface {
	AddHandler(action MessageAction, handler MessageHandler)
	RemoveHandler(action MessageAction)
	Loop(wait bool)
	/** wait until the loop ends	 */
	Wait()
}

func NewMessageLooper() MessageLooper {
	return &messageLooper{}
}

/*hand over the messages*/
type messageLooper struct {
	msgs     chan Message
	handlers map[MessageAction]MessageHandler
	waiting  chan bool
}

func (m *messageLooper) AddHandler(action MessageAction, handler MessageHandler) {
}
func (m *messageLooper) RemoveHandler(action MessageAction) {
}
func (m *messageLooper) Loop(wait bool) {
}

/** wait until the loop ends	 */
func (m *messageLooper) Wait() {
}

type MessageHandler func(msg Message) error
