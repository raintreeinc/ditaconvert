package ditaconvert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/html"
)

type ConvertContext struct {
	Index   *Index
	Topic   *Topic
	Encoder *html.Encoder
	Output  *bytes.Buffer

	Directory string

	Errors []error
}

func NewConversion(index *Index, topic *Topic) *ConvertContext {
	var out bytes.Buffer
	return &ConvertContext{
		Index:   index,
		Topic:   topic,
		Encoder: html.NewEncoder(&out),
		Output:  &out,

		Directory: path.Dir(topic.Path),
	}
}

func (context *ConvertContext) check(err error) bool {
	if err != nil {
		context.Errors = append(context.Errors, err)
		return true
	}
	return false
}

func (context *ConvertContext) errorf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	context.Errors = append(context.Errors, err)
}

func (context *ConvertContext) Run() error {
	topic := context.Topic.Original
	if topic == nil {
		return fmt.Errorf("no associated topic")
	}

	body := ""
	for _, node := range topic.Elements {
		if IsBodyTag(node.XMLName.Local) {
			if body != "" {
				context.errorf("multiple body tags")
				continue
			}
			body = node.Content
		}
	}

	if body == "" {
		context.errorf("empty body tag")
	}

	err := context.Parse(body)
	context.Encoder.WriteRaw(context.RelatedLinksAsHTML())
	context.Encoder.Flush()
	return err
}

// checks wheter dita tag corresponds to some "root element"
func IsBodyTag(tag string) bool { return strings.Contains(tag, "body") }

func (context *ConvertContext) Parse(data string) error {
	dec := xml.NewDecoder(strings.NewReader(data))
	return context.Recurse(dec)
}

var ErrSkip = errors.New("")

func (context *ConvertContext) Recurse(dec *xml.Decoder) error {
	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if _, ended := token.(xml.EndElement); ended {
			return nil
		}
		if err := context.Handle(dec, token); err != nil {
			return err
		}
	}
}

func (context *ConvertContext) Handle(dec *xml.Decoder, token xml.Token) error {
	// should we just skip the token?
	if context.ShouldSkip(token) {
		dec.Skip()
		return nil
	}
	// should we inline something from somewhere else?
	if IsConref(token) {
		return context.HandleConref(dec, token.(xml.StartElement))
	}

	startdepth := context.Encoder.Depth()
	defer func() {
		if startdepth != context.Encoder.Depth() {
			panic("mismatched start and end handle")
		}
	}()

	// should we unwrap the tag?
	if context.ShouldUnwrap(token) {
		return context.Recurse(dec)
	}

	// is it a starting token?
	if start, isStart := token.(xml.StartElement); isStart {
		// is it special already before naming
		if process, isSpecial := Rules.Special[start.Name.Local]; isSpecial {
			return process(context, dec, start)
		}

		// handle tag renaming
		if newname, ok := Rules.Rename[start.Name.Local]; ok {
			start.Name.Local = newname
		}

		// is it special after renaming?
		if process, isSpecial := Rules.Special[start.Name.Local]; isSpecial {
			return process(context, dec, start)
		}

		// encode starting tag and attributes
		if err := context.Encoder.Encode(start); err != nil {
			return err
		}

		// recurse on child tokens
		err := context.Recurse(dec)

		// always encode ending tag
		context.check(context.Encoder.Encode(xml.EndElement{start.Name}))

		return err
	}

	// otherwise, encode as is
	return context.Encoder.Encode(token)
}

func (context *ConvertContext) HandleConref(dec *xml.Decoder, token xml.StartElement) error {
	dec.Skip()
	context.Encoder.WriteRaw(`<div class="conversion-error">TODO: insert conref</div>`)
	return nil
}

func (context *ConvertContext) ShouldSkip(token xml.Token) bool {
	start, isStart := token.(xml.StartElement)
	if !isStart {
		return false
	}
	return Rules.Skip[start.Name.Local]
}

func (context *ConvertContext) ShouldUnwrap(token xml.Token) bool {
	start, isStart := token.(xml.StartElement)
	if !isStart {
		return false
	}
	return Rules.Unwrap[start.Name.Local]
}

func IsConref(token xml.Token) bool {
	start, isStart := token.(xml.StartElement)
	if !isStart {
		return false
	}
	return attribute(start, "conref") != ""
}

func attribute(n xml.StartElement, key string) (val string) {
	for _, attr := range n.Attr {
		if attr.Name.Local == key {
			return attr.Value
		}
	}
	return ""
}

type TokenProcessor func(*ConvertContext, *xml.Decoder, xml.StartElement) error

var Rules = struct {
	Rename  map[string]string
	Skip    map[string]bool
	Unwrap  map[string]bool
	Special map[string]TokenProcessor
}{
	Rename: map[string]string{
		// conversion
		"xref": "a",
		"link": "a",

		//lists
		"choices":         "ul",
		"choice":          "li",
		"steps-unordered": "ul",
		"steps":           "ol",
		"step":            "li",
		"substeps":        "ol",
		"substep":         "li",

		"i":     "em",
		"lines": "pre",

		"codeblock": "code",

		"codeph":      "span",
		"cmdname":     "span",
		"cmd":         "span",
		"shortcut":    "span",
		"wintitle":    "span",
		"filepath":    "span",
		"menucascade": "span",

		"synph":    "span",
		"delim":    "span",
		"sep":      "span",
		"parmname": "span",

		"userinput": "kbd",

		"image": "img",

		// ui
		"uicontrol": "span",

		// divs
		"context":    "div",
		"result":     "div",
		"stepresult": "div",
		"stepxmp":    "div",
		"info":       "div",
		"note":       "div",
		"refsyn":     "div",
		"bodydiv":    "div",
		"fig":        "div",

		"prereq":  "div",
		"postreq": "div",

		// tables
		"simpletable": "table",
		"sthead":      "thead",
		"strow":       "tr",
		"stentry":     "td",

		"colspec": "colgroup",

		"row":   "tr",
		"entry": "td",

		// RAINTREE SPECIFIC
		"keystroke": "span",
		"secright":  "span",

		// faq
		"faq":          "dl",
		"faq-question": "dt",
		"faq-answer":   "dd",

		//UI items
		"ui-item-list": "dl",

		"ui-item-name":        "dt",
		"ui-item-description": "dd",

		// setup options
		"setup-options": "dl",

		"setup-option-name":        "dt",
		"setup-option-description": "dd",

		"settingdesc": "div",
		"settingname": "h4",

		"section":    "div",
		"example":    "div",
		"sectiondiv": "div",
		"title":      "h4",
	},
	Skip: map[string]bool{
		"br":            true,
		"draft-comment": true,

		"imagemap": true, // TODO: Handle properly

		// RAINTREE SPECIFIC
		"settinghead": true,
	},
	Unwrap: map[string]bool{
		"dlentry": true,
		"tgroup":  true,

		// RAINTREE SPECIFIC
		"ui-item":      true,
		"faq-item":     true,
		"setup-option": true,

		"settings": true,
		"setting":  true,
	},
	Special: map[string]TokenProcessor{},
}
