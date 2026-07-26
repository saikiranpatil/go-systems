package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	filepath := "./assets/data.json"
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("error while reading file(%s): %v", filepath, err)
	}

	fileData, err := io.ReadAll(file)
	defer file.Close()

	if err != nil {
		fmt.Printf("error while reading file(%s): %v", filepath, err)
	}

	fmt.Println("JSON file read:")
	fmt.Print(string(fileData))
}
