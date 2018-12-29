package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "fltool"
	app.Usage = "tool for flproject"
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "init",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "new [app name]",
			Action: func(c *cli.Context) error {
				log.Println(c.Args()[0])
				return nil
			},
		},
		{
			Name:    "msggen",
			Aliases: []string{"mg"},
			Usage:   "msggen [msg path]",
			Action: func(c *cli.Context) error {
				return GenMsgs(c.Args()[0])
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
