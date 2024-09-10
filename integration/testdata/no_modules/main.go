package main

import (
	_ "embed"
	"fmt"
)

//go:embed .occam-key
func main() {
	fmt.Println("Hello World!")
}
