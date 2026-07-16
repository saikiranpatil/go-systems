package main

import (
	"fmt"
	"time"
)

func doWork(ch <-chan string) {
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				fmt.Println("channel closed")
				return
			}
			fmt.Printf("data: %v\n", data)
		default:
			fmt.Println("doing some work!")
		}
	}
}

func main() {
	ch := make(chan string)
	go doWork(ch)

	ch<-"data1"
	ch<-"data2"
	close(ch)
	
	
	time.Sleep(time.Second * 1)
}
