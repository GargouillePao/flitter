package app

import (
	"testing"
	"time"
)

type TestApp struct {
	t         *testing.T
	panicword string
}

func (ta *TestApp) Panic() {
	panic(ta.panicword)
}

func (ta *TestApp) OnStart() {
	time.AfterFunc(time.Second*2, func() {
		go func() {
			ta.Panic()
		}()
	})
}
func (ta *TestApp) OnEnd(err interface{}) {
	if err != ta.panicword {
		ta.t.Fatal(err)
	}
}
func TestInit(t *testing.T) {
	myapp := &TestApp{t, "Hello"}
	mysup := New(myapp)
	mysup.Init()
}
