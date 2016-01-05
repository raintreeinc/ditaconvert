package ditaconvert

import (
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/dita"
)

type Link struct {
	Topic *Topic
	Title string
	Type  string
	Scope string
	Href  string

	Selector string
}

type Links struct {
	CollType   dita.CollectionType
	Parent     *Link
	Prev, Next *Link
	Siblings   []*Link
	Children   []*Link
}

func (e *Entry) asLink() *Link {
	return &Link{
		Topic: e.Topic,
		Title: e.Title,
		Type:  e.Type,
		Scope: "",
		Href:  "",
	}
}

func (links *Links) IsEmpty() bool {
	return links.Parent == nil &&
		links.Prev == nil && links.Next == nil &&
		len(links.Siblings) == 0 &&
		len(links.Children) == 0
}

func EmptyLinkSets(links []Links) bool {
	for _, set := range links {
		if !set.IsEmpty() {
			return false
		}
	}
	return true
}

func (context MapContext) AddFamilyLinks(entries []*Entry) {
	linkable := make([]*Entry, 0, len(entries))
	for _, e := range entries {
		if e.Topic != nil {
			linkable = append(linkable, e)
		}
	}

	if context.Entry.Topic != nil && context.Entry.Linking.CanLinkTo() {
		for _, a := range linkable {
			if a.Linking.CanLinkFrom() {
				a.Topic.Links = append(a.Topic.Links, Links{Parent: context.Entry.asLink()})
			}
		}
	}

	if context.Entry.Topic != nil && context.Entry.Linking.CanLinkFrom() {
		links := Links{CollType: context.CollType}
		for _, a := range linkable {
			if a.Linking.CanLinkTo() {
				links.Children = append(links.Children, a.asLink())
			}
		}

		if len(links.Children) > 0 {
			context.Entry.Topic.Links = append(context.Entry.Topic.Links, links)
		}
	}

	switch context.CollType {
	case dita.Family:
		for _, a := range linkable {
			links := Links{}
			if !a.Linking.CanLinkFrom() {
				continue
			}

			for _, b := range linkable {
				if a != b && b.Linking.CanLinkTo() {
					links.Siblings = append(links.Siblings, b.asLink())
				}
			}

			if len(links.Siblings) > 0 {
				a.Topic.Links = append(a.Topic.Links, links)
			}
		}
	case dita.Sequence:
		for i, a := range linkable {
			if !a.Linking.CanLinkFrom() {
				continue
			}

			links := Links{}
			if i-1 >= 0 {
				prev := linkable[i-1]
				if prev.Linking.CanLinkTo() {
					links.Prev = prev.asLink()
				}
			}
			if i+1 < len(linkable) {
				next := linkable[i+1]
				if next.Linking.CanLinkTo() {
					links.Next = next.asLink()
				}
			}
			if links.Prev != nil || links.Next != nil {
				a.Topic.Links = append(a.Topic.Links, links)
			}
		}
	}
}

func InterLinkEntries(A, B []*Entry) {
	for _, a := range A {
		if a.Topic == nil || !a.Linking.CanLinkFrom() {
			continue
		}

		links := Links{}
		for _, b := range B {
			if b.Topic == nil {
				continue
			}

			if b.Linking.CanLinkTo() {
				links.Siblings = append(links.Siblings, b.asLink())
			}
		}

		if len(links.Siblings) > 0 {
			a.Topic.Links = append(a.Topic.Links, links)
		}
	}
}

func (context MapContext) CreateRelatedLinks(a *Entry) {
	if a.Topic == nil || a.Topic.RelatedLinksCreated || a.Topic.Original == nil {
		return
	}
	a.Topic.RelatedLinksCreated = true

	context.Dir = path.Dir(a.Topic.Path)

	links := Links{}
	for _, rellink := range a.Topic.Original.RelatedLink {
		if rellink.Scope == "external" || strings.HasPrefix(rellink.Href, "http:") {
			links.Siblings = append(links.Siblings, &Link{
				Topic: nil,
				Title: rellink.Text,
				Scope: "external",
				Href:  rellink.Href,
			})
			continue
		}

		var topic *Topic
		if rellink.Href != "" {
			name, selector := SplitLink(rellink.Href)
			topic = context.LoadTopic(name)
			if topic != nil {
				links.Siblings = append(links.Siblings, &Link{
					Topic:    topic,
					Title:    rellink.Text,
					Selector: selector,
				})
			}
			continue
		}
	}

	if len(links.Siblings) > 0 {
		a.Topic.Links = append(a.Topic.Links, links)
	}
}

func SplitLink(link string) (name, selector string) {
	tokens := strings.SplitN(link, "#", 2)
	if len(tokens) == 2 {
		return tokens[0], tokens[1]
	}
	return link, ""
}

/*

func (conv *convert) addRelatedLinks() {
	if emptylinks(conv.Topic.Links) {
		return
	}

	text := "<div class=\"dita-related-links\">"

	for _, set := range conv.Topic.Links {
		if len(set.Children) > 0 {
			if set.CollType == dita.Sequence {
				text += "<ol>"
			} else {
				text += "<ul>"
			}

			for _, link := range set.Children {
				text += "<li>" + conv.linkAsHTML(link)
				if link.Topic.Synopsis != "" {
					text += "<p>" + link.Topic.Synopsis + "</p>"
				}
				text += "</li>"
			}
			text += "</ol>"

			if set.CollType == dita.Sequence {
				text += "</ol>"
			} else {
				text += "</ul>"
			}
		}

		if set.Parent != nil {
			text += "<div><b>Parent topic: </b>" + conv.linkAsHTML(set.Parent) + "</div>"
		}
		if set.Prev != nil {
			text += "<div><b>Previous topic: </b>" + conv.linkAsHTML(set.Prev) + "</div>"
		}
		if set.Next != nil {
			text += "<div><b>Next topic: </b>" + conv.linkAsHTML(set.Next) + "</div>"
		}
	}

	grouped := make(map[string][]*Link)
	order := []string{"concept", "task", "reference", "tutorial", "information"}
	for _, set := range conv.Topic.Links {
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
		text += "<div><b>Related " + kind + "</b>"
		for _, link := range links {
			text += "<div>" + conv.linkAsHTML(link) + "</div>"
		}
		text += "</div>"
	}

	text += "</div>"
	conv.Page.Story.Append(kb.HTML(text))
}

func (conv *convert) linkAsHTML(link *Link) string {
	if link.Scope == "external" {
		title := link.Title
		if title == "" {
			title = link.Href
		}
		return "<a href=\"" + link.Href + "\" class=\"external-link\" target=\"_blank\" rel=\"nofollow\">" + title + "</a>"
	}

	if link.Topic == nil {
		return ""
	}

	slug := string(conv.Mapping.ByTopic[link.Topic])
	title := link.Topic.Title
	if link.Title != "" {
		title = link.Title
	}
	return "<a href=\"" + slug + "\" data-link=\"" + slug + "\">" + title + "</a>"
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
*/
