package common

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_ErrIn(t *testing.T) {
	t.Log(Norf("ErrIn Start"))
	filename, err := InitLog(true, "../logs")
	if err != nil {
		t.Fatal(Errf("Init Log %v", err))
	}
	ErrIn(errors.New("Err1"))
	ErrIn(errors.New("Err2"))
	ErrIn(errors.New("Err3"))
	ErrIn(errors.New("Err4"))
	ErrIn(errors.New("Err5"))
	time.Sleep(time.Second * 1)
	fd, err := os.Open(filename)
	if err != nil {
		t.Fatal(Errf("Open File %v", err))
	}
	buf := make([]byte, 1024)
	n, err := fd.Read(buf)
	if err != nil {
		t.Fatal(Errf("Read File %v", err))
	}
	bufstr := strings.Trim(string(buf[:n]), " ")
	shouldbe := `Error(QAQ):[at utils_test.go 17]Err1
Error(QAQ):[at utils_test.go 18]Err2
Error(QAQ):[at utils_test.go 19]Err3
Error(QAQ):[at utils_test.go 20]Err4
Error(QAQ):[at utils_test.go 21]Err5
`
	os.Remove(filename)
	if bufstr == shouldbe {
		t.Log(Infof("Succeed %s", bufstr))
	} else {
		t.Fatal(Errf("Failed \nNow\n%s\nShould\n%s", bufstr, shouldbe))
	}

	t.Log(Norf("ErrIn End"))
}
