package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func clearInit() {
	os.RemoveAll("share")
	os.RemoveAll("client")
	os.RemoveAll("server")
}

func genClient() (err error) {
	err = os.Mkdir("client", os.ModePerm)
	if err != nil {
		return
	}
	err = os.Mkdir("client/internal", os.ModePerm)
	if err != nil {
		return
	}
	err = ioutil.WriteFile("client/app.go", []byte(`
package main

import (
	"../internal"
)

func main() {
	internal.Start()
}
	`), os.ModePerm)
	if err != nil {
		return
	}
	err = ioutil.WriteFile("client/internal/main.go", []byte(`
package internal

import (
	"github.com/gargous/flitter"
	"log"
)

var cli flitter.Client

func init() {
	cli = flitter.NewClient("127.0.0.1:8080")
}

func Start() {
	err := cli.Start()
	if err != nil {
		log.Fatal(err)
	}
}
	`), os.ModePerm)
	return
}

func genInit() (err error) {
	err = os.Mkdir("share", os.ModePerm)
	if err != nil {
		return
	}
	err = os.Mkdir("share/proto", os.ModePerm)
	if err != nil {
		return
	}
	err = os.Mkdir("share/data", os.ModePerm)
	if err != nil {
		return
	}
	err = genClient()
	if err != nil {
		return
	}
	err = os.Mkdir("server", os.ModePerm)
	if err != nil {
		return
	}
	temps, err := readJson(os.Getenv("GOPATH") + "/src/github.com/gargous/flitter/fltool/template.json")
	if err != nil {
		return
	}
	msgs := temps["msg"]
	msgDic := msgs.(map[string]interface{})
	err = writeJson("share/proto/msg.json", msgs)
	if err != nil {
		return
	}
	pbf, err := os.Create("share/proto/msg.proto")
	if err != nil {
		return
	}
	packName := msgDic["package"]
	delete(msgDic, "package")
	pbStrs := make([]string, 0)
	pbStrs = append(pbStrs, "syntax = \"proto3\";")
	pbStrs = append(pbStrs, fmt.Sprintf("package %v;", packName))
	for _, v := range msgDic {
		pbStrs = append(pbStrs, fmt.Sprintf("message %v {\n}", v))
	}
	_, err = pbf.WriteString(strings.Join(pbStrs, "\n\n"))
	if err != nil {
		return
	}
	return
}

func genMsgGen() (err error) {
	err = runCmd("protoc --gofast_out=. ./share/proto/*.proto")
	if err != nil {
		return
	}
	err = GenMsgs("./share/proto/msg.json")
	return
}

func runCmd(cmdstr string) (err error) {
	log.Println("Run", cmdstr)
	cmd := exec.Command("sh", "-c", cmdstr)
	outp, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New(fmt.Sprintf("%v:%s", err, outp))
		return
	}
	return
}

func main() {
	app := cli.NewApp()
	app.Name = "fltool"
	app.Usage = "tool for flproject"
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "make init",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "-f",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("f") {
					clearInit()
				}
				return genInit()
			},
		},
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "make new [app name]",
			Action: func(c *cli.Context) error {
				log.Println(c.Args()[0])
				return nil
			},
		},
		{
			Name:    "msggen",
			Aliases: []string{"mg"},
			Usage:   "make msggen [msg path]",
			Action: func(c *cli.Context) error {
				return genMsgGen()
			},
		},
		{
			Name:    "robot",
			Aliases: []string{"rt"},
			Usage:   "make rt",
			Action: func(c *cli.Context) error {
				return runCmd("go run client/app.go")
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
