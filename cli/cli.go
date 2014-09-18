package cli

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/x-formation/pulsekit"
	"github.com/x-formation/pulsekit/dev"
	"github.com/x-formation/pulsekit/prtg"
	"github.com/x-formation/pulsekit/util"

	"github.com/codegangsta/cli"
	"gopkg.in/v1/yaml"
)

var defaultErr = func(args ...interface{}) {
	for _, arg := range args {
		fmt.Fprintln(os.Stderr, arg)
	}
	os.Exit(1)
}

var defaultOut = func(args ...interface{}) {
	for _, arg := range args {
		fmt.Println(arg)
	}
	os.Exit(0)
}

// Creds holds information required to authenticate an user session from
// the Pulse Remote API endpoint.
type Creds struct {
	URL  string
	User string
	Pass string
}

// CredsStore persists the Creds struct.
type CredsStore interface {
	// Load gives Creds loaded from a persisted storage.
	Load() (*Creds, error)
	// Save saves given Creds to a persisted storage.
	Save(*Creds) error
}

type fileStore struct{}

func config(mode int) (f *os.File, err error) {
	u, err := user.Current()
	if err != nil {
		return
	}
	return os.OpenFile(filepath.Join(u.HomeDir, ".pulsecli"), mode, 0644)
}

func (fileStore) Load() (c *Creds, err error) {
	f, err := config(os.O_RDONLY)
	if err != nil {
		return
	}
	defer f.Close()
	c = &Creds{}
	var b bytes.Buffer
	if _, err = io.Copy(&b, f); err != nil {
		return
	}
	if err = yaml.Unmarshal(b.Bytes(), c); err != nil {
		return
	}
	return
}

func (fileStore) Save(c *Creds) error {
	f, err := config(os.O_TRUNC | os.O_CREATE | os.O_WRONLY)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewBuffer(b))
	return err
}

// CLI is a facade that implements cmd/pulsecli tool.
type CLI struct {
	// Client is used to communicate with a Pulse server.
	Client func(url, user, pass string) (pulse.Client, error)
	// Dev TODO(rjeczalik): document
	Dev func(c pulse.Client, url, user, pass string) (dev.Tool, error)
	// Out terminates the application, writing to os.Stdout and calling os.Exit(0)
	Out func(...interface{})
	// Err terminates the application, writing to os.Stdout and calling os.Exit(1)
	Err func(...interface{})
	// Store is used to persist authorization information.
	Store CredsStore
	app   *cli.App
	cred  *Creds
	c     pulse.Client
	v     dev.Tool
	a     *regexp.Regexp
	p     *regexp.Regexp
	s     *regexp.Regexp
	o     *regexp.Regexp
	patch string
	rev   string
	n     int64
	d     time.Duration
	prtg  bool
}

// New gives a new CLI, sets up command line handling and registers subcommands.
func New() *CLI {
	cl := &CLI{
		Client: pulse.NewClient,
		Dev:    dev.New,
		Store:  fileStore{},
		Err:    defaultErr,
		Out:    defaultOut,
		app:    cli.NewApp(),
		cred:   &Creds{},
	}
	cl.app.Name, cl.app.Version = "pulsecli", "0.1.0"
	cl.app.Usage = "a command-line client for a Pulse server"
	cl.app.Flags = []cli.Flag{
		cli.StringFlag{Name: "addr", Value: "http://pulse/xmlrpc", Usage: "Pulse Remote API endpoint"},
		cli.StringFlag{Name: "user", Usage: "Pulse user name"},
		cli.StringFlag{Name: "pass", Usage: "Pulse user password"},
		cli.StringFlag{Name: "agent, a", Value: ".*", Usage: "Agent name pattern"},
		cli.StringFlag{Name: "project, p", Value: ".*", Usage: `Project name pattern (or "personal")`},
		cli.StringFlag{Name: "stage, s", Value: ".*", Usage: "Stage name pattern"},
		cli.StringFlag{Name: "timeout, t", Value: "15s", Usage: "Maximum wait time"},
		cli.StringFlag{Name: "patch", Usage: "Patch file for a personal build"},
		cli.StringFlag{Name: "revision, r", Value: "HEAD", Usage: "Revision to use for personal build"},
		cli.IntFlag{Name: "build, b", Usage: "Build number"},
		cli.BoolFlag{Name: "prtg", Usage: "PRTG-friendly output"},
		cli.StringFlag{Name: "output, o", Value: ".", Usage: "Output for fetched artifacts"},
	}
	cl.app.Commands = []cli.Command{{
		Name:   "login",
		Usage:  "Creates or updates session for current user",
		Action: cl.Login,
	}, {
		Name:   "trigger",
		Usage:  "Triggers a build",
		Action: cl.Trigger,
	}, {
		Name:   "clean",
		Usage:  "Cleans working directory",
		Action: cl.Clean,
	}, {
		Name:   "init",
		Usage:  "Initialises a project",
		Action: cl.Init,
	}, {
		Name:   "health",
		Usage:  "Performs a health check",
		Action: cl.Health,
	}, {
		Name:   "projects",
		Usage:  "Lists all projct names",
		Action: cl.Projects,
	}, {
		Name:   "stages",
		Usage:  "Lists all stage names",
		Action: cl.Stages,
	}, {
		Name:   "agents",
		Usage:  "Lists all agent names",
		Action: cl.Agents,
	}, {
		Name:   "status",
		Usage:  `Lists build's status`,
		Action: cl.Status,
	}, {
		Name:   "build",
		Usage:  "Gives build ID associated with given request ID",
		Action: cl.Build,
	}, {
		Name:   "wait",
		Usage:  "Waits for a build to complete",
		Action: cl.Wait,
	}, {
		Name:   "personal",
		Usage:  "Sends a personal build request",
		Action: cl.Personal,
	}, {
		Name:   "artifact",
		Usage:  "Downloads all the artifact files",
		Action: cl.Artifact,
	}}
	return cl
}

