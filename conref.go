package ditaconvert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

func SameRootElement(a, b string) bool {
	return strings.EqualFold(path.Dir(a), path.Dir(b))
}

func (context *ConvertContext) HandleConref(dec *xml.Decoder, start xml.StartElement) error {
	dec.Skip()

	conref, conrefend := getAttr(&start, "conref"), getAttr(&start, "conrefend")

	startfile, startpath := SplitLink(conref)
	endfile, endpath := SplitLink(conrefend)
	if endfile == "" && endpath == "" {
		endfile, endpath = startfile, startpath
	}

	if startfile == "" && endfile == "" {
		startfile, endfile = context.DecodingPath, context.DecodingPath
	} else {
		startfile = path.Join(path.Dir(context.DecodingPath), startfile)
		endfile = path.Join(path.Dir(context.DecodingPath), endfile)
	}

	if startfile != endfile {
		return errors.New("conref and conrefend are in different files: " + startfile + " --> " + endfile)
	}

	if !SameRootElement(startpath, endpath) {
		return errors.New("conref and conrefend have different root elements: " + conref + " --> " + conrefend)
	}
	if startpath == "" || endpath == "" {
		return errors.New("invalid conref path: " + conref + " --> " + conrefend)
	}

	previousPath := context.DecodingPath
	defer func() {
		context.DecodingPath = previousPath
	}()

	data, _, err := context.Index.ReadFile(startfile)
	if err != nil {
		return fmt.Errorf("problem opening %v: %v", startfile, err)
	}

	subdec := xml.NewDecoder(bytes.NewReader(data))
	subfirst, err := WalkNodePath(subdec, startpath)
	if err != nil {
		if err == io.EOF {
			return errors.New("did not find conref: " + conref)
		}
		return err
	}

	var subtoken xml.Token = subfirst
	endingid := path.Base(endpath)
	for {
		err := context.Handle(subdec, subtoken)
		if err != nil {
			return err
		}

		// is it ending?
		if substart, isStart := subtoken.(xml.StartElement); isStart {
			if strings.EqualFold(endingid, getAttr(&substart, "id")) {
				return nil
			}
		}

		if _, isEnd := subtoken.(xml.EndElement); isEnd {
			return errors.New("did not find conrefend: " + conrefend)
		}

		subtoken, err = subdec.Token()
		if err != nil {
			return err
		}
	}

	return nil
}
