package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raintreeinc/ditaconvert"
)

func main() {
	fmt.Println(strings.Repeat("\n", 10))
	fmt.Println("<< START >>")

	root := os.Args[1]
	index := ditaconvert.NewIndex(ditaconvert.Dir(filepath.Dir(root)))
	index.LoadMap(filepath.ToSlash(filepath.Base(root)))

	for _, err := range index.Errors {
		fmt.Println(err)
	}

	fmt.Println("<< DONE >>")
}
