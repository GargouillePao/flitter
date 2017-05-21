package core

import (
	"github.com/gargous/flitter/common"
	"strconv"
)

type DataInfo struct {
	Key   string
	Value common.DataItem
	Count int
}

func NewDataInfo(key string) DataInfo {
	return DataInfo{
		Key:   key,
		Count: 1,
	}
}

func (d *DataInfo) Parse(msg Message) (ok bool) {
	count, ok := msg.GetContent(4)
	if !ok {
		return
	}
	key, _ := msg.GetContent(2)
	value, _ := msg.GetContent(3)
	d.Key = string(key)

	if len(value) > 0 {
		err := d.Value.Parse(value)
		if err != nil {
			ok = false
			return
		}
	}
	ok = true
	var err error
	d.Count, err = strconv.Atoi(string(count[:]))
	if err != nil {
		ok = false
	}
	return
}

func (d DataInfo) AppendToMsg(msg Message) (err error) {
	msg.AppendContent([]byte(d.Key))
	if d.Value.Data == nil || len(d.Value.Data) <= 0 {
		msg.AppendContent([]byte("nil"))
	} else {
		b, err := d.Value.Bytes()
		if err != nil {
			return err
		}
		msg.AppendContent(b)
	}
	msg.AppendContent([]byte(strconv.Itoa(d.Count)))
	return
}

func (d DataInfo) AssertCount(all func() (ok bool), none func() (ok bool), one func() (ok bool)) (ok bool) {
	switch {
	case d.Count < 0:
		ok = all()
	case d.Count == 0:
		ok = none()
	case d.Count > 0:
		ok = one()
	}
	return
}

type ClientInfo struct {
	name string
	path NodePath
}

func (c *ClientInfo) SetName(name string) {
	c.name, _ = common.ParseClientName(string(c.path), name)
}

func (c *ClientInfo) SetPath(path NodePath) {
	c.path = path
}

func (c ClientInfo) GetName() (name string) {
	c.name, _ = common.ParseClientName(string(c.path), c.name)
	return c.name
}

func (c ClientInfo) GetPath() (path NodePath) {
	return c.path
}

func NewClientInfo(name string, path NodePath) (cinfo ClientInfo) {
	cinfo.path = path
	cinfo.name, _ = common.ParseClientName(string(path), name)
	return
}

func (c *ClientInfo) Parse(msg Message) (ok bool) {
	contents := msg.GetContents()
	if len(contents) <= 2 {
		ok = false
		return
	}
	cinfo := NewClientInfo(string(contents[0]), NodePath(contents[1]))
	c.path = cinfo.path
	c.name = cinfo.name
	ok = true
	return
}
func (c *ClientInfo) AppendToMsg(msg Message) {
	msg.ClearContent()
	msg.AppendContent([]byte(c.GetName()))
	msg.AppendContent([]byte(c.GetPath()))
}
