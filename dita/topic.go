package dita

import (
	"encoding/xml"
	"strings"
)

type Topic struct {
	XMLName   xml.Name
	ID        string   `xml:"id,attr"`
	Title     string   `xml:"title"`
	NavTitle  string   `xml:"titlealts>navtitle"`
	Prolog    Prolog   `xml:"prolog"`
	ShortDesc InnerXML `xml:"shortdesc"`

	RelatedLink []Link `xml:"related-links>link"`
	Elements    []Body `xml:",any"`
}

type Prolog struct {
	Keywords   Keywords    `xml:"metadata>keywords"`
	OtherMeta  []OtherMeta `xml:"metadata>othermeta"`
	ResourceID []struct {
		Name string `xml:"id,attr"`
	} `xml:"resourceid"`
}

type InnerXML struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

func (x *InnerXML) Text() (string, error) { return xmlstriptags(x.Content) }

type Body struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

type Link struct {
	Href   string `xml:"href,attr"`
	Format string `xml:"format,attr,omitempty"`
	Scope  string `xml:"scope,attr,omitempty"`
	Text   string `xml:"linktext,omitempty"`
}

type OtherMeta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type Keywords struct {
	Content string `xml:",innerxml"`
}

func (keywords *Keywords) Terms() (terms []string) {
	dec := xml.NewDecoder(strings.NewReader(keywords.Content))
	for {
		token, err := dec.Token()
		if err != nil {
			return
		}

		if _, isEnd := token.(xml.EndElement); isEnd {
			return
		}

		if start, isStart := token.(xml.StartElement); isStart {
			if start.Name.Local == "indexterm" {
				extractCombinations(dec, &terms)
			}
		}
	}

	panic("unreachable")
}

func extractCombinations(dec *xml.Decoder, terms *[]string) {
	prefix := ""
	emitted := false
	for {
		token, err := dec.Token()
		if err != nil {
			if !emitted && prefix != "" {
				*terms = append(*terms, prefix)
			}
			return
		}
		if _, isEnd := token.(xml.EndElement); isEnd {
			if !emitted && prefix != "" {
				*terms = append(*terms, prefix)
			}
			return
		}
		if start, isStart := token.(xml.StartElement); isStart {
			if start.Name.Local == "indexterm" {
				emitted = true
				if prefix != "" {
					suffix, _ := xmltext(dec)
					*terms = append(*terms, prefix+":"+suffix)
				}
			} else {
				panic("unhandled keyword:" + start.Name.Local)
			}
		}

		if data, isData := token.(xml.CharData); isData {
			prefix = string(data)
		}
	}
}
