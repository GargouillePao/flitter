package core

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

type MsgHandler interface {
	Handle(Dealer, []byte) (err error)
	handleDirectly(Dealer, interface{}) (err error)
}

type msgHandler struct {
	handler func(Dealer, interface{}) error
	getter  func() proto.Message
}

func (mh *msgHandler) Handle(d Dealer, body []byte) (err error) {
	msg := mh.getter()
	err = proto.Unmarshal(body, msg)
	if err != nil {
		return
	}
	err = mh.handler(d, msg)
	return
}

func (mh *msgHandler) handleDirectly(d Dealer, msg interface{}) (err error) {
	err = mh.handler(d, msg)
	return
}

type MsgProcesser interface {
	Rejister(headId uint32, getter func() proto.Message, act func(Dealer, interface{}) error)
	GetHandler(headId uint32) (h MsgHandler, err error)
	handleErr(d Dealer, err error)
	Process(d Dealer, buf []byte) (n int)
}

type msgProcesser struct {
	handlers map[uint32]MsgHandler
}

func NewMsgProcesser() MsgProcesser {
	mp := &msgProcesser{
		handlers: make(map[uint32]MsgHandler),
	}
	mp.Rejister(MID_INNER_ERROR, nil, func(d Dealer, err interface{}) error {
		log.Println(d, err)
		return nil
	})
	return mp
}

func (mp *msgProcesser) handleErr(d Dealer, err error) {
	mp.handlers[MID_INNER_ERROR].handleDirectly(d, err)
}

func (mp *msgProcesser) GetHandler(headId uint32) (h MsgHandler, err error) {
	h, ok := mp.handlers[headId]
	if !ok {
		err = errors.New(fmt.Sprintf("headid[%d] not rejistered", headId))
		return
	}
	return
}

func (mp *msgProcesser) Process(d Dealer, in []byte) (n int) {
	n, headId, body, err := DecodeMsg(in)
	if err != nil {
		return
	}
	handler, err := mp.GetHandler(headId)
	if err != nil {
		mp.handleErr(d, err)
		return
	}
	err = handler.Handle(d, body)
	if err != nil {
		mp.handleErr(d, err)
		return
	}
	return
}

func (mp *msgProcesser) Rejister(headId uint32, getter func() proto.Message, handler func(Dealer, interface{}) error) {
	mp.handlers[headId] = &msgHandler{
		handler: handler,
		getter:  getter,
	}
}

func DecodeMsg(inp []byte) (n int, head uint32, data []byte, err error) {
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

func EncodeMsg(head uint32, body proto.Message) (oup []byte, err error) {
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