func (cli *CLI) init(ctx *cli.Context) error {
	if ctx.GlobalBool("prtg") {
		cli.Err, cli.Out = prtg.Err, prtg.Out
	}
	var err error
	if cli.cred, err = cli.Store.Load(); err == nil {
		cli.c, err = cli.Client(cli.cred.URL, cli.cred.User, cli.cred.Pass)
	}
	if err != nil {
		cli.cred = &Creds{ctx.GlobalString("addr"), ctx.GlobalString("user"), ctx.GlobalString("pass")}
		if cli.c, err = cli.Client(cli.cred.URL, cli.cred.User, cli.cred.Pass); err != nil {
			return err
		}
	}
	a, p, s, o := ctx.GlobalString("agent"), ctx.GlobalString("project"), ctx.GlobalString("stage"), ctx.GlobalString("output")
	if cli.a, err = regexp.Compile(a); err != nil {
		return err
	}
	if cli.p, err = regexp.Compile(p); err != nil {
		return err
	}
	if cli.s, err = regexp.Compile(s); err != nil {
		return err
	}
	if cli.o, err = regexp.Compile(o); err != nil {
		return err
	}
	if cli.d, err = time.ParseDuration(ctx.GlobalString("timeout")); err != nil {
		return err
	}
	cli.n, cli.rev = int64(ctx.GlobalInt("build")), ctx.GlobalString("revision")
	cli.c.SetTimeout(cli.d)
	return nil
}

// Personal TODO(rjeczalik): document
func (cli *CLI) Personal(ctx *cli.Context) {
	err := cli.init(ctx)
	if err != nil {
		cli.Err(err)
		return
	}
	if p := ctx.GlobalString("patch"); p != "" {
		if _, err = os.Stat(p); err != nil {
			cli.Err(err)
			return
		}
		cli.patch = p
	}
	url := cli.cred.URL
	if n := strings.Index(cli.cred.URL, "/xmlrpc"); n != -1 {
		url = url[:n]
	}
	if cli.v, err = cli.Dev(cli.c, url, cli.cred.User, cli.cred.Pass); err != nil {
		cli.Err(err)
		return
	}
	p := &dev.Personal{
		Patch:    cli.patch,
		Project:  cli.p.String(),
		Revision: cli.rev,
	}
	if s := cli.s.String(); s != "" && s != ".*" {
		s, err := cli.c.Stages(p.Project)
		if err != nil {
			cli.Err(err)
			return
		}
		if len(s) != 0 {
			p.Stages = make([]string, 0, len(s))
			for _, s := range s {
				if cli.s.MatchString(s) {
					continue
				}
				p.Stages = append(p.Stages, s)
			}
			if len(p.Stages) == 0 {
				cli.Err(fmt.Sprintf("pulsecli: no stages found that match %q", cli.s.String()))
				return
			}
		}
	}
	id, err := cli.v.Personal(p)
	if err != nil {
		cli.Err(err)
		return
	}
	cli.Out(id)
}

