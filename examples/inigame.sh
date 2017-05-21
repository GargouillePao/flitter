#!/bin/sh
if [ $1 = r ] 
then
	go run simplegame.go -v
fi
if [ $1 = 1 ]
then
	if [ $# = 1 ]
		then
		go run simplegame.go -v -p scene@127.0.0.1:8000
	fi
	if [ $# = 2 ]
		then
		go run simplegame.go -v -p scene@0:0/s1@127.0.0.1:700$2
	fi
fi
if [ $1 = 2 ]
then
	if [ $# = 1 ]
		then
		go run simplegame.go -v -p login@127.0.0.1:8100
	fi
	if [ $# = 2 ]
		then
		go run simplegame.go -v -p login@0:0/s1@127.0.0.1:710$2
	fi
fi