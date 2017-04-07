package main

import (
	"fmt"
	servers "github.com/GargouillePao/flitter/servers"
	utils "github.com/GargouillePao/flitter/utils"
	//"os"
	"time"
)

func main() {
	fmt.Println(utils.Norf("Start Watch"))
	var err error
	server := servers.NewWatchServer()
	err = server.Config(servers.CA_Recv, servers.ST_Name, "*:8000")
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
	}
	err = server.Config(servers.CA_Send, servers.ST_Name, "127.0.0.1:7000")
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, "*:8001")
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, "127.0.0.1:8081")
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
	}
	err = server.Init()
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
	}
	time.AfterFunc(time.Second*3, func() {
		server.Term()
	})
	server.Start()
	fmt.Println(utils.Norf("End Watch"))
}
