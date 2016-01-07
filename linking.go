package ditaconvert

import (
	"path"
	"sort"
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

func (link *Link) FinalTitle() string {
	if link.Title != "" {
		return link.Title
	}
	if link.Topic != nil {
		return link.Topic.Title
	}
	if link.Scope == "external" {
		return link.Href
	}
	return ""
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

type linksByTitle []*Link

func (xs linksByTitle) Len() int           { return len(xs) }
func (xs linksByTitle) Swap(i, j int)      { xs[i], xs[j] = xs[j], xs[i] }
func (xs linksByTitle) Less(i, j int) bool { return xs[i].FinalTitle() < xs[j].FinalTitle() }

func SortLinks(links []*Link) {
	sort.Sort(linksByTitle(links))
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
