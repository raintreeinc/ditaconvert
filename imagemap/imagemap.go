package imagemap

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"
)

type Content struct {
	Image string `json:"image"` // in base64
	Areas []Area `json:"areas"`
	Size  Point  `json:"size"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Area struct {
	Href string `json:"href"`
	Alt  string `json:"alt,omitempty"`
	Min  Point  `json:"min"`
	Max  Point  `json:"max"`
}

type XMLArea struct {
	Shape  string `xml:"shape"`
	Coords string `xml:"coords"`
	XRef   struct {
		Href string `xml:"href,attr"`
		Alt  string `xml:",chardata"`
	} `xml:"xref"`
}

type XML struct {
	Image struct {
		Href string `xml:"href,attr"`
	} `xml:"image"`
	Area []XMLArea `xml:"area"`
}

func FromXML(mime string, data []byte, areas []XMLArea) (content *Content, err error) {
	content = &Content{}

	for _, area := range areas {
		switch area.Shape {
		case "rect":
			tokens := strings.Split(area.Coords, ",")
			if len(tokens) != 4 {
				return nil, errors.New("invalid imagemap coords \"" + area.Coords + "\"")
			}

			x0, err0 := strconv.Atoi(strings.Trim(tokens[0], " "))
			y0, err1 := strconv.Atoi(strings.Trim(tokens[1], " "))
			x1, err2 := strconv.Atoi(strings.Trim(tokens[2], " "))
			y1, err3 := strconv.Atoi(strings.Trim(tokens[3], " "))


			if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
				return nil, errors.New("invalid imagemap coords \"" + area.Coords + "\"")
			}

			content.Areas = append(content.Areas, Area{
				Href: area.XRef.Href,
				Alt:  area.XRef.Alt,
				Min:  Point{x0, y0},
				Max:  Point{x1, y1},
			})
		default:
			return nil, errors.New("unhandled imagemap shape \"" + area.Shape + "\"")
		}
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	content.Size.X = img.Bounds().Dx()
	content.Size.Y = img.Bounds().Dy()

	encoded := base64.StdEncoding.EncodeToString(data)
	content.Image = "data:image/" + mime + ";base64," + encoded

	return content, nil
}
