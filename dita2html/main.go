package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/raintreeinc/ditaconvert"
	"github.com/raintreeinc/ditaconvert/html"
)

func main() {
	fmt.Println(strings.Repeat("\n", 10))
	fmt.Println("<< START >>")
	start := time.Now()
	defer func() {
		fmt.Printf("<< DONE %v >>\n", time.Since(start))
	}()

	root := os.Args[1]
	index := ditaconvert.NewIndex(ditaconvert.Dir(filepath.Dir(root)))
	index.LoadMap(filepath.ToSlash(filepath.Base(root)))

	for _, err := range index.Errors {
		fmt.Println(err)
	}

	WriteTOC(index.Nav, filepath.FromSlash("output~/_toc.html"))
	for _, topic := range index.Topics {
		filename := path.Join("output~", ReplaceExt(topic.Path, ".html"))
		WriteTopic(index, topic, filepath.FromSlash(filename))
	}
}

func WriteTOC(entry *ditaconvert.Entry, filename string) {
	os.MkdirAll(filepath.Dir(filename), 0755)
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	fmt.Fprintf(out, `<link rel="stylesheet" href="/style.css">`)
	fmt.Fprintf(out, `<base target="dynamic">`)

	var PrintEntry func(entry *ditaconvert.Entry)
	PrintEntry = func(entry *ditaconvert.Entry) {
		if !entry.TOC {
			return
		}
		if entry.Topic == nil {
			fmt.Fprintf(out, `<li>%s`, html.EscapeString(entry.Title))
		} else {
			newpath := ReplaceExt(entry.Topic.Path, ".html")
			fmt.Fprintf(out, `<li><a href="/%s">%s</a>`, html.NormalizeURL(newpath), entry.Title)
		}

		if len(entry.Children) > 0 {
			fmt.Fprintf(out, `<ul>`)
			for _, child := range entry.Children {
				PrintEntry(child)
			}
			fmt.Fprintf(out, `</ul>`)
		}
		fmt.Fprintf(out, "</li>")
	}
	PrintEntry(entry)
}

func WriteTopic(index *ditaconvert.Index, topic *ditaconvert.Topic, filename string) {
	os.MkdirAll(filepath.Dir(filename), 0755)
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	conversion := ditaconvert.NewConversion(index, topic)
	if err := conversion.Run(); err != nil {
		fmt.Printf("[%s] %s: %s\n", topic.Path, topic.Title, err)
		return
	}

	fmt.Fprintf(out, `<link rel="stylesheet" href="/style.css">`)
	fmt.Fprintf(out, `<body>`)
	fmt.Fprintf(out, `<h3>`+html.EscapeString(topic.Title)+`</h3>`)
	fmt.Fprintf(out, `<div>`)
	fmt.Fprintf(out, conversion.Output.String())
	fmt.Fprintf(out, `</div>`)
	fmt.Fprintf(out, `</body>`)
}

func ReplaceExt(name string, newext string) string {
	return name[:len(name)-len(path.Ext(name))] + newext
}
