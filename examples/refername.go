package main

import (
	"fmt"
	core "github.com/GargouillePao/flitter/core"
	servers "github.com/GargouillePao/flitter/servers"
	utils "github.com/GargouillePao/flitter/utils"
	//"os"
	"time"
)

const (
	_Name_Addr_ core.NodeInfo = "127.0.0.1:7000"

	_Watch_Addr_1_      core.NodeInfo = "127.0.0.1:8100"
	_Watch_HeartAddr_1_ core.NodeInfo = "127.0.0.1:8101"
	_Heart_Addr_1_      core.NodeInfo = "127.0.0.1:8181"

	_Watch_Addr_2_      core.NodeInfo = "127.0.0.1:8200"
	_Watch_HeartAddr_2_ core.NodeInfo = "127.0.0.1:8201"
	_Heart_Addr_2_      core.NodeInfo = "127.0.0.1:8281"
)

func main() {
	wg := make(chan bool, 3)
	go func() {
		watch1()
		wg <- true
	}()
	go func() {
		watch2()
		wg <- true
	}()
	go func() {
		name()
		wg <- true
	}()
	<-wg
	<-wg
	<-wg
}
func watch1() {
	fmt.Println(utils.Norf("Start Watch1"))
	var err error
	server := servers.NewWatchServer()
	err = server.Config(servers.CA_Recv, servers.ST_Name, _Watch_Addr_1_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch1]recv name %v", err))
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_Name, _Name_Addr_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch1]send name %v", err))
		return
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, _Watch_HeartAddr_1_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch1]recv heartbeat %v", err))
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, _Heart_Addr_1_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch1]send heartbeat %v", err))
		return
	}
	err = server.Init()
	if err != nil {
		fmt.Println(utils.Errf("[Watch1]init %v", err))
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	fmt.Println(utils.Norf("End Watch1"))
}
func watch2() {

	fmt.Println(utils.Norf("Start Watch2"))
	var err error
	server := servers.NewWatchServer()
	err = server.Config(servers.CA_Recv, servers.ST_Name, _Watch_Addr_2_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch2]recv name %v", err))
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_Name, _Name_Addr_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch2]send name %v", err))
		return
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, _Watch_HeartAddr_2_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch2]recv heartbeat %v", err))
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, _Heart_Addr_2_)
	if err != nil {
		fmt.Println(utils.Errf("[Watch2]send heartbeat %v", err))
		return
	}
	err = server.Init()
	if err != nil {
		fmt.Println(utils.Errf("[Watch2]init %v", err))
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	fmt.Println(utils.Norf("End Watch2"))
}
func name() {
	fmt.Println(utils.Norf("Start Name"))
	var err error
	server := servers.NewNameServer()
	err = server.Config(servers.CA_Recv, servers.ST_Watch, _Name_Addr_)
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
		return
	}
	err = server.Init()
	if err != nil {
		fmt.Println(utils.Errf("%v", err))
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	fmt.Println(utils.Norf("End Name"))
}
