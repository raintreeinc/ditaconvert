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
  <strow deliveryTarget="F1 PDF">
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
			if !isWebAudience(row.GetAttr("audience"), row.GetAttr("print"), row.GetAttr("deliveryTarget")) {
				continue
			}

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

/*
<settings>
  <settinghead>
    <settingnamehd>Setting name</settingnamehd>
    <settingdefaulthd>Default value</settingdefaulthd>
    <settingdeschd>Description</settingdeschd>
    <settingsamplehd>Example</settingsamplehd>
    <settinglevelshd>Levels where it can be defined</settinglevelshd>
  </settinghead>
  <setting id="ConfRefreshInterval">
    <settingname>ConfRefreshInterval</settingname>
    <settingdefault>30</settingdefault>
    <settingdesc>Interval in minutes, during which the agent does not have to <xref href="../CheckPluginChanges.dita">update the plugin configuration</xref>. After the interval, the agent compares the file system state against the database state to see if the EMR plugin configuration is up to date and runs the update, if necessary.</settingdesc>
    <settingsample>
      <codeblock>[Plugins]
ConfRefreshInterval=60</codeblock>
    </settingsample>
    <settinglevels>user, database, scriptpath, homedir</settinglevels>
  </setting>
  <setting id="SuppressDuplicateActivePluginsWarning">
    <settingname>SuppressDuplicateActivePluginsWarning</settingname>
    <settingdefault>N</settingdefault>
    <settingdesc>Option to switch off the duplicate active plugins warning. This warning message is displayed if there are 2 or more active plugins that use RTM files with matching names. The message is displayed after logging in and selecting a database, and it is only displayed to the Raintree Tech Support user "GO".</settingdesc>
    <settingsample>
      <codeblock>[Plugins]
SuppressDuplicateActivePluginsWarning=Y</codeblock>
    </settingsample>
    <settinglevels>user, database, scriptpath, homedir</settinglevels>
  </setting>
</settings>
*/
func HandleSettings(context *Context, dec *xml.Decoder, start xml.StartElement) error {
	var t table.SettingsXML

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

	{
		for _, row := range t.Rows {
			if !isWebAudience(row.GetAttr("audience"), row.GetAttr("print"), row.GetAttr("deliveryTarget")) {
				continue
			}
			row.SetAttr("class", "setting")
			emitStart("div", row.Attr...)

			emitStart("h3", row.Name.Attr...)
			context.Encoder.WriteRaw(row.Name.Text)
			emitEnd("h3")

			emitStart("div", xml.Attr{Name: xml.Name{Local: "class"}, Value: "settingbody"})

			if row.Default != "" {
				context.Encoder.WriteRaw(`<p>Default: ` + row.Default + `</p>`)
			}

			//setting descriptions always exist
			emitStart("div", row.Desc.Attr...)
			recurse(row.Desc.Content)
			emitEnd("div")

			if row.Levels != "" {
				context.Encoder.WriteRaw(`<p>Can be defined in: ` + row.Levels + `</td>`)
			}

			if len(row.Example.Content) > 0 {
				context.Encoder.WriteRaw(`<p>Example:</p>`)
				recurse(row.Example.Content)
			}

			emitEnd("div")
			emitEnd("div")
		}
	}

	if errors != nil {
		return fmt.Errorf("%v", errors)
	}
	return nil
}
