package main

import (
	"time"
)

func main() {}

//export add
func add(a uint32, b uint32) uint64 {
	return uint64(a + b)
}

//export microsleep
func microsleep(a uint32, b uint32) uint64 {
	time.Sleep(time.Microsecond)
	return uint64(a + b)
}

var _ = add
var _ = microsleep
