package ditaconvert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/raintreeinc/ditaconvert/html"
)

func getAttr(n *xml.StartElement, key string) (val string) {
	for _, attr := range n.Attr {
		if attr.Name.Local == key {
			return attr.Value
		}
	}
	return ""
}

type attrByName []xml.Attr

func (xs attrByName) Len() int           { return len(xs) }
func (xs attrByName) Swap(i, j int)      { xs[i], xs[j] = xs[j], xs[i] }
func (xs attrByName) Less(i, j int) bool { return xs[i].Name.Local < xs[j].Name.Local }

func setAttr(n *xml.StartElement, key, val string) {
	n.Attr = append([]xml.Attr{}, n.Attr...)

	for i := range n.Attr {
		attr := &n.Attr[i]
		if attr.Name.Local == key {
			if val == "" {
				n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			} else {
				attr.Value = val
			}
			return
		}
	}

	if val == "" {
		return
	}

	n.Attr = append(n.Attr, xml.Attr{
		Name:  xml.Name{Local: key},
		Value: val,
	})
	sort.Sort(attrByName(n.Attr))
}

func ExtractTitle(data []byte, selector string) (title string, err error) {
	dec := xml.NewDecoder(bytes.NewReader(data))

	start, err := WalkNodePath(dec, selector)
	if err != nil {
		return "", err
	}

	if start.Name.Local == "title" {
		return html.XMLText(dec)
	}

	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return "", nil
			}
			return "", err
		}
		if _, done := token.(xml.EndElement); done {
			return "", nil
		}
		if start, isStart := token.(xml.StartElement); isStart {
			if start.Name.Local == "title" {
				return html.XMLText(dec)
			}
			dec.Skip()
		}
	}

	panic("unreachable")
}

func WalkNodePath(dec *xml.Decoder, selector string) (xml.StartElement, error) {
	if selector == "" {
		return xml.StartElement{}, errors.New("invalid path")
	}

	splitfront := func(p string) (front string, tail string) {
		i := strings.IndexRune(p, '/')
		if i >= 0 {
			return p[:i], p[i+1:]
		}
		return p, ""
	}

	var nextid string

	nextid, selector = splitfront(selector)
	for {
		token, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, err
		}

		start, isStart := token.(xml.StartElement)
		if isStart && strings.EqualFold(nextid, getAttr(&start, "id")) {
			nextid, selector = splitfront(selector)
			if nextid == "" {
				return start, nil
			}
		}
	}

	panic("unreachable")
}
