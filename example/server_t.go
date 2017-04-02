package main

import (
	//"fmt"
	flitter "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	//"time"
	"os"
)

func main() {
	nodeInfo := flitter.NewNodeInfo()
	nodeInfo.SetAddr("127.0.0.1", os.Args[1])
	nodeInfo.SetPath(os.Args[2])
	node, err := flitter.NewNode(nodeInfo)
	utils.ErrQuit(err, "NewNode")
	server := flitter.NewServer(node)
	server.Init()
	server.Listen()
}
