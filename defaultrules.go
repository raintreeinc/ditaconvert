package ditaconvert

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/raintreeinc/ditaconvert/html"
)

func TODO(context *Context, dec *xml.Decoder, start xml.StartElement) error {
	context.Encoder.WriteRaw(`<div class="conversion-error">TODO ` + start.Name.Local + `</div>`)
	dec.Skip()
	return nil
}

/* TODO: unify all rules
func Rename(tag string, attrs... xml.Attr) TokenProcessor {
	return func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
		start.Name.Local = tag
		return context.EmitWithChildren(dec, start)
	}
}

func RenameWithClass(tag string, class string) TokenProcessor {
	return func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
		start.Name.Local = tag
		setAttr(&start, "class", class)
		return context.EmitWithChildren(dec, start)
	}
}*/

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
			"step":            {"li", ""}, // handled with a custom rule
			"substeps":        {"ol", ""},
			"substep":         {"li", ""}, // handled with a custom rule

			"b":     {"strong", ""},
			"i":     {"em", ""},
			"lines": {"pre", ""},

			"codeblock": {"pre", "codeblock"},
			"pre":       {"pre", "pre"},

			"codeph":      {"samp", "codeph"},
			"cmdname":     {"span", "cmdname"},
			"cmd":         {"span", "cmd"},
			"shortcut":    {"span", "shortcut"},
			"wintitle":    {"span", "wintitle"},
			"filepath":    {"span", "filepath"},
			"menucascade": {"span", "menucascade"},
			"msgph":       {"span", "msgph"},
			"reportitem":  {"span", "reportitem"},

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

			"colspec": {"colgroup", ""},

			"row":   {"tr", ""},
			"entry": {"td", ""},

			// RAINTREE SPECIFIC
			"keystroke": {"b", "key"},
			"secright":  {"span", "secright"},

			// faq
			"faq":          {"dl", ""},
			"faq-item":     {"div", ""},
			"faq-question": {"dt", "dlterm"},
			"faq-answer":   {"dd", ""},

			//UI items
			"ui-item-list": {"dl", ""},
			"ui-item":      {"div", ""},
			//"ui-item-name":        {"dt", "dlterm"},
			"ui-item-description": {"dd", ""},

			// setup options
			"setup-options":            {"dl", ""},
			"setup-option":             {"div", ""},
			"setup-option-name":        {"dt", "dlterm"},
			"setup-option-description": {"dd", ""},

			"settingdesc": {"div", ""},
			"settingname": {"h2", ""},

			"section":    {"div", "section"},
			"example":    {"div", ""},
			"sectiondiv": {"div", ""},
			"title":      {"h2", "sectiontitle"},

			// ??
			"dlentry": {"div", ""},
			"dt":      {"dt", "dlterm"},

			"settings": {"div", "settings"},
			"setting":  {"div", "setting"},
		},
		Skip: map[string]bool{
			"br":            true,
			"draft-comment": true,

			// RAINTREE SPECIFIC
			"settinghead": true,
		},
		Unwrap: map[string]bool{
			"tgroup": true,
		},
		Custom: map[string]TokenProcessor{
			"a": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
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
					if ext == ".pdf" || ext == ".doc" || ext == ".xml" || ext == ".rtf" || ext == ".zip" || ext == ".exe" {
						setAttr(&start, "download", path.Base(href))
					} else {
						setAttr(&start, "target", "_blank")
					}
				}

				return context.EmitWithChildren(dec, start)
			},
			"img": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
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
			"data": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				datatype := strings.ToLower(getAttr(&start, "datatype"))
				if datatype == "rttutorial" {
					dec.Skip()
					href := getAttr(&start, "href")

					const videof = `` +
						`<video controls>` +
						`	<source src="%s" type="video/mp4">` +
						`	<object classid="clsid:d27cdb6e-ae6d-11cf-96b8-444553540000" codebase="http://fpdownload.macromedia.com/pub/shockwave/cabs/flash/swflash.cab#version=8,0,0,0" ` +
						`		<param name="SRC" value="http://ie.microsoft.com/testdrive/IEBlog/Common/player.swf?file=%s">` +
						`		<p>Video playback not supported</p>` +
						`	</object>` +
						`</video>`

					srcurl := html.NormalizeURL(href)
					urlarg := url.QueryEscape(href)

					context.Encoder.WriteRaw(fmt.Sprintf(videof, srcurl, urlarg))
					return nil
				}

				return context.EmitWithChildren(dec, start)
			},
			"imagemap": ConvertImageMap,

			// RAINTREE SPECIFIC
			"settingdefault": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				val, _ := html.XMLText(dec)
				if val != "" {
					return context.Encoder.WriteRaw("<p>Default value: " + val + "</p>")
				}
				return nil
			},
			"settinglevels": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				context.check(context.Encoder.WriteRaw("<p>Levels where it can be defined:</p>"))
				return context.EmitWithChildren(dec, start)
			},
			"settingsample": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				context.check(context.Encoder.WriteRaw("<p>Example:</p>"))
				return context.EmitWithChildren(dec, start)
			},

			"note": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				typ := getAttr(&start, "type")
				if typ == "other" {
					typ = getAttr(&start, "othertype")
				}
				if typ == "" {
					typ = "note"
				}
				setAttr(&start, "type", "")
				setAttr(&start, "othertype", "")
				setAttr(&start, "class", "note")
				mdiclass := "note-outline"
				switch typ {
				case "tip":
					mdiclass = "lightbulb-outline"
				case "caution":
					mdiclass = "alert"
				case "Extra":
					mdiclass = "key"
				case "Rev-Edition":
					mdiclass = "elevation-rise"
				case "PDF":
					mdiclass = "book-open"
				}

				context.check(context.Encoder.WriteStart("div", start.Attr...))
				context.check(context.Encoder.WriteRaw(`<i class="mdi mdi-` + mdiclass + `" title="` + typ + `"></i> `))
				context.check(context.Encoder.WriteStart("span"))
				err := context.Recurse(dec)
				context.check(context.Encoder.WriteEnd("span"))
				context.check(context.Encoder.WriteEnd("div"))
				return err
			},
			"menucascade": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
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

			"ui-item-name": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				start.Name.Local = "dt"

				setAttr(&start, "class", "ui-item-name")
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
							context.Encoder.WriteRaw(", ")
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

			"simpletable": HandleSimpleTable,
			"table":       HandleTable,

			"step": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				start.Name.Local = "li"

				context.check(context.Encoder.Encode(start))
				if getAttr(&start, "importance") == "optional" {
					context.Encoder.WriteRaw("(Optional) ")
				}
				err := context.Recurse(dec)
				context.check(context.Encoder.Encode(xml.EndElement{start.Name}))

				return err
			},
			"substep": func(context *Context, dec *xml.Decoder, start xml.StartElement) error {
				start.Name.Local = "li"

				context.check(context.Encoder.Encode(start))
				if getAttr(&start, "importance") == "optional" {
					context.Encoder.WriteRaw("(Optional) ")
				}
				err := context.Recurse(dec)
				context.check(context.Encoder.Encode(xml.EndElement{start.Name}))

				return err
			},
		},
	}
}
