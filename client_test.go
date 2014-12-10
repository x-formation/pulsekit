package pulse

import (
	"testing"

	"github.com/rjeczalik/fakerpc"
)

func fixture(t *testing.T) (Client, func()) {
	url, teardown := fakerpc.Fixture(t)
	c, err := NewClient(url+"/xmlrpc", "pulse_test", "pulse_test")
	if err != nil {
		t.Fatalf("expected err to be nil, was %q instead", err)
	}
	t.Parallel()
	return c, func() { c.Close(); teardown() }
}

func TestAgents(t *testing.T) {
	c, teardown := fixture(t)
	defer teardown()
	a, err := c.Agents()
	if err != nil {
		t.Fatalf("expected err to be nil, was %q instead", err)
	}
	if len(a) == 0 {
		t.Fatal("expected len(a) to be non-zero")
	}
	for _, a := range a {
		if a.Status == "" {
			t.Error("expected a.Status to be non-empty")
		}
		if a.Host == "" {
			t.Error("expected a.Host to be non-empty")
		}
		if a.Name == "" {
			t.Error("expected a.Name to be non-empty")
		}
	}
}

func TestBuildID(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestBuildResult(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestClear(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestClose(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestConfigStage(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestInit(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestLatestBuildResult(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestProjects(t *testing.T) {
	c, teardown := fixture(t)
	defer teardown()
	p, err := c.Projects()
	if err != nil {
		t.Fatalf("expected err to be nil, was %v instead", err)
	}
	if len(p) == 0 {
		t.Fatal("expected len(p) to be non-zero")
	}
	for _, p := range p {
		if p == "" {
			t.Error("expected p to be non-empty")
		}
	}
}

func TestSetStage(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestStages(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTrigger(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestArtifact(t *testing.T) {
	t.Skip("TODO(ppieprzyk)")
}
