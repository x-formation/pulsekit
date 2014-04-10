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

	"github.com/x-formation/int-tools/pulseutil"
	"github.com/x-formation/int-tools/pulseutil/prtg"

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
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewBuffer(b))
	return err
}

// New TODO(rjeczalik): document
type CLI struct {
	Err  func(...interface{})
	Out  func(...interface{})
	app  *cli.App
	c    pulse.Client
	a    *regexp.Regexp
	p    *regexp.Regexp
	s    CredsStore
	n    int64
	prtg bool
	cred *Creds
}

// New TODO(rjeczalik): document
func New() *CLI {
	cl := &CLI{
		Err:  defaultErr,
		Out:  defaultOut,
		app:  cli.NewApp(),
		s:    fileStore{},
		cred: &Creds{},
	}
	cl.app.Name, cl.app.Version = "pulsecli", "0.1.0"
	cl.app.Usage = "a command-line client for a Pulse server"
	cl.app.Flags = []cli.Flag{
		cli.StringFlag{Name: "addr", Value: "http://pulse/xmlrpc", Usage: "Pulse Remote API endpoint"},
		cli.StringFlag{Name: "user", Value: "", Usage: "Pulse user name"},
		cli.StringFlag{Name: "pass", Value: "", Usage: "Pulse user password"},
		cli.StringFlag{Name: "agent, a", Value: ".*", Usage: "Agent name patter"},
		cli.StringFlag{Name: "project, p", Value: ".*", Usage: "Project name pattern"},
		cli.IntFlag{Name: "build, b", Value: 0, Usage: "Build number"},
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
	}}
	return cl
}

func (cli *CLI) init(ctx *cli.Context) {
	if ctx.GlobalBool("prtg") {
		cli.Err, cli.Out = prtg.Err, prtg.Out
	}
	var err error
	if cli.cred, err = cli.s.Load(); err == nil {
		cli.c, err = pulse.NewClient(cli.cred.URL, cli.cred.User, cli.cred.Pass)
	}
	if err != nil {
		cli.cred = &Creds{ctx.GlobalString("addr"), ctx.GlobalString("user"), ctx.GlobalString("pass")}
		if cli.c, err = pulse.NewClient(cli.cred.URL, cli.cred.User, cli.cred.Pass); err != nil {
			cli.Err(err)
		}
	}
	a, p, n := ctx.GlobalString("agent"), ctx.GlobalString("project"), ctx.GlobalInt("build")
	if cli.a, err = regexp.Compile(a); err != nil {
		cli.Err(err)
	}
	if cli.p, err = regexp.Compile(p); err != nil {
		cli.Err(err)
	}
	cli.n = int64(n)
}

// Init TODO(rjeczalik): document
func (cli *CLI) Init(ctx *cli.Context) {
	cli.init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
	}
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		ok, err := cli.c.Init(p)
		if err != nil {
			cli.Err(err)
		}
		msg = append(msg, fmt.Sprintf("%v\t%q", ok, p))
	}
	cli.Out(msg...)
}

// Stages TODO(rjeczalik): document
func (cli *CLI) Stages(ctx *cli.Context) {
	cli.init(ctx)
	s, err := cli.c.Stages(cli.p.String())
	if err != nil {
		cli.Err(err)
	}
	msg := make([]interface{}, 0, len(s))
	for _, s := range s {
		msg = append(msg, s)
	}
	cli.Out(msg...)
}

// Build TODO(rjeczalik): document
func (cli *CLI) Build(ctx *cli.Context) {
	cli.init(ctx)
	reqID := ctx.Args().First()
	if reqID == "" {
		cli.Err("the request ID is missing")
	}
	id, err := cli.c.BuildID(reqID)
	if err != nil {
		cli.Err(err)
	}
	cli.Out(id)
}

// Login TODO(rjeczalik): document
func (cli *CLI) Login(ctx *cli.Context) {
	cli.init(ctx)
	old := []*string{&cli.cred.URL, &cli.cred.User, &cli.cred.Pass}
	for i, s := range []string{ctx.GlobalString("addr"), ctx.GlobalString("user"), ctx.GlobalString("pass")} {
		if s != "" {
			(*old[i]) = s
		}
	}
	cli.init(ctx)
	if err := cli.s.Save(cli.cred); err != nil {
		cli.Err(err)
	}
	cli.Out()
}

// Trigger TODO(rjeczalik): document
func (cli *CLI) Trigger(ctx *cli.Context) {
	cli.init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
	}
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		if err = cli.c.Clear(p); err != nil {
			cli.Err(err)
		}
		s, err := cli.c.Trigger(p)
		if err != nil {
			cli.Err(err)
		}
		for _, s := range s {
			msg = append(msg, fmt.Sprintf("%s\t%q", s, p))
		}
	}
	cli.Out(msg...)
}

// Health TODO(rjeczalik): document
func (cli *CLI) Health(ctx *cli.Context) {
	cli.init(ctx)
	a, err := cli.c.Agents()
	if err != nil {
		cli.Err(err)
	}
	if len(pulse.Filter(a, pulse.IsSync)) >= len(a)/2 {
		cli.Err("Pulse Agents are hanging again!")
	}
	if a = pulse.Filter(a, pulse.IsOffline); len(a) > 0 {
		msg := make([]interface{}, 0, len(a))
		for i := range a {
			msg = append(msg, a[i])
		}
		cli.Err(msg...)
	}
	cli.Out()
}

// Projects TODO(rjeczalik): document
func (cli *CLI) Projects(ctx *cli.Context) {
	cli.init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
	}
	msg := make([]interface{}, 0, len(p))
	for _, p := range p {
		msg = append(msg, p)
	}
	cli.Out(msg...)
}

// Agents TODO(rjeczalik): document
func (cli *CLI) Agents(ctx *cli.Context) {
	cli.init(ctx)
	a, err := cli.c.Agents()
	if err != nil {
		cli.Err(err)
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
	cli.init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Err(err)
	}
	m := make(map[string][]pulse.BuildResult)
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		var (
			b []pulse.BuildResult
			n = cli.n
		)
		if n > 0 {
			b, err = cli.c.BuildResult(p, n)
		} else {
			b, err = cli.c.LatestBuildResult(p)
			if err != nil {
				cli.Err(err)
			}
			var max int64
			for i := range b {
				if b[i].ID > max {
					max = b[i].ID
				}
			}
			if cli.n < 0 {
				b, err = cli.c.BuildResult(p, max+n)
			}
			n = max + n
		}
		if err != nil {
			cli.Err(err)
		}
		m[fmt.Sprintf("%s (build %d)", p, n)] = b
	}
	y, err := yaml.Marshal(m)
	if err != nil {
		cli.Err(err)
	}
	cli.Out(string(y))
}

// Run TODO(rjeczalik): document
func (cli *CLI) Run(args []string) {
	cli.app.Run(args)
}
