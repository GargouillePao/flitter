package network

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"runtime/debug"

	"github.com/dchest/siphash"
	"github.com/gogo/protobuf/proto"
)

var (
	ErrMsgInvalid          error = errors.New("Msg Invalid")
	ErrMsgInvalidLen       error = errors.New("Msg Invalid Length")
	ErrMsgParseError       error = errors.New("Msg Parse Error")
	ErrMsgParseNoCreator   error = errors.New("Msg Parse No Creator")
	ErrClientInvalidID     error = errors.New("Client Invalid Client ID")
	ErrMsgRegisterReqEmpty error = errors.New("Msg Register Req Empty")
	ErrMsgRegisterRepited  error = errors.New("Msg Register Repited")
	ErrMsgNotRegistered    error = errors.New("Msg Not Registered")
)

const MsgLenMax int = 4

type comperWriter interface {
	Flush() error
	Reset(w io.Writer)
	io.WriteCloser
}

type comperReader interface {
	Reset(r io.Reader)
	io.ReadCloser
}

type ComposeType uint8

const (
	CTgzip ComposeType = iota
)

type composor struct {
	ctype  ComposeType
	writer comperWriter
	reader comperReader
}

type encryptor struct {
	blockSize int
	writer    cipher.BlockMode
	reader    cipher.BlockMode
}

type MsgHandler struct {
	reqType reflect.Type
	ackType reflect.Type
	hasAck  bool
	h       func(req proto.Message, ack proto.Message)
}

type Processor struct {
	comp     composor
	encr     encryptor
	r        *bufio.Reader
	rw       io.ReadWriter
	HashKeys []uint64
	n2Id     map[string]uint32
	handlers map[uint32]MsgHandler
}

//NewProcessor NewProcessor
func NewProcessor(rw io.ReadWriter) *Processor {
	return &Processor{
		r:        bufio.NewReader(rw),
		rw:       rw,
		HashKeys: []uint64{0x1234567, 0x89ABCDE},
		n2Id:     make(map[string]uint32),
		handlers: make(map[uint32]MsgHandler),
	}
}
func (p *Processor) compose(input []byte) (output []byte, err error) {
	if p.comp.writer != nil {
		buf := bytes.Buffer{}
		p.comp.writer.Reset(&buf)
		_, err = p.comp.writer.Write(input)
		if err != nil {
			return
		}
		err = p.comp.writer.Flush()
		if err != nil {
			return
		}
		output = buf.Bytes()
	} else {
		output = input
	}
	return
}
func (p *Processor) decompose(input []byte) (output []byte, err error) {
	if p.comp.reader != nil {
		buf := bytes.Buffer{}
		p.comp.reader.Reset(&buf)
		_, err = p.comp.reader.Read(input)
		if err != nil {
			return
		}
		output = buf.Bytes()
	} else {
		output = input
	}
	return
}
func (e encryptor) padding(input []byte) []byte {
	padding := e.blockSize - len(input)%e.blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(input, padtext...)
}

func (e encryptor) unpadding(input []byte) []byte {
	length := len(input)
	unpadding := int(input[length-1])
	return input[:(length - unpadding)]
}
func (p *Processor) encrypt(input []byte) (output []byte) {
	if p.encr.writer != nil {
		input = p.encr.padding(input)
		output = make([]byte, len(input))
		p.encr.writer.CryptBlocks(output, input)
	} else {
		output = input
	}
	return
}
func (p *Processor) decrypt(input []byte) (output []byte) {
	if p.encr.reader != nil {
		output = make([]byte, len(input))
		p.encr.writer.CryptBlocks(output, input)
		output = p.encr.unpadding(output)
	} else {
		output = input
	}
	return
}

func (p *Processor) SetCipher(key []byte) error {
	p.encr.blockSize = len(key)
	wblock, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	rblock, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	p.encr.writer = cipher.NewCBCEncrypter(wblock, key[:p.encr.blockSize])
	p.encr.reader = cipher.NewCBCEncrypter(rblock, key[:p.encr.blockSize])
	return nil
}

