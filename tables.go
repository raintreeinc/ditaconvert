package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/raintreeinc/ditaconvert/table"
)

/*
<table>
  <tgroup cols="2">
    <colspec colname="COLSPEC0" colwidth="121*"/>
    <colspec colname="COLSPEC1" colwidth="76*"/>
    <thead>
      <row>
        <entry colname="COLSPEC0" valign="top">Animal</entry>
        <entry colname="COLSPEC1" valign="top">Gestation</entry>
      </row>
    </thead>
    <tbody>
    <row deliveryTarget="F1 PDF">
        <entry namest="COLSPEC0" nameend="COLSPEC1">Mammals</entry>
        <entry>19-22 months</entry>
      </row>
      <row>
        <entry>Elephant (African and Asian)</entry>
        <entry>19-22 months</entry>
      </row>
      <row>
        <entry>Giraffe</entry>
        <entry>15 months</entry>
      </row>
    </tbody>
  </tgroup>
</table>
*/

func HandleTable(context *Context, dec *xml.Decoder, start xml.StartElement) error {
	var t table.XML

	err := dec.DecodeElement(&t, &start)
	if err != nil {
		return err
	}

	var errors []error

	emitStart := func(tag string, attrs ...xml.Attr) {
		context.check(context.Encoder.WriteStart(tag, attrs...))
	}
	emitEnd := func(tag string) {
		context.check(context.Encoder.WriteEnd(tag))
	}
	recurse := func(content []byte) {
		err := context.Parse(string(content))
		if err != nil {
			errors = append(errors, err)
		}
	}

	emitStart("div", t.Attr...)
	defer emitEnd("div")

	for _, group := range t.Groups {
		emitStart("table", group.Attr...)

		widths := []string{}
		for _, spec := range group.Column {
			w := spec.Width
			if strings.HasSuffix(w, "*") {
				w = strings.TrimRight(w, "*") + "%"
			}
			widths = append(widths, w)
		}

		{
			emitStart("thead")
			for i, head := range group.Head {
				if i < len(widths) {
					head.SetAttr("style", "width:"+widths[i]+";")
				}
				emitStart("th", head.Attr...)
				recurse(head.Content)
				emitEnd("th")
			}
			emitEnd("thead")
		}

		{
			emitStart("tbody")
			for _, row := range group.Rows {
				if !isWebAudience(row.GetAttr("audience"), row.GetAttr("print"), row.GetAttr("deliveryTarget")) {
					continue
				}

				emitStart("tr", row.Attr...)
				for _, entry := range row.Entries {
					if startid, endid := entry.Bounds(); startid != "" || endid != "" {
						entry.ClearBounds()
						si, ei := group.ColumnIndex(startid), group.ColumnIndex(endid)
						if si >= 0 && ei >= 0 && si < ei {
							entry.SetAttr("colspan", strconv.Itoa(ei-si+1))
						}
					}

					if morerows := entry.GetAttr("morerows"); morerows != "" {
						entry.SetAttr("morerows", "")
						rows, _ := strconv.Atoi(morerows)
						entry.SetAttr("rowspan", strconv.Itoa(rows+1))
					}

					emitStart("td", entry.Attr...)
					recurse(entry.Content)
					emitEnd("td")
				}
				emitEnd("tr")
			}
			emitEnd("tbody")
		}

		emitEnd("table")
	}

	if errors != nil {
		return fmt.Errorf("%v", errors)
	}
	return nil
}

/*
<simpletable>
  <sthead>
    <stentry>Type style</stentry>
    <stentry>Elements used</stentry>
  </sthead>
  <strow>
    <stentry>Bold</stentry>
    <stentry>b</stentry>
  </strow>
  <strow>
    <stentry>Italic</stentry>
    <stentry>i</stentry>
  </strow>
  <strow>
    <stentry>Underlined</stentry>
    <stentry>u</stentry>
  </strow>
</simpletable>
*/
func HandleSimpleTable(context *Context, dec *xml.Decoder, start xml.StartElement) error {
	var t table.SimpleXML

	err := dec.DecodeElement(&t, &start)
	if err != nil {
		return err
	}

	var errors []error

	emitStart := func(tag string, attrs ...xml.Attr) {
		context.check(context.Encoder.WriteStart(tag, attrs...))
	}
	emitEnd := func(tag string) {
		context.check(context.Encoder.WriteEnd(tag))
	}
	recurse := func(content []byte) {
		err := context.Parse(string(content))
		if err != nil {
			errors = append(errors, err)
		}
	}

	widths := strings.Split(t.GetAttr("relcolwidth"), " ")
	for i, w := range widths {
		widths[i] = strings.TrimRight(w, "*") + "%"
	}
	t.SetAttr("relcolwidth", "")

	emitStart("table", t.Attr...)
	defer emitEnd("table")

	{
		emitStart("thead")
		for i, head := range t.Head {
			if i < len(widths) {
				head.SetAttr("style", "width:"+widths[i]+";")
			}
			emitStart("th", head.Attr...)
			recurse(head.Content)
			emitEnd("th")
		}
		emitEnd("thead")
	}

	{
		emitStart("tbody")
		for _, row := range t.Rows {
			emitStart("tr", row.Attr...)
			for _, entry := range row.Entries {
				emitStart("td", entry.Attr...)
				recurse(entry.Content)
				emitEnd("td")
			}
			emitEnd("tr")
		}
		emitEnd("tbody")
	}

	if errors != nil {
		return fmt.Errorf("%v", errors)
	}
	return nil
}
