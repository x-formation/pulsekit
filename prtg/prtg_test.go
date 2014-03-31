package prtg

import (
	"bytes"
	"errors"
	"testing"

	"github.com/x-formation/int-tools/pulse"
)

func fixture() (*bytes.Buffer, *int) {
	var (
		buf  bytes.Buffer
		code int
	)
	output, exit = &buf, func(n int) { code = n }
	return &buf, &code
}

func TestOK(t *testing.T) {
	buf, code := fixture()
	OK()
	if buf.String() != "0:0:OK\n" {
		t.Errorf(`expecting buf to be "0:0:OK\n", was %q instead`, buf.String())
	}
	if *code != 0 {
		t.Errorf("expected code to be 0, was %d instead", *code)
	}
}

func TestError(t *testing.T) {
	table := []struct {
		args []interface{}
		exp  string
	}{
		{[]interface{}{errors.New("An error.")}, "2:1:\"An error.\"\n"},
		{[]interface{}{
			&pulse.Agent{Name: "A name 1", Host: "A host 1"},
			&pulse.Agent{Name: "A name 2", Host: "A host 2"},
			&pulse.Agent{Name: "A name 3", Host: "A host 3"},
		}, "2:1:\"A name 1@A host 1\" \"A name 2@A host 2\" \"A name 3@A host 3\"\n"},
		{[]interface{}{"A string"}, "2:1:\"A string\"\n"},
	}
	for i := range table {
		buf, code := fixture()
		Error(table[i].args...)
		if buf.String() != table[i].exp {
			t.Errorf(`expecting buf to be %q, was %q instead`, table[i].exp, buf.String())
		}
		if *code != 1 {
			t.Errorf("expected code to be 1, was %d instead", *code)
		}
	}
}
