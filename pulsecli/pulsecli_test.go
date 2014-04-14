package pulsecli

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/x-formation/int-tools/pulseutil"
	"github.com/x-formation/int-tools/pulseutil/mock"

	"github.com/codegangsta/cli"
)

type Flags struct {
	URL     string
	User    string
	Pass    string
	Agent   string
	Project string
	Timeout time.Duration
	Build   int
	Prtg    bool
}

// NewFlags creates default flag set. The values must be the same as the ones
// set in pulsecli.New().
func newFlags() *Flags {
	return &Flags{
		URL:     "http://pulse/xmlrpc",
		Agent:   ".*",
		Project: ".*",
		Timeout: 15 * time.Second,
	}
}

var errInvalidBuild = &pulse.InvalidBuildError{Status: pulse.BuildUnknown}

type MockCLI struct {
	cli *CLI
	f   *Flags
}

func (mcli *MockCLI) ctx() *cli.Context {
	g := flag.NewFlagSet("pulsecli test", flag.PanicOnError)
	g.String("addr", mcli.f.URL, "")
	g.String("pass", mcli.f.Pass, "")
	g.String("agent", mcli.f.Agent, "")
	g.String("project", mcli.f.Project, "")
	g.String("timeout", mcli.f.Timeout.String(), "")
	g.Int("build", mcli.f.Build, "")
	g.Bool("prtg", mcli.f.Prtg, "")
	return cli.NewContext(mcli.cli.app, nil, g)
}

func (mcli *MockCLI) Wait() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Wait(mcli.ctx())
	return
}

func (mcli *MockCLI) Login() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Login(mcli.ctx())
	return
}

func (mcli *MockCLI) Trigger() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Trigger(mcli.ctx())
	return
}

func (mcli *MockCLI) Init() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Init(mcli.ctx())
	return
}

func (mcli *MockCLI) Health() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Health(mcli.ctx())
	return
}

func (mcli *MockCLI) Projects() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Projects(mcli.ctx())
	return
}

func (mcli *MockCLI) Stages() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Stages(mcli.ctx())
	return
}

func (mcli *MockCLI) Agents() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Agents(mcli.ctx())
	return
}

func (mcli *MockCLI) Status() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Status(mcli.ctx())
	return
}

func (mcli *MockCLI) Build() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Build(mcli.ctx())
	return
}

func NewMockCLI(c pulse.Client) *MockCLI {
	mcli := &MockCLI{
		cli: New(),
		f:   newFlags(),
	}
	mcli.cli.Client = func(_, _, _ string) (pulse.Client, error) {
		return c, nil
	}
	return mcli
}

func fixture() (mc *mock.Client, mcli *MockCLI, f *Flags) {
	mc = mock.NewClient()
	mcli = NewMockCLI(mc)
	f = mcli.f
	return
}

func TestFileStore(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestPersonal(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestWait(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err = make([]error, 2)
	f.Build, f.Project, f.Timeout = 3, "Pulse CLI", time.Second
	out, err := mcli.Wait()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestWait_MissingProject(t *testing.T) {
	mc, mcli, f := fixture()
	f.Timeout = time.Second
	out, err := mcli.Wait()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil || len(err) == 0 {
		t.Error("expected err to not be empty")
	}
}

func TestWait_NormalizeErr(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err = []error{errors.New("err")}
	f.Build, f.Timeout, f.Project = 0, time.Second, "Pulse CLI"
	out, err := mcli.Wait()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil || len(err) != 1 {
		t.Fatalf("expected err!=nil, len(err)=1; was err=%v, len(err)=%d",
			err, len(err))
	}
	e, ok := err[0].(error)
	if !ok {
		t.Fatalf("expecred err[0] to be of error type, was %T instead", err[0])
	}
	if e == nil {
		t.Error("expected e to be non-nil")
	}
}

func TestWait_Timeout(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err, mc.BR = []error{nil, nil}, []pulse.BuildResult{{Complete: false}}
	f.Build, f.Timeout, f.Project = 1, 50*time.Millisecond, "Pulse CLI"
	out, err := mcli.Wait()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil || len(err) != 1 {
		t.Fatalf("expected err!=nil, len(err)=1; was err=%v, len(err)=%d",
			err, len(err))
	}
	e, ok := err[0].(error)
	if !ok {
		t.Fatalf("expecred err[0] to be of error type, was %T instead", err[0])
	}
	if e != pulse.ErrTimeout {
		t.Errorf("expected e to be pulse.ErrTimeout, was %q instead", e)
	}
}

func TestInit(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestStages(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestBuild(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestLogin(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTrigger(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestHealthProject(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err = make([]error, 7)
	mc.M = []pulse.Message{
		{Severity: pulse.SeverityError, Message: "error #1"},
		{Severity: pulse.SeverityWarning, Message: "warn #1"},
		{Severity: pulse.SeverityWarning, Message: "warn #2"},
		{Severity: pulse.SeverityInfo, Message: "info #1"},
		{Severity: pulse.SeverityInfo, Message: "info #2"},
	}
	mc.P = []string{
		"LM-X - Tier 1",
		"LM-X - Tier 2",
		"License Statistics",
	}
	f.Project = "LM-X"
	expected := []string{
		"LM-X - Tier 1",
		"LM-X - Tier 2",
		"error #1",
		"warn #1",
		"warn #2",
		string(pulse.SeverityError),
		string(pulse.SeverityWarning),
	}
	notexpected := []string{
		"License Statistics",
		"info #1",
		"info #2",
		string(pulse.SeverityInfo),
	}
	out, err := mcli.Health()
	mc.Check(t)
	if out != nil && len(out) > 0 {
		t.Errorf("expected out to be empty, was %v instead", err)
	}
	if err == nil || len(err) == 0 {
		t.Fatal("expected err to be non-empty")
	}
	s := fmt.Sprintln(err...)
	for _, exp := range expected {
		if !strings.Contains(s, exp) {
			t.Errorf("expected s to contain %q", exp)
		}
	}
	for _, noexp := range notexpected {
		if strings.Contains(s, noexp) {
			t.Errorf("expected s to not contain %q", noexp)
		}
	}
}

func TestHealthPulse(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestProjects(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestAgents(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestStatus(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}
