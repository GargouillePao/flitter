package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
	//"unsafe"
)

var __errors chan error
var __logAddr io.Writer
var __verbs bool = false

func InitLog(verb bool, fname string) (filename string, err error) {
	if fname != "" {
		filename = path.Join(fname, fmt.Sprintf("%d", time.Now().UnixNano())+".log")
		fd, err := os.Create(filename)
		SetLogAddress(fd)
		return filename, err
	} else {
		SetLogAddress(os.Stdout)
	}
	__verbs = verb
	return "", nil
}

func SetLogAddress(w io.Writer) {
	__logAddr = w
}
func init() {
	__errors = make(chan error, 100)
	go func() {
		for {
			err := <-__errors
			ErrInfo(err)
		}
	}()
}

func ErrIn(err error, info ...string) (ok bool) {
	ok = false
	if err != nil {
		ok = true
		_, file, line, _ := runtime.Caller(1)
		__errors <- ErrAppend(
			ErrAppend(
				errors.New(fmt.Sprintf("[at %v %v]", path.Base(file), line)),
				err.Error(),
			),
			info...,
		)
	}
	return
}

func CloseError() {
	close(__errors)
}

func Logf(logType func(format string, item ...interface{}) string, format string, item ...interface{}) {
	if __verbs {
		log := fmt.Sprintf(format, item...)
		logwithcolor := logType(log)
		if __logAddr != nil {
			if __logAddr == os.Stdout {
				fmt.Fprintln(__logAddr, logwithcolor)
			} else {
				fmt.Fprintln(__logAddr, log)
			}
		}
	}
}

func Infof(format string, item ...interface{}) string {
	log := fmt.Sprintf(format, item...)
	info := fmt.Sprintf("%c[%d;%dm%s%c[0m\n", 0x1B, 1, 36, log, 0x1B)
	return info
}
func Errf(format string, item ...interface{}) string {
	log := fmt.Sprintf(format, item...)
	err := fmt.Sprintf("%c[%d;%dm%s%c[0m\n", 0x1B, 1, 31, log, 0x1B)
	return err
}
func Warningf(format string, item ...interface{}) string {
	log := fmt.Sprintf(format, item...)
	warning := fmt.Sprintf("%c[%d;%dm%s%c[0m\n", 0x1B, 1, 33, log, 0x1B)
	return warning
}
func Norf(format string, item ...interface{}) string {
	log := fmt.Sprintf(format, item...)
	nor := fmt.Sprintf("%c[%d;%dm%s%c[0m\n", 0x1B, 1, 32, log, 0x1B)
	return nor
}

func ErrQuit(err error, info ...string) {
	if ErrInfo(err, info...) {
		os.Exit(0)
	}
}

func ErrInfo(err error, info ...string) (ok bool) {
	if err != nil {
		Logf(Errf, "Error(QAQ):%s", err.Error()+strings.Join(info, ","))
		ok = true
	} else {
		ok = false
	}
	return
}

func ErrAppend(err error, info ...string) error {
	errstr := err.Error() + strings.Join(info, ",")
	return errors.New(errstr)
}

func Sizeof(data interface{}) int {
	size := binary.Size(data)
	return size
}

func Ecode(data interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	err := binary.Write(buffer, binary.BigEndian, data)
	if ErrInfo(err, "Buffer Write Error") {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ByteArrayToUInt16(buf []byte) (out uint16, err error) {
	bufLen := len(buf)
	if bufLen < 2 {
		err = errors.New("Cannot convert []byte to uin16")
		return
	}
	out = uint16(buf[0]) << 8
	out += uint16(buf[1])
	return
}
func ByteArrayToUInt32(buf []byte) (out uint32, err error) {
	bufLen := len(buf)
	if bufLen < 4 {
		err = errors.New("Cannot convert []byte to uin32")
		return
	}
	out = uint32(buf[0]) << (3 * 8)
	out += uint32(buf[1]) << (2 * 8)
	out += uint32(buf[2]) << 8
	out += uint32(buf[3])
	return
}
func ByteArrayToUInt64(buf []byte) (out uint64, err error) {
	bufLen := len(buf)
	if bufLen < 8 {
		err = errors.New("Cannot convert []byte to uin32")
		return
	}
	out = uint64(buf[0]) << (7 * 8)
	out += uint64(buf[1]) << (6 * 8)
	out += uint64(buf[2]) << (5 * 8)
	out += uint64(buf[3]) << (4 * 8)
	out += uint64(buf[4]) << (3 * 8)
	out += uint64(buf[5]) << (2 * 8)
	out += uint64(buf[6]) << 8
	out += uint64(buf[7])
	return
}
