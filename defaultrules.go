package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/html"
)

func NewDefaultRules() *Rules {
	return &Rules{
		Rename: map[string]Renaming{
			// conversion
			"xref": {"a", ""},
			"link": {"a", ""},

			//lists
			"choices":         {"ul", ""},
			"choice":          {"li", ""},
			"steps-unordered": {"ul", ""},
			"steps":           {"ol", ""},
			"step":            {"li", ""},
			"substeps":        {"ol", ""},
			"substep":         {"li", ""},

			"b":     {"strong", ""},
			"i":     {"em", ""},
			"lines": {"pre", ""},

			"codeblock": {"pre", "codeblock"},
			"pre":       {"pre", "pre"},

			"codeph":      {"samp", "codeph"},
			"cmdname":     {"span", "cmdname"},
			"cmd":         {"span", ""},
			"shortcut":    {"span", "shortcut"},
			"wintitle":    {"b", ""},
			"filepath":    {"span", "filepath"},
			"menucascade": {"span", "menucascade"},

			"option":   {"span", "option"},
			"synph":    {"span", ""},
			"delim":    {"span", ""},
			"sep":      {"span", ""},
			"parmname": {"span", ""},

			"userinput": {"kbd", "userinput"},

			"image": {"img", ""},

			// ui
			"uicontrol": {"b", ""},

			// divs
			"context":    {"div", ""},
			"result":     {"div", ""},
			"stepresult": {"div", ""},
			"stepxmp":    {"div", ""},
			"info":       {"div", ""},
			"note":       {"div", ""},
			"refsyn":     {"div", ""},
			"bodydiv":    {"div", ""},
			"fig":        {"div", ""},

			"prereq":  {"div", ""},
			"postreq": {"div", ""},

			// tables
			"simpletable": {"table", ""},
			"sthead":      {"thead", ""},
			"strow":       {"tr", ""},
			"stentry":     {"td", ""},

			"colspec": {"colgroup", ""},

			"row":   {"tr", ""},
			"entry": {"td", ""},

			// RAINTREE SPECIFIC
			"keystroke": {"b", "key"},
			"secright":  {"span", ""},

			// faq
			"faq":          {"dl", ""},
			"faq-question": {"dt", "dlterm"},
			"faq-answer":   {"dd", ""},

			//UI items
			"ui-item-list": {"dl", ""},

			"ui-item-name":        {"dt", "dlterm"},
			"ui-item-description": {"dd", ""},

			// setup options
			"setup-options": {"dl", ""},

			"setup-option-name":        {"dt", "dlterm"},
			"setup-option-description": {"dd", ""},

			"settingdesc": {"div", ""},
			"settingname": {"h4", ""},

			"section":    {"div", "section"},
			"example":    {"div", ""},
			"sectiondiv": {"div", ""},
			"title":      {"h4", "sectiontitle"},

			// ??
			"dt": {"dt", "dlterm"},
		},
		Skip: map[string]bool{
			"br":            true,
			"draft-comment": true,

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
		Custom: map[string]TokenProcessor{
			"a": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				var href, desc string
				var internal bool

				href = getAttr(&start, "href")
				if href != "" {
					href, _, desc, internal = context.ResolveLinkInfo(href)
					setAttr(&start, "href", href)
				}

				if desc != "" && getAttr(&start, "title") == "" {
					setAttr(&start, "title", desc)
				}

				setAttr(&start, "scope", "")
				if internal && href != "" {
					//setAttr(&start, "data-link", href)
				}

				if getAttr(&start, "format") != "" && href != "" {
					setAttr(&start, "format", "")
					ext := strings.ToLower(path.Ext(href))
					if ext != ".htm" && ext != ".html" {
						setAttr(&start, "download", path.Base(href))
					} else {
						setAttr(&start, "target", "_blank")
					}
				}

				return context.EmitWithChildren(dec, start)
			},
			"img": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				href := getAttr(&start, "href")
				//setAttr(&start, "src", context.InlinedImageURL(href))
				setAttr(&start, "src", href)
				setAttr(&start, "href", "")

				placement := getAttr(&start, "placement")
				setAttr(&start, "placement", "")
				if placement == "break" {
					context.Encoder.WriteStart("p",
						xml.Attr{Name: xml.Name{Local: "class"}, Value: "image"})
				}

				err := context.EmitWithChildren(dec, start)

				if placement == "break" {
					context.Encoder.WriteEnd("p")
				}

				return err
			},
			"data": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				datatype := strings.ToLower(getAttr(&start, "datatype"))
				if datatype == "rttutorial" {
					dec.Skip()
					href := getAttr(&start, "href")
					tag := fmt.Sprintf(`<video controls src="%s">Browser does not support video playback.</video>`, html.NormalizeURL(href))
					context.Encoder.WriteRaw(tag)
					return nil
				}

				return context.EmitWithChildren(dec, start)
			},
			"imagemap": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				context.Encoder.WriteRaw(`<div class="conversion-error">TODO imagemap</div>`)
				//TODO:
				dec.Skip()
				return nil
			},

			// RAINTREE SPECIFIC
			"settingdefault": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				val, _ := html.XMLText(dec)
				if val != "" {
					return context.Encoder.WriteRaw("<p>Default value: " + val + "</p>")
				}
				return nil
			},
			"settinglevels": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				context.check(context.Encoder.WriteRaw("<p>Levels where it can be defined:</p>"))
				return context.EmitWithChildren(dec, start)
			},
			"settingsample": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				context.check(context.Encoder.WriteRaw("<p>Example:</p>"))
				return context.EmitWithChildren(dec, start)
			},

			"note": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				typ := getAttr(&start, "type")
				if typ == "" {
					typ = "note"
				}
				setAttr(&start, "type", "")
				setAttr(&start, "class", typ)

				context.check(context.Encoder.WriteStart("div", start.Attr...))
				context.check(context.Encoder.WriteRaw(`<span class="` + typ + `title">` + strings.Title(typ) + `:</span> `))
				err := context.Recurse(dec)
				context.check(context.Encoder.WriteEnd("div"))
				return err
			},
			"menucascade": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				start.Name.Local = "span"

				setAttr(&start, "class", "menucascade")
				// encode starting tag and attributes
				if err := context.Encoder.Encode(start); err != nil {
					return err
				}

				first := true

				// recurse on child tokens
				for {
					token, err := dec.Token()
					if err != nil {
						if err == io.EOF {
							return nil
						}
						return err
					}

					if _, ended := token.(xml.EndElement); ended {
						break
					}
					if _, starting := token.(xml.StartElement); starting {
						if !first {
							context.Encoder.WriteRaw("&gt;")
						}
						first = false
					}
					if err := context.Handle(dec, token); err != nil {
						return err
					}
				}

				// always encode ending tag
				context.check(context.Encoder.Encode(xml.EndElement{start.Name}))

				return nil
			},
		},
	}
}
