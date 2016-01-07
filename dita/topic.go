package dita

import "encoding/xml"

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
	Keywords   []string `xml:"metadata>keywords>indexterm"`
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