func (p *Processor) regId(name string) uint32 {
	sum64 := siphash.Hash(p.HashKeys[0], p.HashKeys[1], []byte(name))
	sum32 := uint32(sum64 << 32 >> 32)
	p.n2Id[name] = sum32
	return sum32
}

func (p *Processor) Register(msg proto.Message) bool {
	h := MsgHandler{
		reqType: reflect.TypeOf(msg),
	}
	name := h.reqType.Elem().Name()
	sum32 := p.regId(name)
	if _, ok := p.handlers[sum32]; ok {
		return false
	}
	p.handlers[sum32] = h
	return true
}

func (p *Processor) GetId(msg proto.Message) uint32 {
	name := reflect.TypeOf(msg).Elem().Name()
	id, ok := p.n2Id[name]
	if !ok {
		return 0
	}
	return id
}

func (p *Processor) GetIdByName(msg string) uint32 {
	id, ok := p.n2Id[msg]
	if !ok {
		return 0
	}
	return id
}

func (p *Processor) OnReq(req, ack proto.Message, cb func(req, ack proto.Message)) bool {
	id := p.GetId(req)
	h, ok := p.handlers[id]
	if !ok {
		return false
	}
	h.ackType = reflect.TypeOf(ack)
	h.hasAck = true
	h.h = cb
	p.handlers[id] = h
	return true
}

func (p *Processor) onNotify(id uint32, cb func(msg proto.Message)) bool {
	h, ok := p.handlers[id]
	if !ok {
		return false
	}
	h.h = func(req, _ proto.Message) {
		cb(req)
	}
	p.handlers[id] = h
	return true
}

func (p *Processor) OnNotify(msg proto.Message, cb func(msg proto.Message)) bool {
	id := p.GetId(msg)
	return p.onNotify(id, cb)
}

func (p *Processor) handle() (err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = errors.New(string(debug.Stack()))
			return
		}
	}()
	h, b, err := p.read()
	if err != nil {
		return
	}
	ha, ok := p.handlers[h]
	if !ok {
		return ErrMsgNotRegistered
	}
	req := reflect.New(ha.reqType).Interface().(proto.Message)
	err = proto.Unmarshal(b, req)
	if err != nil {
		return
	}
	var ack proto.Message
	if ha.hasAck {
		ack = reflect.New(ha.ackType).Interface().(proto.Message)
	}
	ha.h(req, ack)
	return
}

func (p *Processor) send(req proto.Message) (err error) {
	name := reflect.TypeOf(req).Elem().Name()
	id, ok := p.n2Id[name]
	if !ok {
		return ErrMsgNotRegistered
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return
	}
	return p.write(id, data)
}

func (p *Processor) read() (head uint32, body []byte, err error) {
	buffLen := MsgLenMax
	data, err := p.r.Peek(buffLen)
	if err != nil {
		return
	}
	buffLen = buffLen + int(binary.BigEndian.Uint32(data))
	if buffLen <= MsgLenMax {
		err = ErrMsgInvalidLen
	}
	data, err = p.r.Peek(buffLen)
	if err != nil {
		return
	}
	_, err = p.r.Discard(buffLen)
	if err != nil {
		return
	}
	data = data[4:]
	data, err = p.decompose(data)
	if err != nil {
		return
	}
	data = p.decrypt(data)
	head = binary.BigEndian.Uint32(data[:4])
	body = data[4:]
	return
}

func (p *Processor) write(head uint32, data []byte) (err error) {
	var buff [8]byte
	binary.BigEndian.PutUint32(buff[4:], head)
	data = append(buff[4:], data...)
	data = p.encrypt(data)
	data, err = p.compose(data)
	if err != nil {
		return
	}
	binary.BigEndian.PutUint32(buff[:4], uint32(len(data)))
	ret := append(buff[:4], data...)
	_, err = p.rw.Write(ret)
	return
}
