package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_ErrIn(t *testing.T) {
	t.Log(Norf("ErrIn Start"))
	filename := fmt.Sprintf("%d", time.Now().UnixNano()) + ".log"
	fd, err := os.Create(filename)
	if err != nil {
		t.Fatal(Errf("Create File %v", err))
	}
	SetLogAddress(fd)
	ErrIn(errors.New("Err1"))
	ErrIn(errors.New("Err2"))
	ErrIn(errors.New("Err3"))
	ErrIn(errors.New("Err4"))
	ErrIn(errors.New("Err5"))
	time.Sleep(time.Second * 2)
	fd, err = os.Open(filename)
	if err != nil {
		t.Fatal(Errf("Open File %v", err))
	}
	buf := make([]byte, 1024)
	n, err := fd.Read(buf)
	if err != nil {
		t.Fatal(Errf("Read File %v", err))
	}
	bufstr := strings.Trim(string(buf[:n]), " ")
	if bufstr == "Error(QAQ):Err1\nError(QAQ):Err2\nError(QAQ):Err3\nError(QAQ):Err4\nError(QAQ):Err5\n" {
		t.Log(Infof("Succeed %s", bufstr))
	} else {
		t.Fatal(Errf("Failed %s", bufstr))
	}
	t.Log(Norf("ErrIn End"))
}
