package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/raintreeinc/ditaconvert/dita"
)

type Index struct {
	FileSystem

	// cpath(path) --> topic
	Topics map[string]*Topic
	Maps   map[string]*Map

	Nav *Entry

	Errors []error
}

type Topic struct {
	Path string

	Title      string
	ShortTitle string
	Synopsis   string

	Links               []Links
	RelatedLinksCreated bool

	Raw      []byte
	Modified time.Time
	Original *dita.Topic
}

type Map struct {
	Path    string
	Entries []*Entry
	Node    *dita.MapNode
}

type Entry struct {
	Title string
	Type  string

	CollType  dita.CollectionType
	Linking   dita.Linking
	TOC       bool
	LockTitle bool

	Children []*Entry
	Topic    *Topic
}

func NewIndex(fs FileSystem) *Index {
	return &Index{
		FileSystem: fs,

		Nav: &Entry{
			Title:   "Navigation",
			Linking: dita.NormalLinking,
			TOC:     true,
		},

		Maps:   make(map[string]*Map),
		Topics: make(map[string]*Topic),
	}
}

func (index *Index) check(err error) bool {
	if err != nil {
		index.Errors = append(index.Errors, err)
		return true
	}
	return false
}

type LoadContext struct {
	*Index
	Dir      string
	Entry    *Entry
	CollType dita.CollectionType
	Linking  dita.Linking
	TOC      bool
}

// name should be a path relative FileSystem root
func (index *Index) LoadMap(name string) {
	context := LoadContext{
		Index:    index,
		Dir:      path.Dir(name),
		Entry:    index.Nav,
		CollType: dita.Unordered,
		Linking:  dita.NormalLinking,
		TOC:      true,
	}

	entries := context.LoadMap(path.Base(name))
	index.Nav.Children = append(index.Nav.Children, entries...)
}

func (context LoadContext) LoadMap(filename string) []*Entry {
	name := path.Join(context.Dir, filename)
	cname := canonicalName(name)
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
	m.Entries = context.ProcessMapNode(m.Node)

	return m.Entries
}

func (context LoadContext) LoadTopic(filename string) *Topic {
	name := path.Join(context.Dir, filename)
	cname := canonicalName(name)
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

	synopsis, _ := topic.ShortDesc.Text()
	title := topic.NavTitle
	if title == "" {
		title = topic.Title
	}

	return &Topic{
		Path:       name,
		Title:      title,
		ShortTitle: topic.Title,
		Synopsis:   synopsis,

		Raw:      data,
		Modified: modified,
		Original: topic,
	}
}

func (context LoadContext) ProcessMapNode(node *dita.MapNode) []*Entry {
	if node == nil {
		panic("invalid node passed as argument")
	}

	if node.Format != "" || !isWebAudience(node.Audience, node.Print) {
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
			entries = append(entries, context.ProcessMapNode(child)...)
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
		children = append(children, context.ProcessMapNode(child)...)
	}
	entry.Children = append(entry.Children, children...)

	context.AddFamilyLinks(children)
	context.CreateRelatedLinks(entry)

	return []*Entry{entry}
}

func (context LoadContext) ProcessRelRow(node *dita.MapNode) {
	if !isWebAudience(node.Audience, node.Print) {
		return
	}

	var entrysets [][]*Entry
	for _, cell := range node.Children {
		if cell.XMLName.Local != "relcell" {
			continue
		}

		var entries []*Entry
		for _, child := range cell.Children {
			entries = append(entries, context.ProcessMapNode(child)...)
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

func isWebAudience(audience string, print string) bool {
	return audience != "html" && audience != "print" && print != "printonly"
}

func isChildTOC(parenttoc bool, childtoc string) bool {
	if childtoc == "" {
		return parenttoc
	}
	return childtoc == "yes"
}
