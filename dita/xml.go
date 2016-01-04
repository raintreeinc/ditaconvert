package dita

import (
	"encoding/xml"
	"io"
	"strings"
)

func xmltext(decoder *xml.Decoder, start *xml.StartElement) (string, error) {
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
			sub, err := xmltext(decoder, &token)
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

func xmlstriptags(xmlcontent string) (string, error) {
	if xmlcontent == "" {
		return "", nil
	}

	dec := xml.NewDecoder(strings.NewReader(xmlcontent))
	return xmltext(dec, nil)
}
