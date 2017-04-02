package main

import (
	"fmt"
	flitter "github.com/GargouillePao/flitter/core"
	utils "github.com/GargouillePao/flitter/utils"
	"time"
)

func main() {
	waiting := make(chan bool, 4)
	msgChan := make(chan int, 10)
	ctrl := make(chan bool, 0)
	worker := newWorker()
	if worker == nil {
		return
	}
	leader := newLeader()
	if leader == nil {
		return
	}
	go func() {
		msg := flitter.NewMessage(flitter.NewMessageInfo(), []byte("Probe"))
		err := worker.SendToLeader(msg)
		if err != nil {
			utils.ErrIn(err, "Client")
			return
		}
		fmt.Println(utils.Infof("send to leader:%v", msg))
		waiting <- true
	}()
	go func() {
		msg, err := leader.ReceiveFromChilren()
		if err != nil {
			utils.ErrIn(err, "Server")
			return
		}
		fmt.Println(utils.Infof("recv from chilren send:%v", msg))
		msgChan <- 1
		waiting <- true
	}()
	go func() {
		msg, err := worker.ReceiveFromLeader()
		if err != nil {
			utils.ErrIn(err, "Client")
			return
		}
		fmt.Println(utils.Infof("recv from leader broadcast:%v", msg))
		waiting <- true
	}()
	go func() {
		times := <-msgChan
		for {
			time.Sleep(time.Second)
			msg := flitter.NewMessage(flitter.NewMessageInfo(), []byte("Hello"))
			err := leader.BroadcastToChildren(msg)
			if err != nil {
				utils.ErrIn(err, "Server")
				return
			}
			fmt.Println(utils.Infof("broadcast to children:%v", msg))
			times++
			if times >= 1 {
				break
			}
		}
		waiting <- true
	}()
	go func() {
		for {
			select {
			case down := <-ctrl:
				if down {
					fmt.Println("End")
				} else {
					fmt.Println("Start")
				}
			}
		}
	}()
	ctrl <- false
	<-waiting
	<-waiting
	<-waiting
	<-waiting
	ctrl <- true
}
func newWorker() flitter.Node {
	info := flitter.NewNodeInfo()
	info.SetAddr("127.0.0.1", "8001")
	node, err := flitter.NewNode(info)
	if err != nil {
		utils.ErrIn(err, "Client")
		return nil
	}
	leaderInfo := flitter.NewNodeInfo()
	leaderInfo.SetAddr("127.0.0.1", "8080")
	err = node.SetLeader(leaderInfo)
	if err != nil {
		utils.ErrIn(err, "Client")
		return nil
	}
	fmt.Println(utils.Norf("Client now is :%v", node))
	return node
}
func newLeader() flitter.Node {
	info := flitter.NewNodeInfo()
	info.SetAddr("127.0.0.1", "8080")
	node, err := flitter.NewNode(info)
	if err != nil {
		utils.ErrIn(err, "Server")
		return nil
	}
	fmt.Println(utils.Norf("Server now is :%v", node))
	return node
}
