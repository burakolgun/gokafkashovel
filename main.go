package main

import (
	"sync"

	"github.com/burakolgun/gokafkashovel/startup"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go startup.Start()
	wg.Wait()
}
