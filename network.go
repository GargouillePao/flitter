package flitter

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

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

func newDealer(conn net.Conn) *dealer {
	return &dealer{
		conn: conn,
		rbs:  &bufStream{},
		wbs:  &bufStream{},
	}
}

type dealer struct {
	s    *server
	id   uint64
	conn net.Conn
	rbs  *bufStream
	wbs  *bufStream
	addr string
	mp   MsgProcesser
}

func (d *dealer) String() string {
	return fmt.Sprintf("%v", d.conn.RemoteAddr())
}

func (d *dealer) Process() error {
	var buf [1024]byte
	for {
		n, err := d.conn.Read(buf[:])
		if err != nil {
			return err
		}
		data := d.rbs.Begin(buf[:n])
		n = d.mp.Process(d, data)
		d.rbs.End(data[n:])
	}
}

func (d *dealer) Send(head uint32, body proto.Message, pack bool) (err error) {
	data, err := EncodeMsg(head, body, pack)
	if err != nil {
		return
	}
	buf := d.wbs.Begin(data)
	n := 0
	if d.conn != nil {
		n, _ = d.conn.Write(buf)
	}
	d.wbs.End(buf[n:])
	return
}

func (d *dealer) Connect() (err error) {
	if d.conn != nil {
		return
	}
	d.conn, err = net.Dial("tcp", d.addr)
	if err != nil {
		return
	}
	return
}

func (d *dealer) setServer(s *server) {
	d.s = s
	d.mp = s.mp
	atomic.AddInt32(&d.s.dealCnt, 1)
}

func (d *dealer) Close() (err error) {
	if d.conn == nil {
		return
	}
	d.conn.Close()
	d.conn = nil
	if d.s != nil {
		atomic.AddInt32(&d.s.dealCnt, -1)
	}
	return
}

type server struct {
	ln           net.Listener
	mp           *msgProcesser
	clientWait   time.Duration
	serverWait   time.Duration
	dealMax      int32
	dealCnt      int32
	ds           sync.Map
	onconnect    func(d *dealer)
	ondisconnect func(d *dealer)
}

func newServer(mp MsgProcesser, dealCnt int32, clientWait, serverWait time.Duration) server {
	return server{
		clientWait:   clientWait,
		serverWait:   serverWait,
		mp:           mp.(*msgProcesser),
		dealMax:      dealCnt,
		onconnect:    func(d *dealer) {},
		ondisconnect: func(d *dealer) {},
	}
}

func (s *server) AddPeer(name, addr string, mp MsgProcesser) error {
	d := newDealer(nil)
	d.addr = addr
	d.mp = mp
	if err := d.Connect(); err != nil {
		return err
	}
	s.ds.Store(name, d)
	go func() {
		err := d.Process()
		if err != nil {
			d.mp.handleErr(d, err)
		}
	}()
	return nil
}

func (s *server) SendPeer(peer string, head uint32, body proto.Message) (err error) {
	d, ok := s.ds.Load(peer)
	if !ok {
		err = errors.New(fmt.Sprintf("Peer %s Not Found", peer))
		return
	}
	err = d.(*dealer).Send(head, body, false)
	return
}

func (s *server) serve(address string) {
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
			s.onconnect(d)
			for {
				if s.dealCnt > s.dealMax {
					err = errors.New(fmt.Sprintf("More Than %d Players", s.dealCnt))
				} else {
					err = d.conn.SetReadDeadline(time.Now().Add(s.clientWait))
					if err == nil {
						err = d.Process()
					}
				}
				if err != nil {
					s.mp.handleErr(d, err)
					s.ondisconnect(d)
					d.Close()
					return
				}
			}
		}()
	}
}
