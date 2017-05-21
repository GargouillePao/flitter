package common

import (
	"fmt"
	"strings"
	"time"
)

func ParseClientName(hostname string, name string) (targename string, ok bool) {
	if hostname == "" {
		targename = ""
		ok = false
		return
	}
	_, _, ok = DeparseClientName(name)
	if ok {
		targename = name
		return
	}
	targename = fmt.Sprintf("%s|%d|%s", hostname, time.Now().UnixNano(), name)
	ok = true
	return
}
func DeparseClientName(targename string) (hostname string, name string, ok bool) {
	names := strings.Split(targename, "|")
	if len(names) == 3 {
		hostname = names[0]
		name = names[2]
		ok = true
	} else {
		ok = false
	}
	return
}
