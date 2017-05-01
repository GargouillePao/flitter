package main

import (
	"errors"
	"flag"
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/servers"
	"github.com/gargous/flitter/utils"
	socketio "github.com/googollee/go-socket.io"
	"time"
)

const (
	_Referee_Path_ string = "referee@127.0.0.1:5000"
)

func main() {
	nreferee := flag.Bool("r", false, "-r means this is a referee")
	npath := flag.String("p", _Referee_Path_, "-p [your node path]")
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
	scencer := servers.NewScenceService()
	server.ConfigService(servers.ST_Scence, scencer)
	server.TrickClient("position", func(so socketio.Socket) interface{} {
		return func(name string, x float32, y float32) {
			var posData servers.DataItem
			err := posData.Parse([]float32{x, y})
			if err != nil {
				utils.ErrIn(errors.New("Parse Client Data With k=Postion Failed"))
				return
			}
			err = scencer.SetClientData(name, "position", posData)
			if err != nil {
				utils.ErrIn(errors.New("Set Client Data With k=Postion Failed"))
				return
			}
		}
	})
	scencer.OnScenceDataUpdate("position", func(cname string, cdkey string, cdvalue servers.DataItem, hostpath core.NodePath) {
		clientsdata := scencer.GetClientData("", cdkey)
		clientspos := make(map[string][]float32)
		var err error
		ok := false
		for k, v := range clientsdata {
			clientspos[k], err = utils.ByteArrayToFloat32Array(v.Data)
			if err != nil {
				utils.ErrIn(err)
			} else {
				ok = true
			}
		}
		if ok {
			server.GetClientSocket().BroadcastTo("flitter", cdkey, clientspos)
		}
	})
	go func() {
		for {
			time.Sleep(time.Second * 5)
			if scencer.IsAccess() {
				utils.Logf(utils.Norf, "%v", scencer)
			}
		}
	}()
	err = server.Start()
	utils.ErrQuit(err, "Worker")
	utils.Logf(utils.Norf, "End Worker")
}
