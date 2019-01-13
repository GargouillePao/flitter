package flitter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
)

const (
	MID_INNER_ERROR = 1
)

type msgHandler struct {
	handler func(*dealer, interface{}) error
	getter  func() proto.Message
}

func (mh *msgHandler) Handle(d *dealer, body []byte) (err error) {
	msg := mh.getter()
	err = proto.Unmarshal(body, msg)
	if err != nil {
		return
	}
	err = mh.handler(d, msg)
	return
}

func (mh *msgHandler) handleDirectly(d *dealer, msg interface{}) (err error) {
	err = mh.handler(d, msg)
	return
}

type MsgProcesser interface {
	Register(headId uint32, act func(*dealer, interface{}) error)
	handleErr(d *dealer, err error)
	Process(d *dealer, buf []byte) (n int)
}

type msgProcesser struct {
	handlers map[uint32]*msgHandler
	creaters map[uint32]func() proto.Message
	pack     bool
}

func NewMsgProcesser(c map[uint32]func() proto.Message, pack bool) MsgProcesser {
	mp := &msgProcesser{
		handlers: make(map[uint32]*msgHandler),
		creaters: c,
		pack:     pack,
	}
	mp.Register(MID_INNER_ERROR, func(d *dealer, err interface{}) error {
		log.Println(d, err)
		return nil
	})
	return mp
}

func (mp *msgProcesser) handleErr(d *dealer, err error) {
	mp.handlers[MID_INNER_ERROR].handleDirectly(d, err)
}

func (mp *msgProcesser) Process(d *dealer, in []byte) (n int) {
	n, headId, body, err := DecodeMsg(in, mp.pack)
	if err != nil {
		mp.handleErr(d, err)
		return
	}
	handler, ok := mp.handlers[headId]
	if !ok {
		mp.handleErr(d, errors.New(fmt.Sprintf("Handler %d Not Register", headId)))
		return
	}
	err = handler.Handle(d, body)
	if err != nil {
		mp.handleErr(d, err)
		return
	}
	return
}

func (mp *msgProcesser) Register(headId uint32, handler func(*dealer, interface{}) error) {
	creater, ok := mp.creaters[headId]
	if !ok {
		creater = nil
	}
	mp.handlers[headId] = &msgHandler{
		handler: handler,
		getter:  creater,
	}
}

func DecodeMsg(inp []byte, pack bool) (n int, head uint32, data []byte, err error) {
	inpLen := len(inp)
	offset := 0
	if inpLen < offset+4 {
		err = errors.New("Msg No Header")
		return
	}
	head = binary.BigEndian.Uint32(inp[offset : offset+4])
	offset += 4
	if inpLen < offset+4 {
		err = errors.New("Msg No Content")
		return
	}
	dataLen := binary.BigEndian.Uint32(inp[offset : offset+4])
	offset += 4
	n = int(dataLen) + offset
	if inpLen < n {
		n = 0
		err = errors.New("Msg Content Error")
		return
	}
	data = inp[offset:n]
	return
}

func EncodeMsg(head uint32, body proto.Message, pack bool) (oup []byte, err error) {
	oup = make([]byte, 8)
	offset := 0
	binary.BigEndian.PutUint32(oup[offset:offset+4], head)
	offset += 4
	data, err := proto.Marshal(body)
	if err != nil {
		return
	}
	binary.BigEndian.PutUint32(oup[offset:offset+4], uint32(len(data)))
	offset += 4
	oup = append(oup, data...)
	return
}
