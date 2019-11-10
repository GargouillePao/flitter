package gm

import (
	"io"
	"reflect"
	"strings"

	"github.com/chzyer/readline"
)

const (
	//StrSpace StrSpace
	StrSpace = " "
	//StrEmpty StrEmpty
	StrEmpty = ""
)

type GM struct {
	handers  map[string]func(...string)
	instance *readline.Instance
}

func (c *GM) On(name string, cb func(...string)) {
	c.handers[name] = cb
}

type Cmd interface {
	Name() string
}

func (c *GM) AddCmd(cmd Cmd) {
	cval := reflect.ValueOf(cmd)
	for i := 0; i < cval.NumMethod(); i++ {
		mt := cval.Method(i)
		mn := mt.Type().Name()
		onAt := strings.IndexAny(mn, "On")
		if onAt == 0 {
			rname := mn[onAt+2:]
			rnamel := strings.ToLower(rname)
			c.handers[cmd.Name()+" "+rnamel] = func(attrs ...string) {
				nattrs := make([]reflect.Value, 0)
				for _, attr := range attrs {
					nattrs = append(nattrs, reflect.ValueOf(attr))
				}
				mt.Call(nattrs)
			}
			helpName := "Help" + rname
			ht := cval.MethodByName(helpName)
			if ht.IsNil() {
				c.handers[cmd.Name()+" "+rnamel+" help"] = func(attrs ...string) {
					ht.Call([]reflect.Value{})
				}
			}
		}
	}
}

func (c *GM) Trigger(name string, attrs ...string) (err error) {
	if hcb, ok := c.handers[name]; ok {
		hcb(attrs...)
	}
	return
}

func (c *GM) Init() (err error) {
	pc := make([]readline.PrefixCompleterInterface, 0)
	for hname := range c.handers {
		pc = append(pc, readline.PcItem(hname))
	}
	completer := readline.NewPrefixCompleter(pc...)
	c.instance, err = readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return
	}
	for {
		line, err := c.instance.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		hname := StrEmpty
		hattrs := make([]string, 0)
		for _, lineword := range strings.Split(line, StrSpace) {
			if lineword != StrSpace {
				lineword = strings.TrimSpace(lineword)
				if hname == StrEmpty {
					hname = lineword
				} else {
					hattrs = append(hattrs, lineword)
				}
			}
		}
		c.Trigger(hname, hattrs...)
	}
	return
}

//NewCmdGM NewCmdGM
func NewCmdGM() *GM {
	return &GM{handers: make(map[string]func(...string))}
}
