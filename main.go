package main

import (
	"fmt"
	"github.com/burakolgun/gokafkashovel/startup"
	"sync"
)

func init() {
	go startup.Start()
}

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
	}()

	for true {
		wg := sync.WaitGroup{}
		wg.Add(1)

		wg.Wait()
	}
}
