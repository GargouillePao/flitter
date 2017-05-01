package servers

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
)

func init() {
	gob.Register(DataItem{})
	gob.Register(map[string]DataItem{})
	gob.Register(map[string]uint32{})
	gob.Register(dataSet{})
}

type DataItem struct {
	Data    []byte
	Version uint32
}

func (d *DataItem) Parse(value interface{}) (err error) {
	b := bytes.NewBuffer(nil)
	if value == nil {
		err = errors.New("Value is Nil")
		return
	}
	switch valued := value.(type) {
	case []byte:
		b = bytes.NewBuffer(valued)
		decoder := gob.NewDecoder(b)
		err = decoder.Decode(d)
		if err != nil {
			d.Data = valued
			d.Version = 0
			err = nil
		}
	case string:
		b = bytes.NewBuffer([]byte(valued))
		decoder := gob.NewDecoder(b)
		err = decoder.Decode(d)
		if err != nil {
			d.Data = []byte(valued)
			d.Version = 0
			err = nil
		}
	case DataItem:
		d.Data = valued.Data
		d.Version = valued.Version
	default:
		ecoder := gob.NewEncoder(b)
		err = ecoder.Encode(valued)
		if err != nil {
			return
		}
		b = bytes.NewBuffer(b.Bytes())
		decoder := gob.NewDecoder(b)
		err = decoder.Decode(d)
		if err != nil {
			ecoder := gob.NewEncoder(b)
			err = ecoder.Encode(valued)
			if err != nil {
				return
			}
			d.Data = b.Bytes()
			d.Version = 0
			err = nil
		}
	}
	return
}

func (d *DataItem) Bytes() (buf []byte, err error) {
	b := bytes.NewBuffer(nil)
	ecoder := gob.NewEncoder(b)
	err = ecoder.Encode(d)
	if err != nil {
		return
	}
	buf = b.Bytes()
	return
}

type DataSet interface {
	Set(key string, value DataItem)
	Get(key string) (value DataItem, ok bool)
	String() string
}

type dataSet struct {
	datamutex sync.Mutex
	Datas     map[string]DataItem
}

func (s *dataSet) Set(key string, value DataItem) {
	s.datamutex.Lock()
	if s.Datas == nil {
		s.Datas = make(map[string]DataItem)
	}
	oldvalue, ok := s.Get(key)
	if ok {
		if value.Version < oldvalue.Version && oldvalue.Version != 0 {
			value.Data = oldvalue.Data
			value.Version = oldvalue.Version
			return
		}
	}
	//value.Version++
	s.Datas[key] = value
	s.datamutex.Unlock()
}
func (s *dataSet) Get(key string) (value DataItem, ok bool) {
	if s.Datas == nil {
		ok = false
		return
	}
	value, ok = s.Datas[key]
	if !ok {
		return
	}
	return
}
func (s dataSet) String() (str string) {
	if s.Datas == nil {
		str = "<nil>"
	}
	str = ""
	for k, v := range s.Datas {
		str += fmt.Sprintf("\nkey:%s\nvalue:%v\n]", k, v)
	}
	str = str[1:]
	return
}
