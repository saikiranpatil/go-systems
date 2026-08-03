package main

import (
	"fmt"

	"github.com/saikiranpatil/json-parser/internal/lexer"
)

func main() {
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

	l := lexer.NewLexer(jsonStr)
	for t := l.NextToken(); t.Type != lexer.EOF; t = l.NextToken() {
		fmt.Print(t.Literal)
	}
}
