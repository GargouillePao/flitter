package servers

import (
	"errors"
	"flag"
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
)

const __LooperSize int = 10

type Server interface {
	Config(ca ConfigAction, st ServerType, addr string) error
	Init() error
	Start()
	Term()
	String() string
}

var __lauched bool = false

func Lauch() {
	if !__lauched {
		__lauched = true
		verb := flag.Bool("v", false, "verbs")
		filename := flag.String("log", "", "log in your path")
		flag.Parse()
		utils.InitLog(*verb, *filename)
	}
}

func transAddress(oldAddr string, st_from ServerType, st_to ServerType) (addr string, err error) {
	var host string
	var port int
	info := core.NodeInfo(oldAddr)
	host, err = info.GetHost()
	if err != nil {
		return
	}
	port, err = info.GetPort()
	if err != nil {
		return
	}
	switch st_from {
	case ST_Name:
		switch st_to {
		case ST_Watch:
			addr = fmt.Sprintf("%s:%d", host, port)
		default:
			err = errors.New(fmt.Sprintf("Invalid Translate for %v to %v", st_from, st_to))
		}
	case ST_Watch:
		switch st_to {
		case ST_Name:
			addr = fmt.Sprintf("%s:%d", host, port)
		case ST_HeartBeat:
			addr = fmt.Sprintf("%s:%d", host, port+1)
		default:
			err = errors.New(fmt.Sprintf("Invalid Translate for %v to %v", st_from, st_to))
		}
	case ST_HeartBeat:
		switch st_to {
		case ST_Watch:
			addr = fmt.Sprintf("%s:%d", host, port+2)
		case ST_HeartBeat:
			addr = fmt.Sprintf("%s:%d", host, port+3)
		default:
			err = errors.New(fmt.Sprintf("Invalid Translate for %v to %v", st_from, st_to))
		}
	default:
		err = errors.New(fmt.Sprintf("Invalid Translate for %v to %v", st_from, st_to))
	}
	return
}

type ConfigAction uint8

const (
	_ ConfigAction = iota
	CA_Send
	CA_Recv
)

type ServerType uint8

const (
	_ ServerType = iota
	ST_Name
	ST_Watch
	ST_Quorum
	ST_HeartBeat
	ST_Service
)

func (s ServerType) String() string {
	switch s {
	case ST_Name:
		return "ST_Name"
	case ST_Watch:
		return "ST_Watch"
	case ST_Quorum:
		return "ST_Quorum"
	case ST_HeartBeat:
		return "ST_HeartBeat"
	case ST_Service:
		return "ST_Service"
	}
	return ""
}

type ClusterType uint8

const (
	_ ClusterType = iota
	CT_Referee
	CT_Worker
)
