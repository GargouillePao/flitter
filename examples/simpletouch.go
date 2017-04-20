package main

import (
	"flag"
	servers "github.com/gargous/flitter/servers"
	utils "github.com/gargous/flitter/utils"
)

const (
	_Referee_Addr_ string = "127.0.0.1:5000"
)

func main() {
	nreferee := flag.Bool("r", false, "-r means this is a referee")
	nname := flag.String("n", "#1", "-n [your node name]")
	addr := flag.String("a", _Referee_Addr_, "-a [your node address]")
	servers.Lauch()
	flag.Parse()
	if *nreferee {
		handleReferee(*nname, *addr)
	} else {
		handleWorker(*nname, *addr)
	}
}
func handleReferee(name string, addr string) {
	utils.Logf(utils.Norf, "Start Referee")
	var err error
	server, err := servers.NewReferee(name, addr)
	utils.ErrQuit(err, " Referee New")
	server.ConfigService(servers.ST_Name, servers.NewNameService())
	err = server.Start()
	utils.ErrQuit(err, " Referee Start")
	utils.Logf(utils.Norf, "End Referee")
}
func handleWorker(name string, addr string) {
	utils.Logf(utils.Norf, "Start Worker")
	var err error
	server, err := servers.NewWorker(name, addr)
	utils.ErrQuit(err, "Worker")
	watcher := servers.NewWatchService()
	watcher.ConfigRefereeServer(_Referee_Addr_)
	server.ConfigService(servers.ST_Watch, watcher)
	server.ConfigService(servers.ST_HeartBeat, servers.NewHeartbeatService())
	server.ConfigService(servers.ST_Scence, servers.NewScenceService())
	err = server.Start()
	utils.ErrQuit(err, "Worker")
	utils.Logf(utils.Norf, "End Worker")
}
