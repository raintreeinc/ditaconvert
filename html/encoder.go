package html

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Encoder struct {
	buf bytes.Buffer
	w   io.Writer

	stack  []xml.Name
	invoid bool
}

func NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		buf:   bytes.Buffer{},
		w:     out,
		stack: []xml.Name{},
	}
}

func (enc *Encoder) writeStart(token *xml.StartElement) error {
	enc.stack = append(enc.stack, token.Name)
	enc.invoid = voidElements[token.Name.Local]

	enc.buf.WriteByte('<')
	enc.buf.WriteString(token.Name.Local)
	for _, attr := range token.Attr {
		if attr.Name.Local == "" {
			continue
		}

		enc.buf.WriteByte(' ')
		enc.buf.WriteString(attr.Name.Local)
		enc.buf.WriteString(`="`)
		enc.buf.WriteString(html.EscapeString(attr.Value))
		enc.buf.WriteByte('"')
	}
	enc.buf.WriteByte('>')

	return enc.flush()
}

func (enc *Encoder) writeEnd(token *xml.EndElement) error {
	if len(enc.stack) == 0 {
		return fmt.Errorf("no unclosed tags")
	}

	var current xml.Name
	n := len(enc.stack) - 1
	current, enc.stack = enc.stack[n], enc.stack[:n]
	if current != token.Name {
		return fmt.Errorf("writing end tag %v expected %v", token.Name, current)
	}

	enc.invoid = (len(enc.stack) > 0) && voidElements[enc.stack[len(enc.stack)-1].Local]

	// void elements have only a single tag
	if voidElements[token.Name.Local] {
		return nil
	}

	enc.buf.WriteString("</")
	enc.buf.WriteString(token.Name.Local)
	enc.buf.WriteByte('>')

	return enc.flush()
}

func (enc *Encoder) WriteRaw(data string) error {
	_, err := enc.buf.WriteString(data)
	return err
}

func (enc *Encoder) voiderror() error {
	return fmt.Errorf("content not allowed inside void tag %s", enc.stack[len(enc.stack)-1].Local)
}

func (enc *Encoder) Encode(token xml.Token) error {
	switch token := token.(type) {
	case xml.StartElement:
		return enc.writeStart(&token)
	case xml.EndElement:
		return enc.writeEnd(&token)
	case xml.CharData:
		if enc.invoid {
			return enc.voiderror()
		}
		enc.buf.Write([]byte(token))
		return enc.flush()
	case xml.Comment:
		if enc.invoid {
			return enc.voiderror()
		}
		enc.buf.WriteString("<!--")
		enc.buf.Write([]byte(token))
		enc.buf.WriteString("-->")
		return enc.flush()
	case xml.ProcInst:
		if enc.invoid {
			return enc.voiderror()
		}
		// skip processing instructions
		return nil
	case xml.Directive:
		if enc.invoid {
			return enc.voiderror()
		}
		// skip directives
		return nil
	default:
		panic("invalid token")
	}
}

func (enc *Encoder) flush() error {
	if enc.buf.Len() > 1<<8 {
		return enc.Flush()
	}
	return nil
}

func (enc *Encoder) Flush() error {
	_, err := enc.buf.WriteTo(enc.w)
	enc.buf.Reset()
	return err
}

// writeQuoted writes s to w surrounded by quotes. Normally it will use double
// quotes, but if s contains a double quote, it will use single quotes.
// It is used for writing the identifiers in a doctype declaration.
// In valid HTML, they can't contain both types of quotes.
func writeQuoted(w *bytes.Buffer, s string) error {
	var q byte = '"'
	if strings.Contains(s, `"`) {
		q = '\''
	}
	if err := w.WriteByte(q); err != nil {
		return err
	}
	if _, err := w.WriteString(s); err != nil {
		return err
	}
	if err := w.WriteByte(q); err != nil {
		return err
	}
	return nil
}

// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
var voidElements = map[string]bool{
	"area":    true,
	"base":    true,
	"br":      true,
	"col":     true,
	"command": true,
	"embed":   true,
	"hr":      true,
	"img":     true,
	"input":   true,
	"keygen":  true,
	"link":    true,
	"meta":    true,
	"param":   true,
	"source":  true,
	"track":   true,
	"wbr":     true,
}
