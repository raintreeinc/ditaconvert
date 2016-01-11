package table

import "encoding/xml"

type Attributes struct{ Attr []xml.Attr }

func (el *Attributes) GetAttr(name string) string {
	for i := range el.Attr {
		attr := &el.Attr[i]
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func (el *Attributes) SetAttr(name, value string) {
	for i := range el.Attr {
		attr := &el.Attr[i]
		if attr.Name.Local == name {
			if value == "" {
				el.Attr = append(el.Attr[:i], el.Attr[i+1:]...)
			} else {
				attr.Value = value
			}
			return
		}
	}
	if value == "" {
		return
	}

	el.Attr = append(el.Attr, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}
