package main

import (
	"encoding/json"
	"fmt"
	"github.com/dchest/siphash"
	"io/ioutil"
	"os"
	"strings"
)

func hash(str string) (output uint32) {
	h := siphash.Hash(0x12340000, 0x00005678, []byte(str))
	hu := h >> 32
	hd := (h << 32) >> 32
	output = uint32(hd | hu)
	return
}

func output(enumHeader string, enumBody map[string]uint32, enumFormat string, enumFooter string, funcHeader string, funcBody map[uint32]string, funcFormat string, funcFooter string, outPath string) {
	outstr := make([]string, 0)
	outstr = append(outstr, enumHeader)
	for k, v := range enumBody {
		outstr = append(outstr, fmt.Sprintf(enumFormat, k, v))
	}
	outstr = append(outstr, enumFooter)
	outstr = append(outstr, funcHeader)
	for k, v := range funcBody {
		outstr = append(outstr, fmt.Sprintf(funcFormat, k, v))
	}
	outstr = append(outstr, funcFooter)
	ioutil.WriteFile(outPath, []byte(strings.Join(outstr, "\n")), 0777)
}

func main() {
	msgPath := os.Args[1]
	buf, err := ioutil.ReadFile(msgPath)
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
	pacName, ok := msgs["package"]
	if !ok {
		fmt.Printf("Got No Package\n")
		return
	}
	delete(msgs, "package")
	outmap := make(map[string]uint32)
	checkmap := make(map[uint32]string)
	funcmap := make(map[uint32]string)

	for k, v := range msgs {
		ht := hash(k)
		outmap[k] = ht
		oldK, ok := checkmap[ht]
		if ok {
			fmt.Printf("Same Hash Result Of %s and %s\n", k, oldK)
			return
		}
		checkmap[ht] = k
		funcmap[ht] = v.(string)
	}
	output(
		"package "+pacName.(string)+"\n",
		outmap,
		"const %s uint32 = %d",
		"",
		"var MsgCreator map[uint32]func()interface{} = map[uint32]func()interface{}{",
		funcmap,
		"\t%d : func()interface{} { return &%s{}},",
		"}",
		msgPath[0:strings.LastIndex(msgPath, ".")]+".go")
	output(
		"namespace "+pacName.(string)+" {",
		outmap,
		"\tpublic const int %s = %d;",
		"",
		"\tpublic map[] msgCreators = {",
		funcmap,
		"\t\t%d : () => &%s{},",
		"\t}\n}",
		msgPath[0:strings.LastIndex(msgPath, ".")]+".cs")
}
