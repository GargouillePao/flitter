package common

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"
	"sync"
)

func init() {
	gob.Register([]byte{})
	gob.Register(DataItem{})
	gob.Register(map[string]DataItem{})
	gob.Register(map[string]uint32{})
	gob.Register(BaseDataSet{})
}

type DataItem struct {
	Data    []byte
	Version uint32
	buffer  *bytes.Buffer
}

func (d *DataItem) Parse(value interface{}) (err error) {
	if value == nil {
		err = errors.New("Value is Nil")
		return
	}
	if d.buffer == nil {
		d.buffer = bytes.NewBuffer(nil)
	} else {
		d.buffer.Reset()
	}
	switch valued := value.(type) {
	case []byte:
		_, err = d.buffer.Write(valued)
		if err != nil {
			return
		}
		err = gob.NewDecoder(d.buffer).Decode(d)
		if err != nil {
			d.Data = valued
			d.Version = 0
			err = nil
		}
	case string:
		_, err = d.buffer.Write([]byte(valued))
		if err != nil {
			return
		}
		err = gob.NewDecoder(d.buffer).Decode(d)
		if err != nil {
			d.Data = []byte(valued)
			d.Version = 0
			err = nil
		}
	case DataItem:
		d.Data = valued.Data
		d.Version = valued.Version
	default:
		err = gob.NewEncoder(d.buffer).Encode(valued)
		if err != nil {
			return
		}
		err = gob.NewDecoder(d.buffer).Decode(d)
		if err != nil {
			d.buffer.Reset()
			err = gob.NewEncoder(d.buffer).Encode(valued)
			if err != nil {
				return
			}
			d.Data = d.buffer.Bytes()
			d.Version = 0
			err = nil
		}
	}
	return
}
func (d *DataItem) Bytes() (buf []byte, err error) {
	buf = d.buffer.Bytes()
	if len(buf) > 0 {
		return
	}
	d.buffer.Reset()
	ecoder := gob.NewEncoder(d.buffer)
	err = ecoder.Encode(d)
	if err != nil {
		return
	}
	buf = d.buffer.Bytes()
	return
}

func (d DataItem) String() string {
	showLen := 3
	if len(d.Data) > showLen {
		dstrstart := fmt.Sprintf("%v", d.Data[:showLen])
		dstrend := fmt.Sprintf("%v", d.Data[len(d.Data)-showLen:])
		return fmt.Sprintf("[%v ... %v],%d", dstrstart[1:len(dstrstart)-1], dstrend[1:len(dstrend)-1], d.Version)
	} else {
		return fmt.Sprintf("%v,%d", d.Data, d.Version)
	}

}

type DataLog struct {
	Log     map[uint32]DataItem
	Locker  string
	Version uint32
}

func (d *DataLog) use(version uint32) {
	d.Version = version
}

func (d *DataLog) iterate() {
	d.Version++
}

func (d *DataLog) reverse() {
	delete(d.Log, d.Version)
	if d.Version > 0 {
		d.Version--
	}
}
func (d DataLog) String() (str string) {
	logstr := ""
	for k, v := range d.Log {
		logstr += fmt.Sprintf("\n\tVersion:%d\n\tValue:%v\n", k, v)
	}
	str = fmt.Sprintf("Version:%d\nValue:[%s]\nLocker:%v", d.Version, logstr, d.Locker)
	return
}

type DataSet interface {
	Grant(key string, value DataItem) DataItem
	Iterate(key string) (ok bool)
	Reverse(key string) (ok bool)
	Use(key string, version uint32) (ok bool)
	Set(key string, value DataItem) (ok bool)
	Get(key string) (value DataItem, ok bool)
	IsLocked(holder string, key string) (locked bool)
	Lock(holder string, key string) (ok bool)
	Unlock(key string)
	String() string
}

func NewDataSet() DataSet {
	return &BaseDataSet{
		Datas: make(map[string]DataLog),
	}
}