// Wait TODO(rjeczalik): document
func (cli *CLI) Wait(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p := cli.p.String()
	if p == "" || p == ".*" {
		cli.Err("pulsecli: a --project name is missing")
		return
	}
	id, err := util.NormalizeBuildOrRequestID(cli.c, p, cli.n)
	if err != nil {
		cli.Err(err)
		return
	}
	select {
	case <-time.After(cli.d):
		err = pulse.ErrTimeout
	case err = <-util.Wait(cli.c, time.Second, p, id):
	}
	if err != nil {
		cli.Err(err)
		return
	}
}

// Init is a a command line interface to Init method of a pulse.Client.
// It outputs a pair of boolean and name, one per line for every project requested,
// separated by a tab. Boolean value indicates whether initialization request
// was accepted of rejected by Pulse server.
func (cli *CLI) Init(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
		return
	}
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		ok, err := cli.c.Init(p)
		if err != nil {
			cli.Err(err)
			return
		}
		msg = append(msg, fmt.Sprintf("%v\t%q", ok, p))
	}
	cli.Out(msg...)
}

// Stages is a command line interface to a Stages method of a pulse.Client.
// It outputs names one per line and expectes an exact project name to be passed
// through command line, not a regex.
func (cli *CLI) Stages(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p := cli.p.String()
	if p == "" || p == ".*" {
		cli.Err("pulsecli: a --project name is missing")
		return
	}
	s, err := cli.c.Stages(p)
	if err != nil {
		cli.Err(err)
		return
	}
	msg := make([]interface{}, 0, len(s))
	for _, s := range s {
		msg = append(msg, s)
	}
	cli.Out(msg...)
}

// Build is a command line interface to a BuildID method of a pulse.Client.
func (cli *CLI) Build(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	reqID := ctx.Args().First()
	if reqID == "" {
		cli.Err("pulsecli: a request ID is missing")
		return
	}
	id, err := cli.c.BuildID(reqID)
	if err != nil {
		cli.Err(err)
		return
	}
	cli.Out(id)
}

// Login writes Pulse Remote API authentication information to a ~/.pulsecli file
// in a YAML format. On consecutive runs it updates every non-empty field that is
// passed from command line. It fails in doing so, when given credentials are not
// valid.
func (cli *CLI) Login(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	old := []*string{&cli.cred.URL, &cli.cred.User, &cli.cred.Pass}
	for i, s := range []string{ctx.GlobalString("addr"), ctx.GlobalString("user"), ctx.GlobalString("pass")} {
		if s != "" {
			(*old[i]) = s
		}
	}
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	if err := cli.Store.Save(cli.cred); err != nil {
		cli.Err(err)
		return
	}
	cli.Out()
}

// Clean is a command line interface o a Clear method of a pulse.Client.
// It triggers a working directory cleanup for each project requested.
func (cli *CLI) Clean(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p, err := cli.c.Projects()
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		if err = cli.c.Clear(p); err != nil {
			cli.Err(err)
			return
		}
		msg = append(msg, p)
	}
	cli.Out(msg...)
}

// Trigger is a command line interface to a Trigger method of a pulse.Client.
// It outputs pairs of a request ID and a project name one per line, for every
// project requested. Values are separated by a tab.
func (cli *CLI) Trigger(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
		return
	}
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		if err = cli.c.Clear(p); err != nil {
			cli.Err(err)
			return
		}
		s, err := cli.c.Trigger(p)
		if err != nil {
			cli.Err(err)
			return
		}
		for _, s := range s {
			msg = append(msg, fmt.Sprintf("%s\t%q", s, p))
		}
	}
	cli.Out(msg...)
}

// Health checks a status of a Pulse server or a project.
// Pulse server health check fails when at least one agent is offline (meaning
// that pulse-agent on that machine has died or communication went down) or at
// least half of the agents are in synchronization state, which means Pulse server
// has deadlocked because of whatever reasons (the most favorite is DNS handling,
// apparently it is possible for a DNS request to not time out, which makes
// all Pulse worker threads just stuck).
// A project health check requests error and warning messages for latest build
// of a given project, and fails when the list is not empty.
func (cli *CLI) Health(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	if p := cli.p.String(); p != "" && p != ".*" {
		cli.healthProject(ctx)
	} else {
		cli.healthPulse(ctx)
	}
}

