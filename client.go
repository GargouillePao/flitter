package flitter

import (
	"github.com/chzyer/readline"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"strings"
)

type Client interface {
	Start() error
	Register(headId uint32, act func(interface{}) error)
	Invoke(cmd string, desp string, usage string, act func(...string))
	Send(head uint32, body proto.Message) error
}

type cmdHandler struct {
	cb    func(...string)
	desp  string
	usage string
	pri   readline.PrefixCompleterInterface
}

type client struct {
	d    *dealer
	mp   MsgProcesser
	cmds map[string]cmdHandler
}

func NewClient(addr string, c map[uint32]func() proto.Message) Client {
	d := newDealer(nil)
	d.addr = addr
	prcser := NewMsgProcesser(c, true)
	return &client{
		d:    d,
		mp:   prcser,
		cmds: make(map[string]cmdHandler),
	}
}

func (c *client) Register(headId uint32, act func(interface{}) error) {
	c.mp.Register(headId, func(d *dealer, msg interface{}) error {
		return act(msg)
	})
}

func (c *client) Send(head uint32, body proto.Message) error {
	return c.d.Send(head, body, true)
}

func (c *client) Invoke(cmd string, desp string, usage string, cb func(...string)) {
	c.cmds[cmd] = cmdHandler{
		pri:   readline.PcItem(cmd),
		desp:  desp,
		usage: usage,
		cb:    cb,
	}
}

func (c *client) Start() (err error) {
	err = c.d.Connect()
	if err != nil {
		return
	}
	prcs := make([]readline.PrefixCompleterInterface, 0)
	for _, cmd := range c.cmds {
		prcs = append(prcs, cmd.pri)
	}
	helps := make([]readline.PrefixCompleterInterface, 0)
	for cmdName, _ := range c.cmds {
		helps = append(helps, readline.PcItem(cmdName))
	}
	prcs = append(prcs, readline.PcItem("help", helps...))
	var completer = readline.NewPrefixCompleter(
		prcs...,
	)
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
	})
	if err != nil {
		return
	}
	go func() {
		err = c.d.Process(c.mp)
	}()
	for {
		line, err := l.Readline()
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "help "):
			helpCmdName := line[5:]
			helpCmdItem, ok := c.cmds[helpCmdName]
			if ok {
				log.Println(helpCmdItem.desp)
				log.Println(helpCmdItem.usage)
			} else {
				log.Println("not find cmd")
			}
		case line == "help":
			log.Println("help [Tab]")
		case line == "":
		default:
			cmdName := strings.TrimSpace(line)
			cmdItems := strings.Split(cmdName, " ")
			cmdItem, ok := c.cmds[cmdName]
			if ok {
				cmdItem.cb(cmdItems...)
			} else {
				log.Println("not find cmd")
			}
		}
	}
	return
}
