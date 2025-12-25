package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	original := "Hello, OTUS!"
	reversed := reverse.String(original)
	fmt.Println(reversed)
}
