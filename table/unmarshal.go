package table

import "encoding/xml"

func (el *XML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	el.Attr = append(el.Attr, start.Attr...)
	return d.DecodeElement(&el.XMLInner, &start)
}

func (el *TableGroup) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	el.Attr = append(el.Attr, start.Attr...)
	return d.DecodeElement(&el.TableGroupInner, &start)
}

func (el *Entry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	el.Attr = append(el.Attr, start.Attr...)
	return d.DecodeElement(&el.EntryInner, &start)
}

func (el *Row) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	el.Attr = append(el.Attr, start.Attr...)
	return d.DecodeElement(&el.RowInner, &start)
}

func (el *SimpleXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	el.Attr = append(el.Attr, start.Attr...)
	return d.DecodeElement(&el.SimpleXMLInner, &start)
}

func (el *SimpleRow) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	el.Attr = append(el.Attr, start.Attr...)
	return d.DecodeElement(&el.SimpleRowInner, &start)
}
