package ditaconvert

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/html"
)

type TokenProcessor func(*ConvertContext, *xml.Decoder, xml.StartElement) error

type Rules struct {
	Rename map[string]string
	Skip   map[string]bool
	Unwrap map[string]bool
	Custom map[string]TokenProcessor
}

type ConvertContext struct {
	Index   *Index
	Topic   *Topic
	Rules   *Rules
	Encoder *html.Encoder
	Output  *bytes.Buffer

	DecodingPath string

	Errors []error
}

func NewConversion(index *Index, topic *Topic) *ConvertContext {
	var out bytes.Buffer
	return &ConvertContext{
		Index:   index,
		Topic:   topic,
		Encoder: html.NewEncoder(&out),
		Output:  &out,
		Rules:   NewDefaultRules(),

		DecodingPath: topic.Path,
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

	defer context.Encoder.Flush()

	if topic.ShortDesc.Content != "" {
		context.Encoder.WriteStart("p")
		// add shortdesc
		if err := context.Parse(topic.ShortDesc.Content); err != nil {
			return err
		}
		context.Encoder.WriteEnd("p")
	}

	// add body
	if err := context.Parse(body); err != nil {
		return err
	}

	// add related links
	context.Encoder.WriteRaw(context.RelatedLinksAsHTML())
	return nil
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

func (context *ConvertContext) EmitWithChildren(dec *xml.Decoder, start xml.StartElement) error {
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
		// is it custom already before naming
		if process, isCustom := context.Rules.Custom[start.Name.Local]; isCustom {
			return process(context, dec, start)
		}

		// handle tag renaming
		if newname, ok := context.Rules.Rename[start.Name.Local]; ok {
			// setAttr(&start, "data-dita", start.Name.Local)
			start.Name.Local = newname
		}

		// is it custom after renaming?
		if process, isCustom := context.Rules.Custom[start.Name.Local]; isCustom {
			return process(context, dec, start)
		}

		return context.EmitWithChildren(dec, start)
	}

	// otherwise, encode as is
	return context.Encoder.Encode(token)
}

func (context *ConvertContext) InlinedImageURL(href string) string {
	if strings.HasPrefix(href, "http:") || strings.HasPrefix(href, "https:") {
		return href
	}

	directory := path.Dir(context.DecodingPath)
	name := path.Join(directory, href)
	data, _, err := context.Index.ReadFile(name)
	if err != nil {
		context.errorf("invalid image link %s: %s", href, err)
		return href
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	ext := strings.ToLower(path.Ext(name))
	if ext == "" {
		context.errorf("invalid image link: %s", href)
		return href
	}

	if ext == "jpg" {
		ext = "jpeg"
	}
	return "data:image/" + ext + ";base64," + encoded
}

func (context *ConvertContext) ResolveLinkInfo(url string) (href, title, synopsis string, internal bool) {
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		return url, "", "", false
	}

	//TODO: add proper link handling, resolve relative to the current directory
	//TODO: extract title
	//TODO: extract title based on hash

	return trimext(url) + ".html", "", "", true
}

func (context *ConvertContext) ShouldSkip(token xml.Token) bool {
	start, isStart := token.(xml.StartElement)
	if !isStart {
		return false
	}
	return context.Rules.Skip[start.Name.Local]
}

func (context *ConvertContext) ShouldUnwrap(token xml.Token) bool {
	start, isStart := token.(xml.StartElement)
	if !isStart {
		return false
	}
	return context.Rules.Unwrap[start.Name.Local]
}

func IsConref(token xml.Token) bool {
	start, isStart := token.(xml.StartElement)
	if !isStart {
		return false
	}
	return getAttr(&start, "conref") != ""
}
