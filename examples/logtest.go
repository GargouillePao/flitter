package main

import (
	"fmt"
	"path"
	"runtime"
)

func log() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println(path.Base(file), line)
}
func main() {
	fmt.Println(runtime.Caller(0))
	log()
	fmt.Println("OK")
}
