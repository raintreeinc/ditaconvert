package ditaconvert

import (
	"encoding/xml"
	"path"
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

	ext := strings.ToLower(path.Ext(name))
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

	context.check(
		context.Encoder.WriteStart("img",
			xml.Attr{Name: xml.Name{Local: "class"}, Value: "imagemap"},
			xml.Attr{Name: xml.Name{Local: "src"}, Value: content.Image}))
	context.check(
		context.Encoder.WriteEnd("img"))

	return nil
}
