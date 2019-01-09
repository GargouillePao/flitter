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
	Rejister(headId uint32, act func(Dealer, interface{}) error)
	Invoke(cmd string, desp string, usage string, act func(Dealer, ...string))
}

type cmdHandler struct {
	cb    func(Dealer, ...string)
	desp  string
	usage string
	pri   readline.PrefixCompleterInterface
}

type client struct {
	addr string
	d    Dealer
	mp   MsgProcesser
	cmds map[string]cmdHandler
}

func NewClient(addr string, c map[uint32]func() proto.Message) Client {
	d := NewDealer()
	prcser := NewMsgProcesser(c)
	return &client{
		addr: addr,
		d:    d,
		mp:   prcser,
		cmds: make(map[string]cmdHandler),
	}
}

func (c *client) Rejister(headId uint32, act func(Dealer, interface{}) error) {
	c.mp.Rejister(headId, act)
}

func (c *client) Invoke(cmd string, desp string, usage string, cb func(Dealer, ...string)) {
	c.cmds[cmd] = cmdHandler{
		pri:   readline.PcItem(cmd),
		desp:  desp,
		usage: usage,
		cb:    cb,
	}
}

func (c *client) Start() (err error) {
	err = c.d.Connect(c.addr)
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
				cmdItem.cb(c.d, cmdItems...)
			} else {
				log.Println("not find cmd")
			}
		}
	}
	return
}
