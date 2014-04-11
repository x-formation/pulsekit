package pulsecli

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/x-formation/int-tools/pulseutil"

	"github.com/codegangsta/cli"
)

type Flags struct {
	URL     string
	User    string
	Pass    string
	Agent   string
	Project string
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
	}
}

type MockClient struct {
	Err []error
	A   pulse.Agents
	BI  int64
	BR  []pulse.BuildResult
	I   bool
	L   []pulse.BuildResult
	M   pulse.Messages
	P   []string
	S   []string
	T   []string
	W   <-chan struct{}
	i   int
}

func (mc *MockClient) Error() error {
	if mc.i != len(mc.Err) {
		return fmt.Errorf("pulsecli test: expected Mock to be called %d times,"+
			" was called %d times instead", len(mc.Err), mc.i)
	}
	return nil
}

func (mc *MockClient) err() error {
	i := mc.i
	mc.i++
	if mc.Err == nil || len(mc.Err) <= i {
		return nil
	}
	return mc.Err[i]
}

func (mc *MockClient) Agents() (pulse.Agents, error)       { return mc.A, mc.err() }
func (mc *MockClient) BuildID(reqid string) (int64, error) { return mc.BI, mc.err() }
func (mc *MockClient) BuildResult(project string, id int64) ([]pulse.BuildResult, error) {
	return mc.BR, mc.err()
}
func (mc *MockClient) Clear(project string) error        { return mc.err() }
func (mc *MockClient) Close() error                      { return mc.err() }
func (mc *MockClient) Init(project string) (bool, error) { return mc.I, mc.err() }
func (mc *MockClient) LatestBuildResult(project string) ([]pulse.BuildResult, error) {
	return mc.L, mc.err()
}
func (mc *MockClient) Messages(project string, id int64) (pulse.Messages, error) {
	return mc.M, mc.err()
}
func (mc *MockClient) Projects() ([]string, error)                        { return mc.P, mc.err() }
func (mc *MockClient) Stages(project string) ([]string, error)            { return mc.S, mc.err() }
func (mc *MockClient) Trigger(project string) ([]string, error)           { return mc.T, mc.err() }
func (mc *MockClient) WaitBuild(project string, id int64) <-chan struct{} { return mc.W }

func NewMockClient() *MockClient {
	return &MockClient{}
}

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
	g.Int("build", mcli.f.Build, "")
	g.Bool("prtg", mcli.f.Prtg, "")
	return cli.NewContext(mcli.cli.app, nil, g)
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

func NewMockCLI(mc *MockClient) *MockCLI {
	mcli := &MockCLI{
		cli: New(),
		f:   newFlags(),
	}
	mcli.cli.Client = func(_, _, _ string) (pulse.Client, error) {
		return mc, nil
	}
	return mcli
}

func fixture() (mc *MockClient, mcli *MockCLI, f *Flags) {
	mc = NewMockClient()
	mcli = NewMockCLI(mc)
	f = mcli.f
	return
}

func TestFileStore(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
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
	mc.Err = make([]error, 5)
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
	if err := mc.Error(); err != nil {
		t.Errorf("%v", err)
	}
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
