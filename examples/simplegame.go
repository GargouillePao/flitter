package main

import (
	"errors"
	"flag"
	"fmt"
	utils "github.com/gargous/flitter/common"
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/servers"
	socketio "github.com/googollee/go-socket.io"
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
	server.OnClient("pos", func(so socketio.Socket) interface{} {
		return func(name string, x float32, y float32) {
			var posData utils.DataItem
			err := posData.Parse([]float32{x, y})
			if err != nil {
				utils.ErrIn(errors.New("Parse Client Data With k=pos Failed"))
				return
			}
			err = scencer.UpdateClientData(name, "pos", posData, 1)
			if err != nil {
				utils.ErrIn(errors.New("Set Client Data With k=pos Failed"))
				return
			}
		}
	})
	scencer.OnClientUpdate("pos", func(cname string, cdvalue utils.DataItem) error {
		clientsdata := scencer.GetClientData("", "pos")
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
			server.GetClientSocket().BroadcastTo("flitter", "pos", clientspos)
		}
		utils.Logf(utils.Norf, "%v", scencer)
		return nil
	})
	moneydata := make(chan float32, 10)
	server.OnClient("money_lock", func(so socketio.Socket) interface{} {
		return func(name string, money float32) {
			err = scencer.LockClientData(name, "money", -1)
			if err != nil {
				utils.ErrIn(errors.New("Lock Client Data With k=money all Failed"))
				return
			}
			moneydata <- money
		}
	})
	server.OnClient("money_unlock", func(so socketio.Socket) interface{} {
		return func(name string, score float32) {
			err = scencer.UnlockClientData(name, "money", -1)
			if err != nil {
				utils.ErrIn(errors.New("UnLock Client Data With k=score Failed"))
				return
			}
		}
	})
	server.OnClient("money", func(so socketio.Socket) interface{} {
		return func(name string, money float32) {
			err = scencer.LockClientData(name, "money", 1)
			if err != nil {
				utils.ErrIn(errors.New("Lock Client Data With k=monney Failed"))
				return
			}
			moneydata <- money
		}
	})

	scencer.OnClientLock("money", func(cname string) error {
		fmt.Println(cname)
		money := <-moneydata
		var data utils.DataItem
		err := data.Parse(money)
		if err != nil {
			return err
		}
		scencer.UpdateClientData(cname, "money", data, 1)
		return nil
	})
	scencer.OnClientUpdate("money", func(cname string, cdvalue utils.DataItem) error {
		fmt.Println("Update")
		clientsdata := scencer.GetClientData("", "money")
		clientsmoney := make(map[string]float32)
		var err error
		ok := false
		for k, v := range clientsdata {
			clientsmoney[k], err = utils.ByteArrayToFloat32(v.Data)
			if err != nil {
				utils.ErrIn(err)
			} else {
				ok = true
			}
		}
		if ok {
			server.GetClientSocket().BroadcastTo("flitter", "money", clientsmoney)
		}
		utils.Logf(utils.Norf, "%v", scencer)
		return nil
	})

	scoreData := make(chan float32, 10)
	server.OnClient("score_lock", func(so socketio.Socket) interface{} {
		return func(name string, score float32) {
			err = scencer.LockClientData(name, "score", 0)
			if err != nil {
				utils.ErrIn(errors.New("Lock Client Data With k=score Failed"))
				return
			}
			scoreData <- score
		}
	})
	server.OnClient("score_unlock", func(so socketio.Socket) interface{} {
		return func(name string, score float32) {
			err = scencer.UnlockClientData(name, "score", 0)
			if err != nil {
				utils.ErrIn(errors.New("UnLock Client Data With k=score Failed"))
				return
			}
		}
	})
	scencer.OnClientLock("score", func(cname string) error {
		data := <-scoreData
		scoresData, ok := server.Get("score")

		var scores []float32
		if ok {
			scores, err = utils.ByteArrayToFloat32Array(scoresData.Data)
			if err != nil {
				return err
			} else {
				scores = append(scores, data)
			}
		} else {
			scores = []float32{data}
		}
		fmt.Println(scores)
		var scoresDataItem utils.DataItem
		err := scoresDataItem.Parse(scores)
		if err != nil {
			return err
		}
		scencer.UpdateClientData(cname, "score", scoresDataItem, 0)
		return nil
	})
	scencer.OnClientUpdate("score", func(cname string, cdvalue utils.DataItem) error {
		scoresArray, err := utils.ByteArrayToFloat32Array(cdvalue.Data)
		if err != nil {
			return err
		}
		server.GetClientSocket().BroadcastTo("flitter", "score", scoresArray)
		utils.Logf(utils.Norf, "score\n%v", scoresArray)
		return nil
	})

	err = server.Start()
	utils.ErrQuit(err, "Worker")
	utils.Logf(utils.Norf, "End Worker")
}
