package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/raintreeinc/ditaconvert/imagemap"
)

func ConvertImageMap(context *Context, dec *xml.Decoder, start xml.StartElement) error {
	var m imagemap.XML
	err := dec.DecodeElement(&m, &start)
	if err != nil {
		return err
	}

	href := m.Image.Href

	directory := path.Dir(context.DecodingPath)
	name := path.Join(directory, href)
	data, _, err := context.Index.ReadFile(name)
	if err != nil {
		context.errorf("invalid image link %s: %s", href, err)
		return nil
	}

	ext := strings.Trim(strings.ToLower(path.Ext(name)), ".")
	if ext == "" {
		context.errorf("invalid image link: %s", href)
		return nil
	}
	if ext == "jpg" {
		ext = "jpeg"
	}

	content, err := imagemap.FromXML(ext, data, m.Area)
	if err != nil {
		context.check(err)
		return nil
	}

	emitStart := func(tag string, attrs ...xml.Attr) {
		context.check(context.Encoder.WriteStart(tag, attrs...))
	}
	emitEnd := func(tag string) {
		context.check(context.Encoder.WriteEnd(tag))
	}

	emitStart("div", attr("class", "imagemap"))
	defer emitEnd("div")

	defer func() {
		emitStart("div",
			attr("class", "imagemap-tip"))
		context.Encoder.WriteRaw("Click any highlighted region for details.")
		emitEnd("div")
	}()

	emitStart("div",
		attr("class", "imagemap-layers"),
		attr("style", "position: relative; font-size: 0; width: 100%;"))
	defer emitEnd("div")

	emitStart("img",
		attr("src", content.Image),
		attr("style", "width: 100%; max-width:"+strconv.Itoa(content.Size.X)+"px;"),
	)
	emitEnd("img")

	emitStart("div",
		attr("class", "imagemap-overlay"),
		attr("style",
			"position:absolute;"+
				"padding: 0; margin: 0;"+
				"left:0; top: 0; right: 0; bottom: 0;"+
				"max-width:"+strconv.Itoa(content.Size.X)+"px;"),
	)
	defer emitEnd("div")

	for _, area := range content.Areas {
		emitStart("a",
			attr("title", area.Alt),
			attr("alt", area.Alt),

			attr("style",
				fmt.Sprintf("position: absolute; display: block; left: %.3f%%; top: %.3f%%; width: %.3f%%; height: %.3f%%;",
					float64(area.Min.X)*100/float64(content.Size.X),
					float64(area.Min.Y)*100/float64(content.Size.Y),
					float64(area.Max.X-area.Min.X)*100/float64(content.Size.X),
					float64(area.Max.Y-area.Min.Y)*100/float64(content.Size.Y),
				)),
			attr("href", area.Href),
		)
		emitEnd("a")
	}

	return nil
}

func attr(name string, value string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: name}, Value: value}
}
