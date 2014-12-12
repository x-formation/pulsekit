package cli

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/x-formation/pulsekit"
	"github.com/x-formation/pulsekit/mock"

	"github.com/codegangsta/cli"
	"gopkg.in/v1/yaml"
)

type Flags struct {
	URL     string
	User    string
	Pass    string
	Agent   string
	Project string
        Revision string
	Timeout time.Duration
	Build   int
	Prtg    bool
}

// NewFlags creates default flag set. The values must be the same as the ones
// set in pulsecli.New().
func newFlags() *Flags {
	return &Flags{
		URL:     "http://pulse",
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
	g := flag.NewFlagSet("global pulsecli test", flag.PanicOnError)
	g.String("url", mcli.f.URL, "")
	g.String("agent", mcli.f.Agent, "")
	g.String("project", mcli.f.Project, "")
	g.String("timeout", mcli.f.Timeout.String(), "")
	g.Int("build", mcli.f.Build, "")
	g.Bool("prtg", mcli.f.Prtg, "")

        l := flag.NewFlagSet("local pulsecli test", flag.PanicOnError)
        l.String("revision", mcli.f.Revision, "")
        l.String("pass", mcli.f.Pass, "")

	return cli.NewContext(mcli.cli.app, l, g)
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

func (mcli *MockCLI) Clean() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Clean(mcli.ctx())
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

func (mcli *MockCLI) Personal() (out []interface{}, err []interface{}) {
	mcli.cli.Out = func(i ...interface{}) { out = i }
	mcli.cli.Err = func(i ...interface{}) { err = i }
	mcli.cli.Personal(mcli.ctx())
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
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	out, err := mcli.Init()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestInitErr_ProjectRegex(t *testing.T) {
	mc, mcli, f := fixture()
  mc.Err = make([]error, 1)
	f.Project = "("
	out, err := mcli.Init()
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

func TestInitErr(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err, mc.P = []error{nil, errors.New("err")}, []string{"Pulse CLI"}
	f.Project = "Pulse CLI"
	out, err := mcli.Init()
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
	if e.Error() != "err" {
		t.Error("e to be 'err'")
	}
}

func TestStages(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err = make([]error, 1)
	f.Project = "Pulse CLI"
	out, err := mcli.Stages()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestStagesErr(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err = []error{errors.New("err")}
	f.Project = "Pulse CLI"
	out, err := mcli.Stages()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil && len(err) != 1 {
		t.Error("expected err to not be empty")
	}
	e, ok := err[0].(error)
	if !ok {
		t.Fatalf("expecred err[0] to be of error type, was %T instead", err[0])
	}
	if e == nil {
		t.Error("expected e to be non-nil")
	}
}

func TestStagesErr_MissingProjectName(t *testing.T) {
	mc, mcli, _ := fixture()
	out, err := mcli.Stages()
	expected := "pulsecli: a --project name is missing"
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil && len(err) != 1 {
		t.Error("expected err to not be empty")
	}
	s, ok := err[0].(string)
	if !ok {
		t.Fatalf("expecred err[0] to be of string type, was %T instead", err[0])
	}
	if s != expected {
		t.Errorf("expected %s, got %s", expected, s)
	}
}

func TestBuild(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestLogin(t *testing.T) {
	mc, mcli, _ := fixture()
	out, err := mcli.Login()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestClean(t *testing.T) {
	mc, mcli, f := fixture()
	f.Project = "^Go.*"
	mc.Err, mc.P = make([]error, 3), []string{"Go - Master", "Go - Devel", "C++"}
	out, err := mcli.Clean()
	mc.Check(t)
	if n := len(err); n != 0 {
		t.Fatalf("want len(err)=0; got %d", n)
	}
	if p := []interface{}{"Go - Master", "Go - Devel"}; !reflect.DeepEqual(out, p) {
		t.Fatalf("want out=%v; got %v", p, out)
	}
}

func TestTrigger(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err, mc.P, mc.T = make([]error, 3), []string{"Pulse CLI"}, []string{"message"}
	out, err := mcli.Trigger()
	expected := fmt.Sprintf("%s\t%q", mc.T[0], mc.P[0])
	mc.Check(t)
	if out == nil && len(out) != 1 {
		t.Error("expected out to not be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
	m, ok := out[0].(string)
	if !ok {
		t.Fatalf("expected out[0] to be of type string, was %T instead,", out[0])
	}
	if m != expected {
		t.Errorf("expected %s, got %s", expected, m)
	}
}

func TestTrigger_Empty(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	out, err := mcli.Trigger()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestTrigger_NotMatchingProj(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err, mc.P, mc.T, f.Project = make([]error, 1), []string{"Pulse CLI"}, []string{"message"}, "notmachtingregex"
	out, err := mcli.Trigger()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
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
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	mc.A = pulse.Agents{pulse.Agent{Name: "Agent1", Status: pulse.AgentIdle, Host: "Host1"}}
	out, err := mcli.Health()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Errorf("expected err to be empty, got %v", err)
	}
}

func TestHealthPulseErr_Hanging(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	mc.A = pulse.Agents{
		pulse.Agent{Name: "Agent1", Status: pulse.AgentIdle, Host: "Host1"},
		pulse.Agent{Name: "Agent2", Status: pulse.AgentSync, Host: "Host2"},
	}
	expectederr := "pulsecli: >=50% of Pulse agents are hanging now!"
	out, err := mcli.Health()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil && len(err) != 1 {
		t.Error("expected err to not be empty")
	}
	e, ok := err[0].(string)
	if !ok {
		t.Fatalf("expected err[0] to be of string type, was %T instead", err[0])
	}
	if e != expectederr {
		t.Errorf("expected %s, got %s", expectederr, e)
	}
}

func TestHealthPulseErr_Offline(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	mc.A = pulse.Agents{
		pulse.Agent{Name: "Agent1", Status: pulse.AgentOffline, Host: "Host1"},
		pulse.Agent{Name: "Agent2", Status: pulse.AgentIdle, Host: "Host2"},
	}
	expectedAgent := mc.A[0]
	out, err := mcli.Health()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err == nil && len(err) != 1 {
		t.Error("expected err to not be empty")
	}
	a, ok := err[0].(pulse.Agent)
	if !ok {
		t.Fatalf("expected err[0] to be of pulse.Agent type, was %T instead", err[0])
	}
	if !reflect.DeepEqual(a, expectedAgent) {
		t.Errorf("expected %#v, got %#v", expectedAgent, a)
	}
}

func TestProjects(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err, mc.P = make([]error, 1), []string{"Pulse CLI"}
	expected := mc.P[0]
	out, err := mcli.Projects()
	mc.Check(t)
	if out == nil && len(out) != 1 {
		t.Error("expected out to not be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
	s, ok := out[0].(string)
	if !ok {
		t.Fatalf("expected out[0] to be of string type, was %T instead", out[0])
	}
	if expected != s {
		t.Errorf("expected %s, got %s", expected, s)
	}
}

func TestProjects_Empty(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	out, err := mcli.Projects()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestAgents(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	mc.A = pulse.Agents{pulse.Agent{Name: "Agent1", Status: pulse.AgentIdle, Host: "Host1"}}
	expected := fmt.Sprintf("%s\t%q", "Host1", "Agent1")
	out, err := mcli.Agents()
	mc.Check(t)
	if out == nil && len(out) != 1 {
		t.Error("expected out to not be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
	s, ok := out[0].(string)
	if !ok {
		t.Fatalf("expected out[0] to be of string type, was %T instead", out[0])
	}
	if expected != s {
		t.Errorf("expected %s, got %s", expected, s)
	}
}

func TestAgents_Empty(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	out, err := mcli.Agents()
	mc.Check(t)
	if out != nil && len(out) != 0 {
		t.Error("expected out to be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
}

func TestStatus(t *testing.T) {
	mc, mcli, f := fixture()
	mc.Err, mc.P = make([]error, 3), []string{"LM-X"}
	f.Build = 1
	out, err := mcli.Status()
	mc.Check(t)

	if out == nil || len(out) != 1 {
		t.Error("expected out to not be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
	y, e := yaml.Marshal(map[string][]pulse.BuildResult{fmt.Sprintf("%s (build 1)", mc.P[0]): {}})
	if e != nil {
		t.Error("expected err to be nil")
	}
	s, ok := out[0].(string)
	if !ok {
		t.Fatalf("expected out[0] to be of string type, was %T instead", out[0])
	}
	if string(y) != s {
		t.Errorf("expected %s, got %s", y, out[0])
	}
}

func TestStatusEmpty(t *testing.T) {
	mc, mcli, _ := fixture()
	mc.Err = make([]error, 1)
	out, err := mcli.Status()
	mc.Check(t)

	if out == nil || len(out) != 1 {
		t.Error("expected out to not be empty")
	}
	if err != nil && len(err) != 0 {
		t.Error("expected err to be empty")
	}
	y, e := yaml.Marshal(struct{}{})
	if e != nil {
		t.Error("expected err to be nil")
	}
	s, ok := out[0].(string)
	if !ok {
		t.Fatalf("expected out[0] to be of string type, was %T instead", out[0])
	}
	if string(y) != s {
		t.Errorf("expected %s, got %s", y, s)
	}
}

func TestArtifact(t *testing.T) {
	t.Skip("TODO(ppieprzyk)")
}
