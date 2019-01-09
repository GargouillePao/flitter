package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dchest/siphash"
	"io/ioutil"
	"os"
	"strings"
)

func writeJson(msgPath string, v interface{}) (err error) {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return
	}
	return ioutil.WriteFile(msgPath, data, os.ModePerm)
}

func readJson(msgPath string) (msgs map[string]interface{}, err error) {
	buf, err := ioutil.ReadFile(msgPath)
	if err != nil {
		return
	}
	msgs = make(map[string]interface{})
	err = json.Unmarshal(buf, &msgs)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func hash(str string) (output uint32) {
	h := siphash.Hash(0x12340000, 0x00005678, []byte(str))
	hu := h >> 32
	hd := (h << 32) >> 32
	output = uint32(hd | hu)
	return
}

func appendstr(header string, body map[interface{}]interface{}, format string, footer string) string {
	outstr := make([]string, 0)
	outstr = append(outstr, header)
	for k, v := range body {
		outstr = append(outstr, fmt.Sprintf(format, k, v))
	}
	outstr = append(outstr, footer)
	return strings.Join(outstr, "\n")
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
	ioutil.WriteFile(outPath, []byte(strings.Join(outstr, "\n")), os.ModePerm)
}

func GenMsgs(msgPath string) (err error) {
	buf, err := ioutil.ReadFile(msgPath)
	if err != nil {
		return
	}
	msgs := make(map[string]interface{})
	err = json.Unmarshal(buf, &msgs)
	if err != nil {
		return
	}
	pacName, ok := msgs["package"]
	if !ok {
		err = errors.New("Got No Package")
		return
	}
	delete(msgs, "package")
	outmap := make(map[interface{}]interface{})
	checkmap := make(map[interface{}]interface{})
	funcmap := make(map[interface{}]interface{})

	for k, v := range msgs {
		ht := hash(k)
		outmap[k] = ht
		oldK, ok := checkmap[ht]
		if ok {
			err = errors.New(fmt.Sprintf("Same Hash Result Of %v and %v\n", k, oldK))
			return
		}
		checkmap[ht] = k
		funcmap[ht] = v
	}
	gostr := fmt.Sprintf("package %s \nimport \"github.com/golang/protobuf/proto\"\n", pacName)
	gostr += appendstr(
		"const(",
		outmap,
		"\t%s uint32 = %d",
		")\n",
	)
	gostr += appendstr(
		"var Creater map[uint32]func() proto.Message = map[uint32]func() proto.Message {",
		funcmap,
		"\t%d : func()proto.Message { return &%s{}},",
		"}\n",
	)
	gostr += appendstr(
		"var MsgNames map[uint32] string = map[uint32] string {",
		checkmap,
		"\t%d : \"%s\",",
		"}\n",
	)
	return ioutil.WriteFile(msgPath[0:strings.LastIndex(msgPath, ".")]+".go", []byte(gostr), os.ModePerm)
}
