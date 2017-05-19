package servers

import (
	//"fmt"
	core "github.com/gargous/flitter/core"
	//common "github.com/gargous/flitter/common"
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
