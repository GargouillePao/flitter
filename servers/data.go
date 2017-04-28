package servers

import (
	"encoding/gob"
	"fmt"
	"sync"
)

func init() {
	gob.Register(map[string][]byte{})
	gob.Register(map[string]uint32{})
	gob.Register(dataSet{})
}

type DataSet interface {
	Set(name string, data []byte)
	Get(name string) (data []byte, ok bool)
	String() string
}

type dataSet struct {
	datamutex sync.Mutex
	Datas     map[string][]byte
	Versions  map[string]uint32
}

func (s *dataSet) Set(key string, value []byte) {
	s.datamutex.Lock()
	if s.Datas == nil {
		s.Datas = make(map[string][]byte)
		s.Versions = make(map[string]uint32)
	}
	s.Datas[key] = value
	s.Versions[key]++
	s.datamutex.Unlock()
}
func (s *dataSet) Get(key string) (value []byte, ok bool) {
	if s.Datas == nil {
		s.Datas = make(map[string][]byte)
	}
	value, ok = s.Datas[key]
	return
}
func (s dataSet) String() (str string) {
	if s.Datas == nil {
		str = "<nil>"
	}
	str = ""
	for k, v := range s.Datas {
		str += fmt.Sprintf("\nkey:%s[\n\tvalue:%v\n\tversion:%d\n]", k, v, s.Versions[k])
	}
	str = str[1:]
	return
}
