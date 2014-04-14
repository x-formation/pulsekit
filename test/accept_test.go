package accept_test

import (
	"testing"
	"time"

	"github.com/x-formation/int-tools/pulseutil"
	"github.com/x-formation/int-tools/pulseutil/util"
)

const (
	OkProject     = "Pulse CLI - Success"
	BrokenProject = "Pulse CLI - Failure"
)

func fixture(t *testing.T) pulse.Client {
	c, err := pulse.NewClient("http://pulse/xmlrpc", "pulse_test", "pulse_test")
	if err != nil {
		t.Fatalf("expected err to be nil, was %v instead", err)
	}
	t.Parallel()
	return c
}

func accept(t *testing.T, p string, ok bool) {
	c := fixture(t)
	reqid, err := c.Trigger(p)
	if err != nil {
		t.Fatalf("error triggering build %q: %v", p, err)
	}
	if len(reqid) != 1 {
		t.Fatalf("invalid length of the trigger response: len(reqid)=%d", len(reqid))
	}
	id, err := c.BuildID(reqid[0])
	if err != nil {
		t.Fatalf("error requesting build ID: %v", err)
	}
	done := util.Wait(c, 125*time.Millisecond, p, id)
	select {
	case <-done:
	case <-time.After(time.Minute):
	}
	build, err := c.BuildResult(p, id)
	if err != nil {
		t.Fatalf("error requesting build state: %v", err)
	}
	if len(build) != 1 {
		t.Errorf("expected len(build) to be 1, was %d instead", len(build))
	}
	if !util.Pending(build) {
		if !build[0].Complete {
			t.Errorf("expected p=%q build=%d to be completed", p, id)
		}
		if build[0].Success != ok {
			t.Errorf("expected p=%q build=%d to be successful=%v", p, id, ok)
		}
	}
}

func TestAcceptOk(t *testing.T) {
	accept(t, OkProject, true)
}

func TestAcceptBroken(t *testing.T) {
	accept(t, BrokenProject, false)
}
