package dev

import (
	"testing"

	"github.com/x-formation/pulsekit/mock"

	"github.com/rjeczalik/fakerpc"
)

func fixture(t *testing.T) (*mock.Client, Tool, func()) {
	addr, teardown := fakerpc.Fixture(t)
	mc := mock.NewClient()
	tool, err := New(mc, addr, "pulse_test", "pulse_test")
	if err != nil {
		t.Skip("pulsekit/dev: skipping test: ", err)
	}
	return mc, tool, teardown
}

func TestPersonalOK(t *testing.T) {
	_, tool, teardown := fixture(t)
	defer teardown()
	p := []Personal{{
		Patch:   "dev_test.go",
		Project: "Pulse CLI - Failure",
	}, {
		Patch:    "dev_test.go",
		Project:  "Pulse CLI - Failure",
		Revision: "HEAD",
	}}
	for i, p := range p {
		id, err := tool.Personal(&p)
		if err != nil {
			t.Errorf("expected err to be nil, was %q instead (i=%d)", err, i)
		}
		if id <= 0 {
			t.Errorf("expected i to be greater than 0, was %d instead (i=%d)", id, i)
		}
	}
}

func TestPersonalErr(t *testing.T) {
	_, tool, teardown := fixture(t)
	defer teardown()
	p := []Personal{{
		Patch:   "dev_test.go",
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
