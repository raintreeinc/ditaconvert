package main

import (
	"encoding/xml"
	"io"
	"strings"
)

func XMLText(decoder *xml.Decoder, start *xml.StartElement) (string, error) {
	r := ""
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return r, io.EOF
		}

		switch token := token.(type) {
		case xml.EndElement:
			return r, nil
		case xml.CharData:
			r += string(token)
		case xml.StartElement:
			sub, err := Text(decoder, &token)
			r += sub
			if err != nil {
				return r, err
			}
		case xml.Comment: // ignore
		case xml.ProcInst: // ignore
		case xml.Directive: // ignore
		default:
			panic("unknown token")
		}
	}
}

func XMLStripTags(xmlcontent string) (string, error) {
	if xmlcontent == "" {
		return "", nil
	}

	dec := xml.NewDecoder(strings.NewReader(xmlcontent))
	return Text(dec, nil)
}
