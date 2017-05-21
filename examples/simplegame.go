package main

import (
	"errors"
	"flag"
	utils "github.com/gargous/flitter/common"
	"github.com/gargous/flitter/core"
	"github.com/gargous/flitter/servers"
	socketio "github.com/googollee/go-socket.io"
)

const (
	_Referee_Path_ string = "referee@127.0.0.1:5000"
)

func main() {
	npath := flag.String("p", _Referee_Path_, "-p [your node path]")
	servers.Lauch()
	flag.Parse()
	groupName, ok := core.NodePath(*npath).GetGroupName()
	if !ok {
		return
	}

	switch groupName {
	case "referee":
		handleReferee(*npath)
	case "scene":
		handleScene(*npath)
	case "login":
		handleLogin(*npath)
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
func handleWorker(npath string) servers.Worker {
	var err error
	server, err := servers.NewWorker(core.NodePath(npath))
	if err != nil {
		utils.ErrQuit(err, "Worker")
		return nil
	}
	watcher := servers.NewWatchService()
	watcher.ConfigRefereeServer(core.NodePath(_Referee_Path_))
	server.ConfigService(servers.ST_Watch, watcher)
	server.ConfigService(servers.ST_HeartBeat, servers.NewHeartbeatService())

	return server
}
func handleScene(npath string) {
	utils.Logf(utils.Norf, "Start Scene")
	server := handleWorker(npath)
	if server == nil {
		return
	}
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
			cinfo := core.NewClientInfo(name, server.GetPath())
			dinfo := core.NewDataInfo("pos")
			dinfo.Value = posData
			err = scencer.UpdateClientData(cinfo, dinfo)
			if err != nil {
				utils.ErrIn(errors.New("Set Client Data With k=pos Failed"))
				return
			}
		}
	})
	scencer.OnClientUpdate(core.NewDataInfo("pos"), func(cInfo core.ClientInfo, dInfo core.DataInfo) error {
		clientsdata := scencer.GetClientData(core.NewClientInfo("", ""), core.NewDataInfo("pos"))
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
	// moneydata := make(chan float32, 10)
	// server.OnClient("money_lock", func(so socketio.Socket) interface{} {
	// 	return func(name string, money float32) {
	// 		err := scencer.LockClientData(name, "money", -1)
	// 		if err != nil {
	// 			utils.ErrIn(errors.New("Lock Client Data With k=money all Failed"))
	// 			return
	// 		}
	// 		moneydata <- money
	// 	}
	// })
	// server.OnClient("money_unlock", func(so socketio.Socket) interface{} {
	// 	return func(name string, score float32) {
	// 		err := scencer.UnlockClientData(name, "money", -1)
	// 		if err != nil {
	// 			utils.ErrIn(errors.New("UnLock Client Data With k=score Failed"))
	// 			return
	// 		}
	// 	}
	// })
	// server.OnClient("money", func(so socketio.Socket) interface{} {
	// 	return func(name string, money float32) {
	// 		err := scencer.LockClientData(name, "money", 1)
	// 		if err != nil {
	// 			utils.ErrIn(errors.New("Lock Client Data With k=monney Failed"))
	// 			return
	// 		}
	// 		moneydata <- money
	// 	}
	// })

	// scencer.OnClientLock("money", func(cname string) error {
	// 	fmt.Println(cname)
	// 	money := <-moneydata
	// 	var data utils.DataItem
	// 	err := data.Parse(money)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	scencer.UpdateClientData(cname, "money", data, 1)
	// 	return nil
	// })
	// scencer.OnClientUpdate("money", func(cname string, cdvalue utils.DataItem) error {
	// 	fmt.Println("Update")
	// 	clientsdata := scencer.GetClientData("", "money")
	// 	clientsmoney := make(map[string]float32)
	// 	var err error
	// 	ok := false
	// 	for k, v := range clientsdata {
	// 		clientsmoney[k], err = utils.ByteArrayToFloat32(v.Data)
	// 		if err != nil {
	// 			utils.ErrIn(err)
	// 		} else {
	// 			ok = true
	// 		}
	// 	}
	// 	if ok {
	// 		server.GetClientSocket().BroadcastTo("flitter", "money", clientsmoney)
	// 	}
	// 	utils.Logf(utils.Norf, "%v", scencer)
	// 	return nil
	// })

	// scoreData := make(chan float32, 10)
	// server.OnClient("score_lock", func(so socketio.Socket) interface{} {
	// 	return func(name string, score float32) {
	// 		err := scencer.LockClientData(name, "score", 0)
	// 		if err != nil {
	// 			utils.ErrIn(errors.New("Lock Client Data With k=score Failed"))
	// 			return
	// 		}
	// 		scoreData <- score
	// 	}
	// })
	// server.OnClient("score_unlock", func(so socketio.Socket) interface{} {
	// 	return func(name string, score float32) {
	// 		err := scencer.UnlockClientData(name, "score", 0)
	// 		if err != nil {
	// 			utils.ErrIn(errors.New("UnLock Client Data With k=score Failed"))
	// 			return
	// 		}
	// 	}
	// })
	// scencer.OnClientLock("score", func(cname string) error {
	// 	data := <-scoreData

	// 	scoresData, ok := server.Get("score")

	// 	var scores []float32
	// 	if ok {
	// 		scores, err := utils.ByteArrayToFloat32Array(scoresData.Data)
	// 		if err != nil {
	// 			return err
	// 		} else {
	// 			scores = append(scores, data)
	// 		}
	// 	} else {
	// 		scores = []float32{data}
	// 	}
	// 	fmt.Println(scores)
	// 	var scoresDataItem utils.DataItem
	// 	err := scoresDataItem.Parse(scores)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	scencer.UpdateClientData(cname, "score", scoresDataItem, 0)
	// 	return nil
	// })
	// scencer.OnClientUpdate("score", func(cname string, cdvalue utils.DataItem) error {
	// 	scoresArray, err := utils.ByteArrayToFloat32Array(cdvalue.Data)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	server.GetClientSocket().BroadcastTo("flitter", "score", scoresArray)
	// 	utils.Logf(utils.Norf, "score\n%v", scoresArray)
	// 	return nil
	// })

	err := server.Start()
	utils.ErrQuit(err, "Worker")
	utils.Logf(utils.Norf, "End Scene")
}

func handleLogin(npath string) {
	utils.Logf(utils.Norf, "Start Login")
	server := handleWorker(npath)
	if server == nil {
		return
	}
	scencer := servers.NewScenceService()
	server.ConfigService(servers.ST_Scence, scencer)
	server.OnClient("login", func(so socketio.Socket) interface{} {
		return func(name string, account string, password string) (_account string, ok bool) {
			cinfo := core.NewClientInfo(name, server.GetPath())
			accountDatas := scencer.GetClientData(cinfo, core.NewDataInfo("account"))
			accountData, ok := accountDatas[cinfo.GetName()]
			if !ok || len(accountData.Data) <= 0 {
				ok = false
				return
			}
			pswDatas := scencer.GetClientData(cinfo, core.NewDataInfo("password"))
			pswData, ok := pswDatas[cinfo.GetName()]
			if !ok || len(pswData.Data) <= 0 {
				ok = false
				return
			}

			_account = string(accountData.Data)
			psw := string(pswData.Data)
			if _account == account && psw == password {
				ok = true
				return
			}
			ok = false
			return
		}
	})
	sessions := make(map[string]socketio.Socket)
	server.OnClient("signup", func(so socketio.Socket) interface{} {
		return func(name string, password string) (cname string, account string, signed bool, ok bool) {
			cinfo := core.NewClientInfo(name, server.GetPath())
			cname = cinfo.GetName()
			accountDatas := scencer.GetClientData(cinfo, core.NewDataInfo("account"))
			accountData, ok := accountDatas[cname]
			if ok && len(accountData.Data) > 0 {
				account = string(accountData.Data)
				signed = true
				ok = false
				return
			}
			account = name
			pswInfo := core.NewDataInfo("password")
			pswInfo.Value = utils.DataItem{Data: []byte(password)}
			err := scencer.UpdateClientData(cinfo, pswInfo)
			if err != nil {
				utils.ErrIn(err)
				ok = false
				return
			}
			accountInfo := core.NewDataInfo("account")
			accountInfo.Value = utils.DataItem{Data: []byte(name)}
			err = scencer.UpdateClientData(cinfo, accountInfo)
			if err != nil {
				utils.ErrIn(err)
				ok = false
				return
			}
			ok = true
			signed = false
			sessions[cname] = so
			return
		}
	})
	scencer.OnClientUpdate(core.NewDataInfo("account"), func(cinfo core.ClientInfo, dinfo core.DataInfo) (err error) {
		account := string(dinfo.Value.Data)
		so, ok := sessions[cinfo.GetName()]
		if ok {
			so.Emit("signup", account)
		}
		return
	})
	err := server.Start()
	utils.ErrQuit(err, " At Worker Start")
	utils.Logf(utils.Norf, "End Login")
}
