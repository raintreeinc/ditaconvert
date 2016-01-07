package dita

import (
	"reflect"
	"testing"
)

func TestKeywords(t *testing.T) {
	keywords := Keywords{`<indexterm>A<indexterm>B</indexterm><indexterm>C</indexterm></indexterm>`}
	if !reflect.DeepEqual(keywords.Terms(), []string{"A:B", "A:C"}) {
		t.Fail()
	}
}
