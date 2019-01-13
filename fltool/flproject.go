package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func genFileWithTemp(projectName, inFilePath, inFileName, extName string) (err error) {
	inBuff, err := ioutil.ReadFile(filepath.Join(inFilePath, inFileName))
	if err != nil {
		return
	}
	if filepath.Ext(inFileName) != ".temp" {
		err = errors.New("Invalid File With Ext " + filepath.Ext(inFileName))
		return
	}
	seps := strings.Split(inFileName, ".")
	seps = append([]string{projectName}, seps...)
	relName := filepath.Join(seps[:len(seps)-1]...) + extName
	pathN := seps[0]
	for i := 1; i < len(seps)-1; i++ {
		os.Mkdir(pathN, os.ModePerm)
		log.Println("Mkdir", pathN)
		pathN = filepath.Join(pathN, seps[i])
	}
	inBuff = []byte(strings.Replace(string(inBuff), "_PROJECT_", projectName, -1))
	inBuff = []byte(strings.Replace(string(inBuff), "_RT_", projectName, -1))
	err = ioutil.WriteFile(relName, inBuff, os.ModePerm)
	return
}

func genInit(projectName string) (err error) {
	if projectName == "" {
		err = errors.New("No Project Name")
		return
	}
	rootPath := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "gargous", "flitter", "fltool", "template")
	err = genFileWithTemp(projectName, rootPath, "client.app.temp", ".go")
	if err != nil {
		return
	}
	err = genFileWithTemp(projectName, rootPath, "client.internal.main.temp", ".go")
	if err != nil {
		return
	}
	err = genFileWithTemp(projectName, rootPath, "Makefile.temp", "")
	if err != nil {
		return
	}
	err = genFileWithTemp(projectName, rootPath, "share.proto.msgn.temp", ".json")
	if err != nil {
		return
	}
	err = genFileWithTemp(projectName, rootPath, "share.proto.msgs.temp", ".proto")
	if err != nil {
		return
	}
	err = genMsgGen(projectName)
	return
}

func genMsgGen(projectName string) (err error) {
	cmd := exec.Command("sh", "-c", "protoc --gofast_out=. "+projectName+"/share/proto/*.proto")
	buf, err := cmd.Output()
	if err != nil {
		return errors.New(fmt.Sprintf("%v:%s", err, string(buf)))
	}
	log.Println(string(buf))
	err = GenMsgs(projectName + "/share/proto/msgn.json")
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
			Action: func(c *cli.Context) error {
				return genInit(c.Args().First())
			},
		},
		{
			Name:    "msggen",
			Aliases: []string{"mg"},
			Usage:   "make msggen [msg path]",
			Action: func(c *cli.Context) error {
				return genMsgGen("")
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
