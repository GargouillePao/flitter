package main

import (
	servers "github.com/GargouillePao/flitter/servers"
	utils "github.com/GargouillePao/flitter/utils"
	"time"
)

const (
	_Referee_Addr_ string = "127.0.0.1:7000"
	_Worker_1_     string = "127.0.0.1:8100"
	_Worker_2_     string = "127.0.0.1:8200"
)

func main() {
	wg := make(chan bool, 1)
	servers.Lauch()
	go func() {
		name()
		wg <- true
	}()

	go func() {
		heart1()
	}()
	go func() {
		heart2()
	}()
	go func() {
		time.Sleep(time.Second * 2)
		watch1()
	}()
	go func() {
		time.Sleep(time.Second * 2)
		watch2()
	}()
	<-wg
}
func heart1() {
	utils.Logf(utils.Norf, "Start Heart1")
	var err error
	server := servers.NewHeartbeatServer()
	err = server.Config(servers.CA_Recv, servers.ST_Watch, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart1]recv watch %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_Watch, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart1]send watch %v", err)
		return
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart1]receive heartbeat %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart1]send heartbeat %v", err)
		return
	}
	err = server.Init()
	if err != nil {
		utils.Logf(utils.Errf, "[Heart1]init %v", err)
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	utils.Logf(utils.Norf, "End Heart1")
}
func heart2() {
	utils.Logf(utils.Norf, "Start Heart2")
	var err error
	server := servers.NewHeartbeatServer()
	err = server.Config(servers.CA_Recv, servers.ST_Watch, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart2]recv watch %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_Watch, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart2]send watch %v", err)
		return
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart2]receive heartbeat %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Heart2]send heartbeat %v", err)
		return
	}
	err = server.Init()
	if err != nil {
		utils.Logf(utils.Errf, "[Heart2]init %v", err)
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	utils.Logf(utils.Norf, "End Heart2")
}
func watch1() {
	utils.Logf(utils.Norf, "Start Watch1")
	var err error
	server := servers.NewWatchServer()
	err = server.Config(servers.CA_Recv, servers.ST_Name, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch1]recv name %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_Name, _Referee_Addr_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch1]send name %v", err)
		return
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch1]recv heartbeat %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, _Worker_1_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch1]send heartbeat %v", err)
		return
	}
	err = server.Init()
	if err != nil {
		utils.Logf(utils.Errf, "[Watch1]init %v", err)
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	utils.Logf(utils.Norf, "End Watch1")
}
func watch2() {
	utils.Logf(utils.Norf, "Start Watch2")
	var err error
	server := servers.NewWatchServer()
	err = server.Config(servers.CA_Recv, servers.ST_Name, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch2]recv name %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_Name, _Referee_Addr_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch2]send name %v", err)
		return
	}
	err = server.Config(servers.CA_Recv, servers.ST_HeartBeat, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch2]recv heartbeat %v", err)
		return
	}
	err = server.Config(servers.CA_Send, servers.ST_HeartBeat, _Worker_2_)
	if err != nil {
		utils.Logf(utils.Errf, "[Watch2]send heartbeat %v", err)
		return
	}
	err = server.Init()
	if err != nil {
		utils.Logf(utils.Errf, "[Watch2]init %v", err)
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	utils.Logf(utils.Norf, "End Watch2")
}
func name() {
	utils.Logf(utils.Norf, "Start Name")
	var err error
	server := servers.NewNameServer()
	err = server.Config(servers.CA_Recv, servers.ST_Watch, _Referee_Addr_)
	if err != nil {
		utils.Logf(utils.Errf, "%v", err)
		return
	}
	err = server.Init()
	if err != nil {
		utils.Logf(utils.Errf, "%v", err)
		return
	}
	time.AfterFunc(time.Second*10, func() {
		server.Term()
	})
	server.Start()
	utils.Logf(utils.Norf, "End Name")
}
