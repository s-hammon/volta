package main

import (
	"fmt"
	"os"
)

var msgPath string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: extract <path>")
		os.Exit(1)
	}
	msgPath = os.Args[1]
	fmt.Printf("%s\n", msgPath)
}
