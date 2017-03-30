package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	//"io"
	"errors"
	"os"
	//"unsafe"
)

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
		errstr := Errf("Error(QAQ):[%s]  %s", err.Error(), info)
		fmt.Println(errstr)
		ok = true
	} else {
		ok = false
	}
	return
}

func ErrAppend(err error, info ...string) error {
	return errors.New(fmt.Sprintf("[%s]:%s", err.Error(), info))
}
func ErrIn(channel chan error, err error, info ...string) (ok bool) {
	ok = false
	if err != nil {
		ok = true
		channel <- ErrAppend(err, info...)
	}
	return
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
