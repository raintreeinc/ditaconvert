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

func (context *Context) ResolveKeyRef(keyref string) (abspath, itempath string) {
	if keyref == "" {
		return "", ""
	}

	items := strings.SplitN(keyref, "/", 2)
	if len(items) < 2 {
		context.errorf("invalid conkeyref %v", keyref)
		return "", ""
	}

	abspath, ok := context.Index.KeyDef[items[0]]
	if !ok {
		context.errorf("keydef missing for %v (%v)", items[0], keyref)
		return "", ""
	}

	return abspath, items[1]
}

func (context *Context) HandleConref(dec *xml.Decoder, start xml.StartElement) error {
	dec.Skip()

	conref, conkeyref, conrefend := getAttr(&start, "conref"), getAttr(&start, "conkeyref"), getAttr(&start, "conrefend")
	keyfile, keypath := context.ResolveKeyRef(conkeyref)

	startfile, startpath := SplitLink(conref)
	endfile, endpath := SplitLink(conrefend)

	// startfile and endfile are relative to current direcotry
	// keyfile is absolute relative to the root

	if startfile != "" {
		startfile = path.Join(path.Dir(context.DecodingPath), startfile)
	}
	if endfile != "" {
		endfile = path.Join(path.Dir(context.DecodingPath), endfile)
	}

	// conref is missing, try to use conkeyref instead
	if startfile == "" && keyfile != "" {
		if startpath != "" || endpath != "" {
			return errors.New("invalid conkeyref setup")
		}
		startfile, startpath = keyfile, keypath
	}

	// conrefend is missing, fallback to either conref or conkeyref
	if endfile == "" && endpath == "" {
		endfile, endpath = startfile, startpath
	}

	// start/end files are both missing, use the current file
	if startfile == "" && endfile == "" {
		startfile, endfile = context.DecodingPath, context.DecodingPath
	}

	// sanity check
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
