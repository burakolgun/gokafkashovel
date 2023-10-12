package main

import (
	"fmt"
	"sync"

	"github.com/burakolgun/gokafkashovel/startup"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go startup.Start()
	wg.Wait()
}
	