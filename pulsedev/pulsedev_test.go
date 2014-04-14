package pulsedev

import (
	"testing"

	"github.com/x-formation/int-tools/pulseutil/mock"
)

func fixture(t *testing.T) (*mock.Client, Tool) {
	mc := mock.NewClient()
	tool, err := New(mc, "http://pulse", "pulse_test", "pulse_test")
	if err != nil {
		t.Fatalf("expected err to be nil, was %q instead", err)
	}
	return mc, tool
}

func TestPersonalOK(t *testing.T) {
	_, tool := fixture(t)
	p := []Personal{{
		Patch:   "pulsedev_test.go",
		Project: "Pulse CLI - Failure",
	}, {
		Patch:    "pulsedev_test.go",
		Project:  "Pulse CLI - Failure",
		Revision: "HEAD",
	}}
	for _, p := range p {
		id, err := tool.Personal(&p)
		if err != nil {
			t.Errorf("expected err to be nil, was %q instead", err)
		}
		if id <= 0 {
			t.Errorf("expected i to be greater than 0, was %d instead", id)
		}
	}
}

func TestPersonalErr(t *testing.T) {
	_, tool := fixture(t)
	p := []Personal{{
		Patch:   "pulsedev_test.go",
		Project: "X",
	}, {
		Project:  "Pulse CLI - Failure",
		Revision: "HEAD",
	}, {}}
	for _, p := range p {
		id, err := tool.Personal(&p)
		if err == nil {
			t.Error("expected err to be non-nil")
		}
		if id != 0 {
			t.Error("expected i to be 0")
		}
	}
}
