package core

import (
	"errors"
	"fmt"
	utils "github.com/GargouillePao/flitter/utils"
	"os"
	"time"
)

/*(NULL)*/
type Server interface {
	Init()
	Listen()
}

func NewServer(node Node) Server {
	looper := NewMessageLooper(10)
	_server := &server{
		node:              node,
		looper:            looper,
		heartbeatDuration: time.Second * 10,
	}
	_server.timeoutDurantion = _server.heartbeatDuration * 3
	return _server
}

/*(NULL)*/
type server struct {
	heartbeatDuration time.Duration
	timeoutDurantion  time.Duration
	node              Node
	looper            MessageLooper
}

func (s *server) Init() {
	s.handleNetworkMessage()
	s.handleHeartbeat()
	s.handleJoinleader()
}
func (s *server) handleNetworkMessage() {
	go func() {
		for {
			msg, err := s.node.ReceiveFromChilren()
			if err != nil {
				utils.ErrIn(err, "Receive from children")
				continue
			}
			fmt.Println(msg)
			s.looper.Push(msg)
		}
	}()
	go func() {
		for {
			msg, err := s.node.ReceiveFromLeader()
			if err != nil {
				utils.ErrIn(err, "Receive from leader")
				continue
			}
			s.looper.Push(msg)
		}
	}()
}
func (s *server) handleJoinleader() {
	//todo a lot
	s.looper.AddHandler(time.Second*10, MA_Join, func(msg Message) error {
		_, state, _ := msg.GetInfo().Info()
		serier := NewSerializer()
		switch state {
		case MS_Probe:
			fmt.Println(utils.Norf("Join Prob"))
			//getleader
			leader := NewNodeInfo()
			_, err := serier.Decode(leader, msg.GetContent())
			if err != nil {
				return err
			}
			//disconnect Referees
			//connect leader
			msg.GetInfo().SetState(MS_Ask)
			msg.GetInfo().SetTime(time.Now())

			myNodeInfo, _, _ := s.node.Info()
			_, myNodebuf, err := serier.Encode(myNodeInfo)
			if err != nil {
				return err
			}
			msg.SetContent(myNodebuf)
			err = s.node.SetLeader(leader)
			if err != nil {
				return err
			}
			err = s.node.SendToLeader(msg)
			if err != nil {
				return err
			}
		case MS_Ask:
			//add child
			child := NewNodeInfo()
			_, err := serier.Decode(child, msg.GetContent())
			if err != nil {
				return err
			}
			err = s.node.AddChild(child)
			if err != nil {
				return err
			}
			msg.GetInfo().SetState(MS_Succeed)
			s.node.ReplyToChild(child.GetName(), msg)
		case MS_Succeed:
			fmt.Println(utils.Infof("Join Succeed"))
		case MS_Failed:
			fmt.Println(utils.Warningf("Join Faild"))
			s.joinLeaderProbe()
		case MS_Error:
			utils.ErrIn(errors.New(string(msg.GetContent())), "Join Leader")
		}
		return nil
	})
}
func (s *server) handleHeartbeat() {
	verbs := true
	var timer *time.Timer

	s.looper.AddHandler(0, MA_Heartbeat, func(msg Message) error {
		_, state, _time := msg.GetInfo().Info()
		var err error
		timeout := s.timeoutDurantion
		switch state {
		case MS_Local:
			//local print time
			ttime := _time
			if verbs {
				fmt.Println(ttime.Format("2006.01.02 15:04:05"))
			}
			msg.GetInfo().SetState(MS_Probe)
			s.looper.Push(msg)
		case MS_Probe:
			//set timer and do something when time out
			if timer == nil {
				timer = time.AfterFunc(timeout, func() {
					//do time out
					fmt.Println(utils.Errf("time out"))
					timer.Stop()
					timer = nil
				})
			}
			msg.GetInfo().SetState(MS_Ask)
			err = s.node.BroadcastToChildren(msg)
		case MS_Ask:
			//someone ask for heartbeat then he must be alive
			fmt.Println(utils.Infof("on time"))
			timer.Stop()
			timer = nil
		case MS_Error:
			utils.ErrIn(err, "HeartBeat")
		}

		return err
	})
}
func (s *server) heartbeating() {
	s.looper.SetInterval(s.heartbeatDuration, func(t time.Time) error {
		info := NewMessageInfo()
		info.SetAcion(MA_Heartbeat)
		info.SetState(MS_Local)
		info.SetTime(time.Now())
		msg := NewMessage(info, []byte(""))
		s.looper.Push(msg)
		return nil
	})
}
func (s *server) joinLeaderProbe() {
	var leaderinfo NodeInfo
	if len(os.Args) >= 5 {
		serier := NewSerializer()
		leaderinfo = NewNodeInfo()
		leaderinfo.SetAddr(os.Args[3], os.Args[4])
		_, leaderinfobuf, err := serier.Encode(leaderinfo)
		utils.ErrQuit(err, "Leader Wrong")
		joinMsgInfo := NewMessageInfo()
		joinMsgInfo.SetAcion(MA_Join)
		joinMsgInfo.SetState(MS_Probe)
		s.looper.Push(NewMessage(joinMsgInfo, leaderinfobuf))
	}
}
func (s *server) Listen() {
	s.looper.Loop(false)
	s.heartbeating()
	//fake thing:join leader probe
	s.joinLeaderProbe()
	s.looper.Wait()
}
