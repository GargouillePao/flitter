package app

import (
	"fmt"
	"os"
	"os/signal"
)

//Supervisor Supervisor
type Supervisor interface {
	Init()
}

type supervisor struct {
	a Application
}

//New NewSupervisor
func New(a Application) Supervisor {
	return &supervisor{a}
}

func (s supervisor) Init() {
	c := make(chan os.Signal)
	signal.Notify(c)
	s.a.OnStart()
	sig := <-c
	fmt.Println("退出信号", s)
	s.a.OnEnd(sig)
}
