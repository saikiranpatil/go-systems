package main

import (
	"fmt"
)

func sliceToChannel(nums []int) chan int {
	ch := make(chan int)

	go func() {
		for _, num := range nums {
			ch <- num
		}
		close(ch)
	}()

	return ch
}

func squared(ch <-chan int) <-chan int {
	processedChannel := make(chan int)

	go func() {
		for data := range ch {
			processedChannel <- data * data
		}
		close(processedChannel)
	}()

	return processedChannel
}

func printChannelData(ch <-chan int) {
	for data := range ch {
		fmt.Print(data, " ")
	}
}

func main() {
	// input
	nums := []int{1, 4, 6, 23}

	// stage 1: add slice to channel
	dataChannel := sliceToChannel(nums)

	// stage 2: apply square operation
	processedChannel := squared(dataChannel)

	// stage 3: print the processed channel data
	printChannelData(processedChannel)
}
