package pulsecli

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

	"github.com/x-formation/int-tools/pulseutil"
	"github.com/x-formation/int-tools/pulseutil/prtg"
	"github.com/x-formation/int-tools/pulseutil/pulsedev"
	"github.com/x-formation/int-tools/pulseutil/util"

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

// Creds TODO(rjeczalik): document
type Creds struct {
	URL  string
	User string
	Pass string
}

// Login TODO(rjeczalik): document
type CredsStore interface {
	Load() (*Creds, error)
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

// Load TODO(rjeczalik): document
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

// Save TODO(rjeczalik): document
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

// New TODO(rjeczalik): document
type CLI struct {
	Client func(url, user, pass string) (pulse.Client, error)
	Dev    func(c pulse.Client, url, user, pass string) (pulsedev.Tool, error)
	Out    func(...interface{})
	Err    func(...interface{})
	Store  CredsStore
	app    *cli.App
	cred   *Creds
	c      pulse.Client
	v      pulsedev.Tool
	a      *regexp.Regexp
	p      *regexp.Regexp
	patch  string
	rev    string
	n      int64
	d      time.Duration
	prtg   bool
}

// New TODO(rjeczalik): document
func New() *CLI {
	cl := &CLI{
		Client: pulse.NewClient,
		Dev:    pulsedev.New,
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
		cli.StringFlag{Name: "agent, a", Value: ".*", Usage: "Agent name patter"},
		cli.StringFlag{Name: "project, p", Value: ".*", Usage: "Project name pattern"},
		cli.StringFlag{Name: "timeout, t", Value: "15s", Usage: "Maximum wait time"},
		cli.StringFlag{Name: "patch", Usage: "Patch file for a personal build"},
		cli.StringFlag{Name: "revision, r", Value: "HEAD", Usage: "Revision to use for personal build"},
		cli.IntFlag{Name: "build, b", Usage: "Build number"},
		cli.BoolFlag{Name: "prtg", Usage: "PRTG-friendly output"},
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
	a, p, r := ctx.GlobalString("agent"), ctx.GlobalString("project"), ctx.GlobalString("revision")
	if cli.a, err = regexp.Compile(a); err != nil {
		return err
	}
	if cli.p, err = regexp.Compile(p); err != nil {
		return err
	}
	n, t := ctx.GlobalInt("build"), ctx.GlobalString("timeout")
	if cli.d, err = time.ParseDuration(t); err != nil {
		return err
	}
	cli.n = int64(n)
	cli.c.SetTimeout(cli.d)
	cli.rev = r
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
		if _, err := os.Open(p); err != nil {
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
	p := &pulsedev.Personal{
		Patch:    cli.patch,
		Project:  cli.p.String(),
		Revision: cli.rev,
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

// Init TODO(rjeczalik): document
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

// Stages TODO(rjeczalik): document
func (cli *CLI) Stages(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	s, err := cli.c.Stages(cli.p.String())
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

// Build TODO(rjeczalik): document
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

// Login TODO(rjeczalik): document
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

// Trigger TODO(rjeczalik): document
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

// Health TODO(rjeczalik): document
func (cli *CLI) Health(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p := cli.p.String()
	if p != "" && p != ".*" {
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
	if len(a.Filter(pulse.Sync)) >= len(a)/2 {
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

// Projects TODO(rjeczalik): document
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

// Agents TODO(rjeczalik): document
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
		msg = append(msg, fmt.Sprintf("%s\t %q", h, a.Name))
	}
	cli.Out(msg...)
}

// Status TODO(rjeczalik): document
func (cli *CLI) Status(ctx *cli.Context) {
	if err := cli.init(ctx); err != nil {
		cli.Err(err)
		return
	}
	p, err := cli.c.Projects()
	if err != nil {
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

// Run TODO(rjeczalik): document
func (cli *CLI) Run(args []string) {
	cli.app.Run(args)
}
