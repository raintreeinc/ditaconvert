package ditaconvert

import (
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

// name should be a path relative FileSystem root
func (index *Index) LoadMap(name string) {
	context := MapContext{
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
