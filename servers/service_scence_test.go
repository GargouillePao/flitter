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
	server, err := NewWorker("#1", "127.0.0.1:8900")
	if err != nil {
		t.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.Norf("End Watch"))
}
