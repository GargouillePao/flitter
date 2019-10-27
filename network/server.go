package network

import (
	"log"
	"net"
	"time"
)

type Server struct {
	l          net.Listener
	network    string
	addr       string
	handler    func(*Agent)
	retryTimes int
}

//NewServer NewServer
func NewServer(network, addr string) *Server {
	return &Server{
		network:    network,
		addr:       addr,
		retryTimes: 3,
	}
}

func (s *Server) Handle(handler func(*Agent)) *Server {
	s.handler = handler
	return s
}

func (s *Server) initAgent(a *Agent) {
	retry := s.retryTimes
	defer a.Close()
	a.conncb()
	for {
		err := a.handle()
		if err != nil || retry <= 0 {
			if nerr, ok := err.(net.Error); ok {
				if nerr.Temporary() || nerr.Timeout() {
					time.Sleep(time.Second * 3)
					retry = retry - 1
					continue
				}
			}
			return
		}
		retry = s.retryTimes
	}
}

func (s *Server) Start() (err error) {
	log.Println("Start")
	s.l, err = net.Listen(s.network, s.addr)
	if err != nil {
		return
	}
	log.Println(s.l)
	for {
		var conn net.Conn
		conn, err = s.l.Accept()
		if err != nil {
			return
		}
		a := NewAgent(conn)
		s.handler(a)
		go s.initAgent(a)
	}
}
