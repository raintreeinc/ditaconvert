package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	printcaption("TOPICS")

	topics := make([]*ditaconvert.Topic, 0, len(index.Topics))
	for _, topic := range index.Topics {
		topics = append(topics, topic)
	}

	sort.Sort(TopicByName(topics))

	for _, topic := range topics {
		fmt.Println()
		fmt.Println("====", topic.Title, "==== ", topic.Path)
		for _, links := range topic.Links {
			if links.Parent != nil {
				fmt.Print("\t^  ", links.Parent.Title, "\n")
			}
			if links.Prev != nil || links.Next != nil {
				fmt.Print("\t")
				if links.Prev != nil {
					fmt.Print("<- ", links.Prev.Title)
				}
				if links.Next != nil {
					if links.Prev != nil {
						fmt.Print(" - ")
					}
					fmt.Print(links.Next.Title, " ->")
				}
				fmt.Println()
			}
			if len(links.Children) > 0 {
				fmt.Print("\tv  ")
				for _, child := range links.Children {
					fmt.Print(child.Title, " ")
				}
				fmt.Println()
			}
			if len(links.Siblings) > 0 {
				fmt.Print("\t~  ")
				for _, sibling := range links.Siblings {
					fmt.Print(sibling.Title, " ")
				}
				fmt.Println()
			}
		}
	}
}

func printcaption(name string) {
	fmt.Println()
	fmt.Println("==================")
	fmt.Printf("= %-14s =\n", name)
	fmt.Println("==================")
	fmt.Println()
}

func printnav(e *ditaconvert.Entry, prefix, indent string) {
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

type TopicByName []*ditaconvert.Topic

func (a TopicByName) Len() int           { return len(a) }
func (a TopicByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TopicByName) Less(i, j int) bool { return a[i].Path < a[j].Path }
