package pulse

import "testing"

func fixture(t *testing.T) Client {
	c, err := NewClient("http://pulse/xmlrpc", "pulse_test", "pulse_test")
	if err != nil {
		t.Fatalf("expected err to be nil, was %v instead", err)
	}
	t.Parallel()
	return c
}

func TestAgents(t *testing.T) {
	c := fixture(t)
	a, err := c.Agents()
	if err != nil {
		t.Fatalf("expected err to be nil, was %v instead", err)
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

func TestInit(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestLatestBuildResult(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestProjects(t *testing.T) {
	c := fixture(t)
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

func TestStages(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTrigger(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}
