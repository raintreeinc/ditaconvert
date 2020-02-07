package table

import "strings"

type XML struct {
	Attributes
	XMLInner
}
type XMLInner struct {
	Groups []TableGroup `xml:"tgroup"`
}

type TableGroup struct {
	Attributes
	TableGroupInner
}
type TableGroupInner struct {
	Column []ColSpec `xml:"colspec"`
	Head   []Entry   `xml:"thead>row>entry"`
	Rows   []Row     `xml:"tbody>row"`
}

func (table *TableGroup) ColumnIndex(name string) int {
	for i, col := range table.Column {
		if strings.EqualFold(col.Name, name) {
			return i
		}
	}
	return -1
}

type ColSpec struct {
	Name  string `xml:"colname,attr"`
	Num   int    `xml:"colnum,attr"`
	Width string `xml:"colwidth,attr"`
}

type Entry struct {
	Attributes
	EntryInner
}
type EntryInner struct {
	Content []byte `xml:",innerxml"`
}

func (e *Entry) Bounds() (start, end string) { return e.GetAttr("namest"), e.GetAttr("nameend") }

func (e *Entry) ClearBounds() {
	e.SetAttr("namest", "")
	e.SetAttr("nameend", "")
}

type Row struct {
	Attributes
	RowInner
}
type RowInner struct {
	Entries []Entry `xml:"entry"`
}

type SimpleXML struct {
	Attributes
	SimpleXMLInner
}
type SimpleXMLInner struct {
	Head []Entry     `xml:"sthead>stentry"`
	Rows []SimpleRow `xml:"strow"`
}

type SimpleRow struct {
	Attributes
	SimpleRowInner
}
type SimpleRowInner struct {
	Entries []Entry `xml:"stentry"`
}

type SettingsXML struct {
	Attributes
	SettingsXMLInner
}
type SettingsXMLInner struct {
	Rows []SettingRow `xml:"setting"`
}

type SettingRow struct {
	Attributes
	SettingRowInner
}
type SettingRowInner struct {
	Name	SettingName `xml:"settingname"`
	Default	string 		`xml:"settingdefault"`
	Desc	Entry 		`xml:"settingdesc"`
	Example	Entry 		`xml:"settingsample"`
	Levels	string 		`xml:"settinglevels"`
}
type SettingName struct {
	Attributes
	Text string `xml:",innerxml"`
}