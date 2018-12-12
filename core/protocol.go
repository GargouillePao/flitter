package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/dchest/siphash"
	"github.com/golang/protobuf/proto"
)

func hash(str string) (output uint32) {
	h := siphash.Hash(0x12340000, 0x00005678, []byte(str))
	hu := h >> 32
	hd := (h << 32) >> 32
	output = uint32(hd | hu)
	return
}

type MsgHandler interface {
	GetName() string
	Handle(body []byte) (err error)
}

type msgHandler struct {
	name    string
	handler func(interface{}) error
	getter  func() proto.Message
}

func (mh *msgHandler) GetName() string {
	return mh.name
}

func (mh *msgHandler) Handle(body []byte) (err error) {
	msg := mh.getter()
	err = proto.Unmarshal(body, msg)
	if err != nil {
		return
	}
	err = mh.handler(msg)
	return
}

type MsgProcesser interface {
	Rejister(head string, getter func() proto.Message, act func(interface{}) error) error
	GetHandler(headId uint32) (h MsgHandler, err error)
	Decode(inp []byte) (head uint32, data []byte, err error)
	Encode(head uint32, body proto.Message) (oup []byte, err error)
	GetHeadId(head string) (headId uint32, err error)
	Process(buf []byte) (err error)
}

type msgProcesser struct {
	handlers   map[uint32]MsgHandler
	handlerIds map[string]uint32
}

func NewMsgProcesser() MsgProcesser {
	return &msgProcesser{
		handlers:   make(map[uint32]MsgHandler),
		handlerIds: make(map[string]uint32),
	}
}

func (mp *msgProcesser) GetHeadId(head string) (headId uint32, err error) {
	headId, ok := mp.handlerIds[head]
	if !ok {
		err = errors.New(fmt.Sprintf("headid[%d] not rejistered", headId))
		return
	}
	return
}

func (mp *msgProcesser) GetHandler(headId uint32) (h MsgHandler, err error) {
	h, ok := mp.handlers[headId]
	if !ok {
		err = errors.New(fmt.Sprintf("headid[%d] not rejistered", headId))
		return
	}
	return
}

func (mp *msgProcesser) Process(in []byte) (err error) {
	headId, body, err := mp.Decode(in)
	if err != nil {
		return
	}
	handler, err := mp.GetHandler(headId)
	if err != nil {
		return
	}
	err = handler.Handle(body)
	if err != nil {
		return
	}
	return
}

func (mp *msgProcesser) Rejister(head string, getter func() proto.Message, act func(interface{}) error) error {
	headId := hash(head)
	handler, ok := mp.handlers[headId]
	if ok {
		return errors.New(fmt.Sprintf("head[%s] has the same hash with head[%s]", head, handler.GetName()))
	}
	mp.handlers[headId] = &msgHandler{
		name:    head,
		handler: act,
		getter:  getter,
	}
	mp.handlerIds[head] = headId
	return nil
}

func (mp *msgProcesser) Decode(inp []byte) (head uint32, data []byte, err error) {
	inpLen := len(inp)
	offset := 0
	if inpLen < offset+4 {
		err = errors.New("Invalid Header Length")
		return
	}
	head = binary.BigEndian.Uint32(inp[offset : offset+4])
	offset += 4
	if inpLen < offset+4 {
		err = errors.New("Invalid Body Length Length")
		return
	}
	dataLen := binary.BigEndian.Uint32(inp[offset : offset+4])
	offset += 4
	if inpLen < int(dataLen)+offset {
		err = errors.New("Invalid Body")
		return
	}

	data = inp[offset : int(dataLen)+offset]
	return
}

func (mp *msgProcesser) Encode(head uint32, body proto.Message) (oup []byte, err error) {
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
