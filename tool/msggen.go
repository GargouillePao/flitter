package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/dchest/siphash"
	"fmt"
	"strings"
)

func main() {
	msgPath := os.Args[1]
	buf,err := ioutil.ReadFile(msgPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	msgs := make(map[string]interface{})
	err = json.Unmarshal(buf, &msgs)
	if err != nil {
		fmt.Println(err)
		return
	}
	outstr := make([]string,0)
	for k :=range(msgs) {
		h := siphash.Hash(0x12340000, 0x00005678, []byte(k))
		outstr = append(outstr, fmt.Sprintf("const %s uint64 = %d",k,h))
	}
	outPath := msgPath[0:strings.LastIndex(msgPath,".")] + ".go"
	ioutil.WriteFile(outPath, []byte(strings.Join(outstr,"\n")), 0777)
}
