package servers

import (
	common "github.com/gargous/flitter/common"
	"testing"
	"time"
)

func Test_Name(t *testing.T) {
	t.Parallel()
	t.Log(common.Norf("Start Name"))
	var err error
	service := NewNameService()
	server, err := NewWorker("#2", "127.0.0.1:8100")
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
	t.Log(common.Norf("End Name"))
}
