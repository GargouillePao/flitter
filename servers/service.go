package servers

import (
	"flag"
	common "github.com/gargous/flitter/common"
	core "github.com/gargous/flitter/core"
)

const __LooperSize int = 10

type Service interface {
	Init(srv interface{}) error
	Start()
	Term()
	Push(msg core.Message)
	String() string
}

type baseService struct {
	looper core.MessageLooper
}

func (s *baseService) Push(msg core.Message) {
	action, _, _ := msg.GetInfo().Info()
	_, ok := s.looper.GetHandler()[action]
	if ok {
		s.looper.Push(msg)
	}
}

var __lauched bool = false

func Lauch() {
	if !__lauched {
		__lauched = true
		verb := flag.Bool("v", false, "verbs")
		filename := flag.String("log", "", "log in your path")
		flag.Parse()
		common.InitLog(*verb, *filename)
	}
}

type ConfigAction uint8

const (
	_ ConfigAction = iota
	CA_Send
	CA_Recv
)

type ServiceType uint8

const (
	_ ServiceType = iota
	ST_Name
	ST_Watch
	ST_Quorum
	ST_HeartBeat
	ST_Scence
)

func (s ServiceType) String() string {
	switch s {
	case ST_Name:
		return "ST_Name"
	case ST_Watch:
		return "ST_Watch"
	case ST_Quorum:
		return "ST_Quorum"
	case ST_HeartBeat:
		return "ST_HeartBeat"
	case ST_Scence:
		return "ST_Scence"
	}
	return ""
}
