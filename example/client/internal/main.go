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
