package core

import (
	"bytes"
)

type Serializable interface {
	/** read out to the buf */
	Read(buf []byte) (int, error)
	/** write in from the buf */
	Write(buf []byte) (int, error)

	Size() int
}
type Serializer struct {
}

var (
	_serializer *Serializer
)

func NewSerializer() *Serializer {
	if _serializer == nil {
		_serializer = &Serializer{}
	}
	return _serializer
}

func (s *Serializer) Encode(seris Serializable) (int, []byte, error) {
	buf := make([]byte, seris.Size())
	size, err := seris.Read(buf)
	return size, buf, err
}
func (s *Serializer) Decode(seris Serializable, buf []byte) (int, error) {
	size, err := seris.Write(buf)
	if err != nil {
		return 0, err
	} else {
		return size, nil
	}
}
func (s *Serializer) Serialize(bufs ...[]byte) []byte {
	buf := bytes.Join(bufs, []byte("\n"))
	return buf
}

var (
	_stringSet *StringSet
)

type StringSet struct {
}

func NewStringSet() *StringSet {
	if _stringSet == nil {
		_stringSet = &StringSet{}
	}
	return _stringSet
}

func (s *StringSet) IndexOf(set []string, str string) int {
	index := -1
	for i, _s := range set {
		if _s == str {
			index = i
		}
	}
	return index
}

func (s StringSet) Differ(set1 []string, set2 []string) []string {
	strs := make([]string, 0)
	for _, _s1 := range set1 {
		if s.IndexOf(set2, _s1) < 0 {
			strs = append(strs, _s1)
		}
	}
	return strs
}

func (s StringSet) Minus(str []string, sub []string) []string {
	strs := make([]string, 0)
	for _, _str := range str {
		if s.IndexOf(sub, _str) < 0 {
			strs = append(strs, _str)
		}
	}
	return strs
}
