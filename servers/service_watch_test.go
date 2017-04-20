package servers

import (
	utils "github.com/gargous/flitter/utils"
	"testing"
	"time"
)

func Test_Watch(t *testing.T) {
	t.Parallel()
	t.Log(utils.Norf("Start Watch"))
	var err error
	service := NewWatchService()
	server, err := NewWorker("#1", "127.0.0.1:8000")
	if err != nil {
		t.Fatal(err)
	}
	time.AfterFunc(time.Second*3, func() {
		service.Term()
	})
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.Norf("End Watch"))
}
