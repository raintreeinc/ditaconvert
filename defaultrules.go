package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/normalize"
)

func DefaultProcessAttributes(context *ConvertContext, start *xml.StartElement) {
	var href, desc string
	var internal bool

	href = getAttr(start, "href")
	if href != "" {
		if start.Name.Local == "img" || start.Name.Local == "image" || start.Name.Local == "fig" {
			setAttr(start, "href", context.InlinedImageURL(href))
		} else {
			href, _, desc, internal = context.ResolveLinkInfo(href)
			setAttr(start, "href", href)
		}
	}

	if desc != "" && getAttr(start, "title") == "" {
		setAttr(start, "title", desc)
	}

	if internal && href != "" {
		setAttr(start, "data-link", href)
	}

	if getAttr(start, "format") != "" && href != "" {
		setAttr(start, "download", path.Base(href))
	}
}

func NewDefaultRules() *Rules {
	return &Rules{
		ProcessAttributes: DefaultProcessAttributes,

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
		Special: map[string]TokenProcessor{
			"data": func(context *ConvertContext, dec *xml.Decoder, start xml.StartElement) error {
				datatype := strings.ToLower(getAttr(&start, "datatype"))
				if datatype == "rttutorial" {
					dec.Skip()
					DefaultProcessAttributes(context, &start)
					href := getAttr(&start, "href")
					tag := fmt.Sprintf(`<video controls src="%s">Browser does not support video playback.</video>`, normalize.URL(href))
					context.Encoder.WriteRaw(tag)
					return nil
				}

				return context.EmitWithChildren(dec, start)
			},
		},
	}
}
