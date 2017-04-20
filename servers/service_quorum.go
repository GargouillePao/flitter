package servers

import (
	//"fmt"
	core "github.com/gargous/flitter/core"
	//utils "github.com/gargous/flitter/utils"
)

type QuorumService interface {
	Service
}

type quorumsrv struct {
	loop core.MessageLooper
}

func NewQuorumService() QuorumService {
	// server := &quorumsrv{
	// 	loop: core.NewMessageLooper(__LooperSize),
	// }
	return nil
}
