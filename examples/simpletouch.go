package main

import (
	"flag"
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/servers"
	"github.com/gargous/flitter/utils"
)

const (
	_Referee_Path_ string = "referee@127.0.0.1:5000"
)

func main() {
	nreferee := flag.Bool("r", false, "-r means this is a referee")
	npath := flag.String("p", _Referee_Path_, "-n [your node path]")
	servers.Lauch()
	flag.Parse()
	if *nreferee {
		handleReferee(*npath)
	} else {
		handleWorker(*npath)
	}
}
func handleReferee(npath string) {
	utils.Logf(utils.Norf, "Start Referee")
	var err error
	server, err := servers.NewReferee(core.NodePath(npath))
	utils.ErrQuit(err, " Referee New")
	server.ConfigService(servers.ST_Name, servers.NewNameService())
	err = server.Start()
	utils.ErrQuit(err, " Referee Start")
	utils.Logf(utils.Norf, "End Referee")
}
func handleWorker(npath string) {
	utils.Logf(utils.Norf, "Start Worker")
	var err error
	server, err := servers.NewWorker(core.NodePath(npath))
	utils.ErrQuit(err, "Worker")
	watcher := servers.NewWatchService()
	watcher.ConfigRefereeServer(core.NodePath(_Referee_Path_))
	server.ConfigService(servers.ST_Watch, watcher)
	server.ConfigService(servers.ST_HeartBeat, servers.NewHeartbeatService())
	server.ConfigService(servers.ST_Scence, servers.NewScenceService())
	err = server.Start()
	utils.ErrQuit(err, "Worker")
	utils.Logf(utils.Norf, "End Worker")
}
