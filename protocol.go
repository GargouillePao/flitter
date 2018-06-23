package flitter

import (
	"fmt"
	"github.com/dchest/siphash"
	"reflect"
)

type MessageProcessor struct {
	headers map[uint64]string
	bodies  map[uint64]reflect.Type
}

func (m *MessageProcessor) Register(header string, body interface{}) {
	bodyType := reflect.TypeOf(body)
	h := siphash.Hash(0x12340000, 0x00005678, header)
	m.header[h] = header
	m.bodies[h] = bodyType.Kind()
}