func (cli *CLI) healthProject(ctx *cli.Context) {
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
		return
	}
	all := make(map[string]pulse.Messages)
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		id, err := util.NormalizeBuildOrRequestID(cli.c, p, cli.n)
		if err != nil {
			cli.Err(err)
			return
		}
		m, err := cli.c.Messages(p, id)
		if err != nil {
			cli.Err(err)
			return
		}
		if m = m.FilterOut(pulse.Info); len(m) > 0 {
			all[fmt.Sprintf("%s (build %d)", p, id)] = m
		}
	}
	if len(all) == 0 {
		cli.Out()
	} else {
		y, err := yaml.Marshal(all)
		if err != nil {
			cli.Err(err)
			return
		}
		cli.Err(string(y))
	}
}

func (cli *CLI) healthPulse(ctx *cli.Context) {
	a, err := cli.c.Agents()
	if err != nil {
		cli.Err(err)
		return
	}
	if len(a.Filter(pulse.Sync)) >= (len(a)+1)/2 {
		cli.Err("pulsecli: >=50% of Pulse agents are hanging now!")
		return
	}
	if a = a.Filter(pulse.Offline); len(a) == 0 {
		cli.Out()
	} else {
		msg := make([]interface{}, 0, len(a))
		for i := range a {
			msg = append(msg, a[i])
		}
		cli.Err(msg...)
	}
}

// Projects is a command line interface to a Projects method of a pulse.Client.
// It outputs a name for every project, one per line.
func (cli *CLI) Projects(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
		return
	}
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		msg = append(msg, p)
	}
	cli.Out(msg...)
}

// Agents is a command line interface to a Agents method of a pulse.Client.
// It prints hostname-agentname pairs one per line for every agent.
// It tries to extract the hostname from every agent's URL.
func (cli *CLI) Agents(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	a, err := cli.c.Agents()
	if err != nil {
		cli.Err(err)
		return
	}
	msg := make([]interface{}, 0, len(a))
	for _, a := range a {
		h := a.Host
		if u, err := url.Parse(h); err == nil {
			if host, _, err := net.SplitHostPort(u.Host); err == nil {
				h = host
			}
		}
		msg = append(msg, fmt.Sprintf("%s\t%q", h, a.Name))
	}
	cli.Out(msg...)
}

// Status is a command line interface to a BuildResult methos of a pulse.Client.
// It outputs []BuildResult for every requested project in an YAML format.
func (cli *CLI) Status(ctx *cli.Context) {
	err := cli.init(ctx)
	if err != nil {
		cli.Err(err)
		return
	}
	var p []string
	if cli.p.String() == pulse.ProjectPersonal {
		p = append(p, pulse.ProjectPersonal)
	} else if p, err = cli.c.Projects(); err != nil {
		cli.Err(err)
		return
	}
	m := make(map[string][]pulse.BuildResult)
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		id, err := util.NormalizeBuildOrRequestID(cli.c, p, cli.n)
		if err != nil {
			cli.Err(err)
			return
		}
		b, err := cli.c.BuildResult(p, id)
		if err != nil {
			cli.Err(err)
			return
		}
		m[fmt.Sprintf("%s (build %d)", p, id)] = b
	}
	y, err := yaml.Marshal(m)
	if err != nil {
		cli.Err(err)
		return
	}
	cli.Out(string(y))
}

// Run takes command line arguments and starts the application.
func (cli *CLI) Run(args []string) {
	cli.app.Run(args)
}

// Artifact is a a command line interface to Artifact method of a pulse.Client.
// It downloads all artifacts captured from given project and build number
func (cli *CLI) Artifact(ctx *cli.Context) {
	var projects []string
	err := cli.init(ctx)
	if err != nil {
		cli.Err(err)
		return
	}
	if cli.p.String() == pulse.ProjectPersonal {
		projects = append(projects, pulse.ProjectPersonal)
	} else if projects, err = cli.c.Projects(); err != nil {
		cli.Err(err)
		return
	}
	var build int64
	dir, url := cli.o.String(), strings.Trim(cli.cred.URL, "/xmlrpc")
	for _, p := range projects {
		if !cli.p.MatchString(p) {
			continue
		}
		if build, err = util.NormalizeBuildOrRequestID(cli.c, p, cli.n); err != nil {
			cli.Err(err)
			return
		}
		if err = cli.c.Artifact(build, p, dir, url); err != nil {
			cli.Err(err)
			return
		}
	}
}
