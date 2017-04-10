package servers

import (
	"flag"
	core "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
)

var __verb bool = false

const __LooperSize int = 10

type Server interface {
	Config(ca ConfigAction, st ServerType, info core.NodeInfo) error
	Init() error
	Start()
	Term()
}

var __lauched bool = false

func lauchServer() {
	if !__lauched {
		__lauched = true
		verb := flag.Bool("v", false, "verbs")
		filename := flag.String("log", "", "log in your path")
		flag.Parse()
		__verb = *verb
		utils.InitLog(__verb, *filename)
	}
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
