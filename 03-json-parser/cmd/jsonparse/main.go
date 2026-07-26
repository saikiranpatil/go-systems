package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// 1. reading from string literal
	jsonStr := `{
  "users": [
    {
      "id": 1,
      "name": "Aarav Sharma",
      "email": "aarav.sharma@example.com",
      "is_active": true,
      "age": 28,
      "skills": ["Python", "JavaScript", "SQL"],
      "address": {
        "city": "Bengaluru",
        "pincode": "560102"
      },
      "spouse": null
    },
    {
      "id": 2,
      "name": "Priya Patel",
      "email": "priya.patel@example.com",
      "is_active": false,
      "age": 34,
      "skills": ["Project Management", "Agile", "Scrum"],
      "address": {
        "city": "Mumbai",
        "pincode": "400001"
      },
      "spouse": "Rohan Patel"
    }
  ],
  "total_count": 2
}`
	fmt.Println(jsonStr)

	// 2. reading from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Echo:", line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	// 3. reading from file
	filepath := "./assets/data.json"
	file, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("error while reading file(%s): %v", filepath, err)
	}

	fmt.Println("JSON file read:")
	fmt.Print(string(file))
}
