package pulse

import "testing"

const (
	OkProject     = "Pulse CLI - Success"
	BrokenProject = "Pulse CLI - Failure"
)

func fixture(t *testing.T) Client {
	c, err := NewClient("http://pulse/xmlrpc", "pulse_test", "pulse_test")
	if err != nil {
		t.Fatalf("expected err to be nil, was %v instead", err)
	}
	return c
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

func TestTriggerOk(t *testing.T) {
	c := fixture(t)
	reqid, err := c.Trigger(OkProject)
	if err != nil {
		t.Fatalf("error triggering build %q: %v", OkProject, err)
	}
	if len(reqid) != 1 {
		t.Fatalf("invalid length of the trigger response: len(reqid)=%d", len(reqid))
	}
	_, err = c.BuildID(reqid[0])
	if err != nil {
		t.Fatalf("error requesting build ID: %v", err)
	}
	// TODO(rjeczalik): https://github.com/kolo/xmlrpc/issues/17#issuecomment-39439257
	// _, err := c.BuildResult(OkProject, id)
	// if err != nil {
	//   t.Fatalf("error requesting build state: %v", err)
	// }
}
