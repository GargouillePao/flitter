package flitter

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
	"sync/atomic"
	"time"
)

type Dealer interface {
	Send(head uint32, body proto.Message) error
	Connect(address string) error
	Process(mp MsgProcesser) error
	Close() error
}

type bufStream struct {
	b []byte
}

func (bs *bufStream) Begin(packet []byte) (data []byte) {
	data = packet
	if len(bs.b) > 0 {
		bs.b = append(bs.b, data...)
		data = bs.b
	}
	return data
}

func (bs *bufStream) End(data []byte) {
	if len(data) > 0 {
		if len(data) != len(bs.b) {
			bs.b = append(bs.b[:0], data...)
		}
	} else if len(bs.b) > 0 {
		bs.b = bs.b[:0]
	}
}

func NewDealer() Dealer {
	return &dealer{
		bs: &bufStream{},
	}
}

func newDealer(conn net.Conn) *dealer {
	return &dealer{
		conn: conn,
		bs:   &bufStream{},
	}
}

type dealer struct {
	s      *server
	conn   net.Conn
	bs     *bufStream
	closed bool
}

func (d *dealer) String() string {
	return fmt.Sprintf("%v", d.conn.RemoteAddr())
}

func (d *dealer) Process(mp MsgProcesser) error {
	var buf [1024]byte
	for {
		n, err := d.conn.Read(buf[:])
		if err != nil {
			return err
		}
		data := d.bs.Begin(buf[:n])
		n = mp.Process(d, data)
		d.bs.End(data[n:])
	}
}

func (d *dealer) Send(head uint32, body proto.Message) (err error) {
	data, err := EncodeMsg(head, body)
	if err != nil {
		return
	}
	_, err = d.conn.Write(data)
	return
}

func (d *dealer) Connect(address string) (err error) {
	d.conn, err = net.Dial("tcp", address)
	return
}

func (d *dealer) setServer(s *server) {
	d.s = s
	atomic.AddInt32(&d.s.dealCnt, 1)
}

func (d *dealer) Close() (err error) {
	if d.closed {
		return
	}
	d.closed = true
	d.conn.Close()
	if d.s != nil {
		atomic.AddInt32(&d.s.dealCnt, -1)
	}
	return
}

type server struct {
	ln       net.Listener
	mp       *msgProcesser
	readWait time.Duration
	dealMax  int32
	dealCnt  int32
}

func NewServer(mp MsgProcesser, dealCnt int32, readWait time.Duration) *server {
	return &server{
		readWait: readWait,
		mp:       mp.(*msgProcesser),
		dealMax:  dealCnt,
	}
}

func (s *server) Serve(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		s.mp.handleErr(nil, err)
		return
	}
	s.ln = ln
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			s.mp.handleErr(nil, err)
			return
		}
		d := newDealer(conn)
		go func() {
			d.setServer(s)
			if s.dealCnt > s.dealMax {
				err = errors.New(fmt.Sprintf("More Than %d Players", s.dealCnt))
			} else {
				err = d.conn.SetReadDeadline(time.Now().Add(s.readWait))
				if err == nil {
					err = d.Process(s.mp)
				}
			}
			if err != nil {
				s.mp.handleErr(d, err)
				d.Close()
			}
		}()
	}
}
