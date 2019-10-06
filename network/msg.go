package network

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrMsgInvalid        error = errors.New("Msg Invalid")
	ErrMsgInvalidLen     error = errors.New("Msg Invalid Length")
	ErrMsgParseError     error = errors.New("Msg ErrMsgParseError")
	ErrMsgParseNoCreator error = errors.New("Msg ErrMsgParseNoCreator")
	ErrClientInvalidID   error = errors.New("Client Invalid Client ID")
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

type Processor struct {
	comp     composor
	encr     encryptor
	msgMaker map[uint32]func() interface{}
	r        *bufio.Reader
	rw       io.ReadWriter
}

//NewProcessor NewProcessor
func NewProcessor(rw io.ReadWriter) *Processor {
	return &Processor{
		r:  bufio.NewReader(rw),
		rw: rw,
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

func (p *Processor) Read() (head uint32, body []byte, err error) {
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

func (p *Processor) Write(head uint32, data []byte) (err error) {
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
