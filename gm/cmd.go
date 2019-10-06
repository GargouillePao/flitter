package gm

import (
	"io"
	"strings"

	"github.com/chzyer/readline"
)

const (
	//StrSpace StrSpace
	StrSpace = " "
	//StrEmpty StrEmpty
	StrEmpty = ""
)

//GM GM
type GM interface {
	On(name string, cb func(...string))
	Init() (err error)
}

type cmdGM struct {
	handers  map[string]func(...string)
	instance *readline.Instance
}

func (c *cmdGM) On(name string, cb func(...string)) {
	c.handers[name] = cb
}

func (c *cmdGM) Init() (err error) {
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
		if hcb, ok := c.handers[hname]; ok {
			hcb(hattrs...)
		}
	}
	return
}

//NewCmdGM NewCmdGM
func NewCmdGM() GM {
	return &cmdGM{handers: make(map[string]func(...string))}
}
