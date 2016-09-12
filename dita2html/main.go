package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/raintreeinc/ditaconvert"
	"github.com/raintreeinc/ditaconvert/dita"
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

	fmt.Fprint(out, `<link rel="stylesheet" href="/style.css">`)
	fmt.Fprint(out, `<base target="dynamic">`)

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
			fmt.Fprint(out, `<ul>`)
			for _, child := range entry.Children {
				PrintEntry(child)
			}
			fmt.Fprint(out, `</ul>`)
		}
		fmt.Fprint(out, "</li>")
	}
	PrintEntry(entry)
}

func WriteTopic(index *ditaconvert.Index, topic *ditaconvert.Topic, filename string) {
	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	out := bufio.NewWriter(file)
	defer out.Flush()

	conversion := ditaconvert.NewConversion(index, topic)
	if err := conversion.Run(); err != nil || len(conversion.Errors) > 0 {
		fmt.Printf("[%s] %s: %v\n", topic.Path, topic.Title, err)
		for _, err := range conversion.Errors {
			fmt.Printf("\t%v\n", err)
		}
		if err != nil {
			return
		}
	}

	fmt.Fprint(out, "<!--INGREDIENTS:\n")
	fmt.Fprint(out, "Keywords=")
	for i, key := range topic.Original.Prolog.Keywords.Terms() {
		if i > 0 {
			fmt.Fprint(out, ",")
		}
		fmt.Fprint(out, key)
	}
	for _, meta := range topic.Original.Prolog.OtherMeta {
		fmt.Fprintf(out, "%s=%s\n", html.EscapeString(meta.Name), html.EscapeString(meta.Content))
	}
	fmt.Fprint(out, "-->\n")
	fmt.Fprint(out, `<body id="`+topic.Original.ID+`">`)
	fmt.Fprint(out, `<h3>`+html.EscapeString(topic.Title)+`</h3>`)
	fmt.Fprint(out, `<div>`)
	fmt.Fprint(out, conversion.Output.String())
	fmt.Fprint(out, `</div>`)
	fmt.Fprint(out, RelatedLinksAsHTML(conversion))
	fmt.Fprint(out, `</body>`)
}

func ReplaceExt(name string, newext string) string {
	return name[:len(name)-len(path.Ext(name))] + newext
}

func RelatedLinksAsHTML(context *ditaconvert.Context) (div string) {
	topic := context.Topic
	if topic == nil || ditaconvert.EmptyLinkSets(topic.Links) {
		return "<div></div>"
	}

	contains := func(xs []string, s string) bool {
		for _, x := range xs {
			if x == s {
				return true
			}
		}
		return false
	}

	div += `<div>`

	var hasFamilyLinks bool
	for _, set := range topic.Links {
		hasFamilyLinks = hasFamilyLinks || set.Parent != nil
		hasFamilyLinks = hasFamilyLinks || set.Prev != nil
		hasFamilyLinks = hasFamilyLinks || set.Next != nil

		if len(set.Children) == 0 {
			continue
		}

		if set.CollType == dita.Sequence {
			div += `<ol class="ullinks">`
		} else {
			div += `<ul class="ullinks">`
		}

		for _, link := range set.Children {
			div += `<li class="ulchildlink">` + LinkAsAnchorNoTitle(context, link)
			if link.Topic.Synopsis != "" {
				div += `<p>` + link.Topic.Synopsis + `</p>`
			}
			div += `</li>`
		}

		if set.CollType == dita.Sequence {
			div += `</ol>`
		} else {
			div += `</ul>`
		}
	}

	if hasFamilyLinks {
		div += `<div class="familylinks">`
		for _, set := range topic.Links {
			if set.Parent == nil && set.Prev == nil && set.Next == nil {
				continue
			}
			if set.Parent != nil {
				div += `<div class="parentlink"><strong>Parent topic: </strong>` + LinkAsAnchor(context, set.Parent) + `</div>`
			}
			if set.Prev != nil {
				div += `<div class="previouslink"><strong>Previous topic: </strong>` + LinkAsAnchor(context, set.Prev) + `</div>`
			}
			if set.Next != nil {
				div += `<div class="nextlink"><strong>Next topic: </strong>` + LinkAsAnchor(context, set.Next) + `</div>`
			}
		}
		div += `</div>`
	}

	grouped := make(map[string][]*ditaconvert.Link)
	order := []string{"tutorial", "concept", "task", "reference", "information"}
	for _, set := range topic.Links {
		for _, link := range set.Siblings {
			kind := ""
			if link.Topic != nil {
				kind = link.Topic.Original.XMLName.Local
			}
			if link.Type != "" {
				kind = link.Type
			}
			if !contains(order, kind) {
				kind = "information"
			}

			grouped[kind] = append(grouped[kind], link)
		}
	}

	// for _, links := range grouped {
	// 	ditaconvert.SortLinks(links)
	// }

	for _, kind := range order {
		links := grouped[kind]
		if len(links) == 0 {
			continue
		}

		if kind != "information" {
			class := kindclass[kind]
			if len(links) > 1 {
				kind += "s"
			}
			div += `<div class="relinfo ` + class + `"><strong>Related ` + kind + `</strong>`
		} else {
			div += `<div class="relinfo"><strong>Related information</strong>`
		}
		for _, link := range links {
			div += "<div>" + LinkAsAnchor(context, link) + "</div>"
		}
		div += "</div>"
	}
	div += "</div>"

	return div
}

var kindclass = map[string]string{
	"tutorial":  "relconcepts", //"reltutorials",
	"reference": "relref",
	"concept":   "relconcepts",
	"task":      "reltasks",
}

func LinkAsAnchorNoTitle(context *ditaconvert.Context, link *ditaconvert.Link) string {
	title := html.EscapeCharData(link.FinalTitle())

	if link.Scope == "external" {
		return `<a href="` + html.NormalizeURL(link.Href) + `" class="external-link" target="_blank" rel="nofollow">` + title + `</a>`
	}

	if link.Topic == nil {
		return `<span style="background: #f00">` + title + `</span>`
	}

	reldir := PathRel(
		path.Dir(context.Topic.Path),
		path.Dir(link.Topic.Path),
	)
	ref := path.Join(reldir, trimext(path.Base(link.Topic.Path))+".html")

	return `<a href="` + html.NormalizeURL(ref) + `">` + title + `</a>`
}

func LinkAsAnchor(context *ditaconvert.Context, link *ditaconvert.Link) string {
	title := html.EscapeCharData(link.FinalTitle())
	if link.Scope == "external" {
		return `<a href="` + html.NormalizeURL(link.Href) + `" class="external-link" target="_blank" rel="nofollow">` + title + `</a>`
	}

	if link.Topic == nil {
		return `<span style="background: #f00">` + title + `</span>`
	}

	reldir := PathRel(
		path.Dir(context.Topic.Path),
		path.Dir(link.Topic.Path),
	)
	ref := path.Join(reldir, trimext(path.Base(link.Topic.Path))+".html")

	desc := link.Topic.Synopsis
	if desc == "" {
		return `<a href="` + html.NormalizeURL(ref) + `">` + title + `</a>`
	}
	return `<a href="` + html.NormalizeURL(ref) + `" title="` + html.EscapeAttribute(desc) + `">` + title + `</a>`
}

func trimext(name string) string { return name[0 : len(name)-len(filepath.Ext(name))] }

func PathRel(basepath, targpath string) string {
	relpath, err := filepath.Rel(
		filepath.FromSlash(basepath),
		filepath.FromSlash(targpath),
	)
	if err != nil {
		return targpath
	}
	return filepath.ToSlash(relpath)
}
