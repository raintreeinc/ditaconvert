package ditaconvert

import "encoding/xml"

func getAttr(n *xml.StartElement, key string) (val string) {
	for _, attr := range n.Attr {
		if attr.Name.Local == key {
			return attr.Value
		}
	}
	return ""
}

func setAttr(n *xml.StartElement, key, val string) {
	for i := range n.Attr {
		attr := &n.Attr[i]
		if attr.Name.Local == key {
			attr.Value = val
			return
		}
	}

	n.Attr = append(n.Attr, xml.Attr{
		Name:  xml.Name{Local: key},
		Value: val,
	})
}
