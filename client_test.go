package pulse

import (
	"testing"
	"time"
)

const (
	OkProject     = "Pulse CLI - Success"
	BrokenProject = "Pulse CLI - Failure"
)

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

func TestWaitBuild(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestGetBuildID(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func accept(t *testing.T, project string, ok bool) {
	c := fixture(t)
	reqid, err := c.Trigger(project)
	if err != nil {
		t.Fatalf("error triggering build %q: %v", project, err)
	}
	if len(reqid) != 1 {
		t.Fatalf("invalid length of the trigger response: len(reqid)=%d", len(reqid))
	}
	id, err := c.BuildID(reqid[0])
	if err != nil {
		t.Fatalf("error requesting build ID: %v", err)
	}
	done := c.WaitBuild(project, id)
	select {
	case <-done:
	case <-time.After(time.Minute):
	}
	build, err := c.BuildResult(project, id)
	if err != nil {
		t.Fatalf("error requesting build state: %v", err)
	}
	if len(build) != 1 {
		t.Errorf("expected len(build) to be 1, was %d instead", len(build))
	}
	if !Pending(build) {
		if !build[0].Complete {
			t.Errorf("expected project=%q build=%d to be completed", project, id)
		}
		if build[0].Success != ok {
			t.Errorf("expected project=%q build=%d to be successful=%v", project, id, ok)
		}
	}
}

func TestAcceptOk(t *testing.T) {
	accept(t, OkProject, true)
}

func TestAcceptBroken(t *testing.T) {
	accept(t, BrokenProject, false)
}
