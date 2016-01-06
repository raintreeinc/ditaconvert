package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/html"
)

func NewDefaultRules() *Rules {
	return &Rules{
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

				if internal && href != "" {
					setAttr(&start, "data-link", href)
				}

				if getAttr(&start, "format") != "" && href != "" {
					setAttr(&start, "download", path.Base(href))
				}

				return context.EmitWithChildren(dec, start)
			},
			"img": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				href := getAttr(&start, "href")
				setAttr(&start, "src", context.InlinedImageURL(href))
				setAttr(&start, "href", "")

				placement := getAttr(&start, "placement")
				if placement == "break" {
					context.Encoder.Encode(xml.StartElement{
						Name: xml.Name{Local: "p"},
						Attr: []xml.Attr{
							{Name: xml.Name{Local: "class"}, Value: "image"},
						},
					})
				}

				err := context.EmitWithChildren(dec, start)

				if placement == "break" {
					context.Encoder.Encode(xml.EndElement{
						Name: xml.Name{Local: "p"},
					})
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
		},
	}
}