type BaseDataSet struct {
	datamutex sync.Mutex
	Datas     map[string]DataLog
}

func (s BaseDataSet) Grant(key string, value DataItem) DataItem {
	_, ok := s.Get(key)
	if !ok {
		return value
	}
	value.Version = s.Datas[key].Version + 1
	return value
}

func (s *BaseDataSet) Iterate(key string) (ok bool) {
	_, ok = s.Get(key)
	if !ok {
		return
	}
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	datalog := s.Datas[key]
	datalog.iterate()
	s.Datas[key] = datalog
	return
}

func (s *BaseDataSet) Reverse(key string) (ok bool) {
	_, ok = s.Get(key)
	if !ok {
		return
	}
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	datalog := s.Datas[key]
	datalog.reverse()
	s.Datas[key] = datalog
	return
}

func (s *BaseDataSet) Use(key string, version uint32) (ok bool) {
	_, ok = s.Get(key)
	if !ok {
		return
	}
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	datalog := s.Datas[key]
	datalog.use(version)
	s.Datas[key] = datalog
	return
}

func (s BaseDataSet) IsLocked(holder string, key string) (locked bool) {
	_, ok := s.Get(key)
	if !ok {
		locked = false
		return
	}
	log := s.Datas[key]
	if log.Locker == "" {
		locked = false
		return
	}
	if log.Locker == holder {
		locked = false
	} else {
		locked = true
	}
	return
}

func (s *BaseDataSet) Lock(holder string, key string) (ok bool) {
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	_, ok = s.Get(key)
	if !ok {
		s.set(key, DataItem{
			Data: []byte(""),
		})
	}
	log := s.Datas[key]
	if log.Locker != "" && log.Locker != holder {
		ok = false
		return
	}
	log.Locker = holder
	ok = true
	s.Datas[key] = log
	return
}

func (s *BaseDataSet) Unlock(key string) {
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	_, ok := s.get(key)
	if !ok {
		return
	}
	log := s.Datas[key]
	log.Locker = ""
	s.Datas[key] = log
	return
}

func (s *BaseDataSet) set(key string, value DataItem) (ok bool) {
	if s.Datas == nil {
		s.Datas = make(map[string]DataLog)
	}
	valueLog, ok := s.Datas[key]
	if !ok {
		valueLog = DataLog{
			Version: 0,
			Log:     make(map[uint32]DataItem),
			Locker:  "",
		}
	}
	v, ok := valueLog.Log[value.Version]
	if !ok || v.Data == nil || len(v.Data) <= 0 {
		valueLog.Log[value.Version] = value
		if value.Version > valueLog.Version {
			valueLog.Version = value.Version
		}
		ok = true
	} else {
		ok = false
	}
	s.Datas[key] = valueLog
	return
}

func (s BaseDataSet) get(key string) (value DataItem, ok bool) {
	if s.Datas == nil {
		ok = false
		return
	}
	valueLog, ok := s.Datas[key]
	if !ok {
		return
	}
	if valueLog.Log == nil {
		ok = false
		return
	}
	value, ok = valueLog.Log[valueLog.Version]
	if !ok {
		return
	}
	return
}

func (s *BaseDataSet) Set(key string, value DataItem) (ok bool) {
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	return s.set(key, value)
}
func (s BaseDataSet) Get(key string) (value DataItem, ok bool) {
	value, ok = s.get(key)
	if value.Data == nil || len(value.Data) <= 0 {
		ok = false
		return
	}
	return
}
func (s BaseDataSet) String() (str string) {
	if s.Datas == nil {
		str = "<nil>"
	}
	str = ""
	for k, v := range s.Datas {
		vstr := fmt.Sprintf("%v", v)
		vstr = strings.Join(strings.Split(vstr, "\n"), "\n\t")
		str += fmt.Sprintf("\nKey:%s\nData:[\n\t%s\n]", k, vstr)
	}
	if len(str) == 0 {
		return
	}
	str = str[1:]
	return
}
