package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/dita"
)

type MapContext struct {
	*Index
	Dir      string
	Entry    *Entry
	CollType dita.CollectionType
	Linking  dita.Linking
	TOC      bool
}

func (context MapContext) LoadMap(filename string) []*Entry {
	name := path.Join(context.Dir, filename)
	cname := CanonicalPath(name)
	if m, loaded := context.Maps[cname]; loaded {
		return m.Entries
	}

	data, _, err := context.ReadFile(name)
	if err != nil {
		context.check(fmt.Errorf("failed to read map %s: %v", name, err))
		return nil
	}

	m := &Map{Node: &dita.MapNode{}}
	if err := xml.Unmarshal(data, m.Node); err != nil {
		context.check(fmt.Errorf("failed to unmarshal map %s: %v", name, err))
		return nil
	}
	context.Maps[cname] = m

	context.Dir = path.Dir(name)
	m.Entries = context.ProcessNode(m.Node)

	return m.Entries
}

func (context MapContext) LoadTopic(filename string) *Topic {
	name := path.Join(context.Dir, filename)
	cname := CanonicalPath(name)
	if topic, loaded := context.Topics[cname]; loaded {
		return topic
	}

	data, modified, err := context.ReadFile(name)
	if err != nil {
		context.check(fmt.Errorf("failed to read topic %s: %v", name, err))
		return &Topic{
			Path:  name,
			Title: trimext(path.Base(filename)),
		}
	}

	topic := &dita.Topic{}
	if err := xml.Unmarshal(data, topic); err != nil {
		context.check(fmt.Errorf("failed to unmarshal topic %s: %v", name, err))
		return &Topic{
			Path:  name,
			Title: trimext(path.Base(filename)),
		}
	}

	top := &Topic{
		Path:       name,
		Title:      topic.NavTitle,
		ShortTitle: topic.Title,

		Raw:      data,
		Modified: modified,
		Original: topic,
	}
	top.Synopsis, _ = topic.ShortDesc.Text()
	if top.Title == "" {
		top.Title = topic.Title
	}

	context.Topics[cname] = top
	return top
}

func (context MapContext) ProcessNode(node *dita.MapNode) []*Entry {
	if node == nil {
		panic("invalid node passed as argument")
	}

	if node.Format != "" || !isWebAudience(node.Audience, node.Print, node.DeliveryTarget) {
		return nil
	}

	href, err := url.QueryUnescape(node.Href)
	if !context.check(err) {
		node.Href = href
	}

	if node.CollType != "" {
		context.CollType = node.CollType
	} else {
		context.CollType = dita.Unordered
	}

	if node.Linking != "" {
		context.Linking = node.Linking
	}

	if node.XMLName.Local == "topicgroup" || node.XMLName.Local == "map" {
		var entries []*Entry
		context.TOC = isChildTOC(context.TOC, node.TOC)
		for _, child := range node.Children {
			entries = append(entries, context.ProcessNode(child)...)
		}
		context.Entry = &Entry{}
		context.AddFamilyLinks(entries)
		return entries
	}

	if node.XMLName.Local == "mapref" {
		context.TOC = isChildTOC(context.TOC, node.TOC)
		return context.LoadMap(node.Href)
	}

	if node.XMLName.Local == "reltable" {
		for _, row := range node.Children {
			if row.XMLName.Local == "relrow" {
				context.ProcessRelRow(row)
			}
		}
		return nil
	}

	entry := &Entry{
		Title:     node.NavTitle,
		LockTitle: node.LockTitle == "yes",
		Type:      node.Type,
		TOC:       isChildTOC(context.TOC, node.TOC),
	}
	if entry.Title == "" {
		entry.Title = node.Title
	}

	if node.Href != "" {
		entry.Topic = context.LoadTopic(node.Href)
		if entry.Title == "" && entry.Topic != nil && !entry.LockTitle {
			entry.Title = entry.Topic.Title
		}
	}

	context.Entry = entry
	context.TOC = entry.TOC

	children := []*Entry{}
	for _, child := range node.Children {
		children = append(children, context.ProcessNode(child)...)
	}
	entry.Children = append(entry.Children, children...)

	context.AddFamilyLinks(children)
	context.CreateRelatedLinks(entry)

	return []*Entry{entry}
}

func (context MapContext) ProcessRelRow(node *dita.MapNode) {
	if !isWebAudience(node.Audience, node.Print, node.DeliveryTarget) {
		return
	}

	var entrysets [][]*Entry
	for _, cell := range node.Children {
		if cell.XMLName.Local != "relcell" {
			continue
		}

		var entries []*Entry
		for _, child := range cell.Children {
			entries = append(entries, context.ProcessNode(child)...)
		}

		entrysets = append(entrysets, entries)
	}

	for i, a := range entrysets {
		for j, b := range entrysets {
			if i != j {
				InterLinkEntries(a, b)
			}
		}
	}
}

func isWebAudience(audience string, print, deliveryTarget string) bool {
	return !(audience == "html" ||
		audience == "print" ||
		print == "printonly" ||
		(deliveryTarget != "" && !strings.Contains(" "+deliveryTarget+" ", " KB ")))
}

func isChildTOC(parenttoc bool, childtoc string) bool {
	if childtoc == "" {
		return parenttoc
	}
	return childtoc == "yes"
}
