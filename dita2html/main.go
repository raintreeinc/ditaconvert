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

	printcaption("NAV")
	printnav(index.Nav, "", "    ")

	//printcaption("TOPICS")

	fmt.Println("<< DONE >>")
}

func printcaption(name string) {
	fmt.Println()
	fmt.Println("==================")
	fmt.Printf("= %-14s =\n", name)
	fmt.Println("==================")
	fmt.Println()
}

func printnav(e *ditaconvert.Entry, prefix, indent string) {
	if !e.TOC {
		return
	}

	link := ""
	if e.Topic != nil {
		link = ">"
	}
	if len(e.Children) == 0 {
		fmt.Printf("%s- %s %s\n", prefix, e.Title, link)
		return
	}

	fmt.Printf("%s+ %s %s\n", prefix, e.Title, link)
	for _, child := range e.Children {
		printnav(child, prefix+indent, indent)
	}
}
