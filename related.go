package ditaconvert

import (
	"github.com/raintreeinc/ditaconvert/dita"
	"github.com/raintreeinc/ditaconvert/html"
)

func (context *ConvertContext) RelatedLinksAsHTML() (div string) {
	topic := context.Topic
	if topic == nil || EmptyLinkSets(topic.Links) {
		return ""
	}

	contains := func(xs []string, s string) bool {
		for _, x := range xs {
			if x == s {
				return true
			}
		}
		return false
	}

	div += `<div class="related-links">`

	for _, set := range topic.Links {
		if len(set.Children) > 0 {
			if set.CollType == dita.Sequence {
				div += "<ol>"
			} else {
				div += "<ul>"
			}

			for _, link := range set.Children {
				div += "<li>" + context.LinkAsAnchor(link)
				if link.Topic.Synopsis != "" {
					div += "<p>" + link.Topic.Synopsis + "</p>"
				}
				div += "</li>"
			}

			if set.CollType == dita.Sequence {
				div += "</ol>"
			} else {
				div += "</ul>"
			}
		}

		if set.Parent != nil {
			div += "<div><b>Parent topic: </b>" + context.LinkAsAnchor(set.Parent) + "</div>"
		}
		if set.Prev != nil {
			div += "<div><b>Previous topic: </b>" + context.LinkAsAnchor(set.Prev) + "</div>"
		}
		if set.Next != nil {
			div += "<div><b>Next topic: </b>" + context.LinkAsAnchor(set.Next) + "</div>"
		}
	}

	grouped := make(map[string][]*Link)
	order := []string{"concept", "task", "reference", "tutorial", "information"}
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

	for _, kind := range order {
		links := grouped[kind]
		if len(links) == 0 {
			continue
		}

		if kind != "information" {
			kind += "s"
		}
		div += "<div><b>Related " + kind + "</b>"
		for _, link := range links {
			div += "<div>" + context.LinkAsAnchor(link) + "</div>"
		}
		div += "</div>"
	}

	div += "</div>"

	return div
}

func (context *ConvertContext) LinkAsAnchor(link *Link) string {
	if link.Scope == "external" {
		title := link.Title
		if title == "" {
			title = link.Href
		}
		return `<a href="` + html.NormalizeURL(link.Href) + `" class="external-link" target="_blank" rel="nofollow">` + title + `</a>`
	}

	if link.Topic == nil {
		return `<span style="background: #f00">` + html.EscapeString(link.Title) + `</span>`
	}

	//TODO: adjust for mapping
	// slug := string(conv.Mapping.ByTopic[link.Topic])
	ref := "/" + trimext(link.Topic.Path) + ".html"
	slug := trimext(link.Topic.Path)
	title := link.Topic.Title
	if link.Title != "" {
		title = link.Title
	}
	return `<a href="` + html.NormalizeURL(ref) + `" data-link="` + slug + `">` + html.EscapeString(title) + `</a>`
}
