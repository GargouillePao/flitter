#!/bin/sh
if [ $1 = r ]
then
	go run simplegame.go -r -v
fi
if [ $1 = l ]
then
	go run simplegame.go -v -p scene@127.0.0.1:8000
fi
if [ $1 = 1 ]
then
	go run simplegame.go -v -p scene@0:0/s1@127.0.0.1:7000
fi
if [ $1 = 2 ]
then
	go run simplegame.go -v -p scene@0:0/s2@127.0.0.1:7100
fi