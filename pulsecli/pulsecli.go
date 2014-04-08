package pulsecli

import (
	"fmt"
	"regexp"

	"github.com/x-formation/int-tools/pulse"
	"github.com/x-formation/int-tools/pulse/prtg"

	"github.com/codegangsta/cli"
	"gopkg.in/v1/yaml"
)

// New TODO(rjeczalik): document
type CLI struct {
	Error func(...interface{})
	app   *cli.App
	c     pulse.Client
	a     *regexp.Regexp
	p     *regexp.Regexp
	n     int64
}

// New TODO(rjeczalik): document
func New() *CLI {
	cl := &CLI{
		Error: prtg.Error,
		app:   cli.NewApp(),
	}
	cl.app.Name, cl.app.Version = "pulsecli", "0.1.0"
	cl.app.Usage = "a command-line client for a Pulse server"
	cl.app.Flags = []cli.Flag{
		cli.StringFlag{"addr", "http://pulse/xmlrpc", "Pulse Remote API endpoint"},
		cli.StringFlag{"user", "", "Pulse user name"},
		cli.StringFlag{"pass", "", "Pulse user password"},
		cli.StringFlag{"agent, a", ".*", "Agent name patter"},
		cli.StringFlag{"project, p", ".*", "Project name pattern"},
		cli.IntFlag{"build, b", 0, "Build number"},
	}
	cl.app.Commands = []cli.Command{{
		Name:   "trigger",
		Usage:  "trigger a build",
		Action: cl.Trigger,
	}, {
		Name:   "health",
		Usage:  "perform a health check",
		Action: cl.Health,
	}, {
		Name:   "projects",
		Usage:  "list all projcts",
		Action: cl.Projects,
	}, {
		Name:   "agents",
		Usage:  "list all agents",
		Action: cl.Agents,
	}, {
		Name:   "status",
		Usage:  "list build status",
		Action: cl.Status,
	}}
	return cl
}

// Init TODO(rjeczalik): document
func (cli *CLI) Init(ctx *cli.Context) {
	var err error
	addr, user, pass := ctx.GlobalString("addr"), ctx.GlobalString("user"), ctx.GlobalString("pass")
	if cli.c, err = pulse.NewClient(addr, user, pass); err != nil {
		cli.Error(err)
	}
	a, p, n := ctx.GlobalString("agent"), ctx.GlobalString("project"), ctx.GlobalInt("build")
	if cli.a, err = regexp.Compile(a); err != nil {
		cli.Error(err)
	}
	if cli.p, err = regexp.Compile(p); err != nil {
		cli.Error(err)
	}
	cli.n = int64(n)
}

// Trigger TODO(rjeczalik): document
func (cli *CLI) Trigger(ctx *cli.Context) {
	cli.Init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Error(err)
	}
	for _, p := range p {
		if !cli.p.MatchString(p) {
			continue
		}
		if err = cli.c.Clear(p); err != nil {
			cli.Error(err)
		}
		if _, err = cli.c.Trigger(p); err != nil {
			cli.Error(err)
		}
	}
	prtg.OK()
}

// Health TODO(rjeczalik): document
func (cli *CLI) Health(ctx *cli.Context) {
	cli.Init(ctx)
	a, err := cli.c.Agents()
	if err != nil {
		cli.Error(err)
	}
	if len(pulse.Filter(a, pulse.IsSync)) >= len(a)/2 {
		cli.Error("Pulse Agents are hanging again!")
	}
	if a = pulse.Filter(a, pulse.IsOffline); len(a) > 0 {
		args := make([]interface{}, 0, len(a))
		for i := range a {
			args = append(args, a[i])
		}
		cli.Error(args...)
	}
	prtg.OK()
}

// Projects TODO(rjeczalik): document
func (cli *CLI) Projects(ctx *cli.Context) {
	cli.Init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Error(err)
	}
	for _, p := range p {
		fmt.Println(p)
	}
}

// Agents TODO(rjeczalik): document
func (cli *CLI) Agents(ctx *cli.Context) {
	cli.Init(ctx)
	a, err := cli.c.Agents()
	if err != nil {
		cli.Error(err)
	}
	for _, a := range a {
		fmt.Printf("%+v\n", a)
	}
}

// Status TODO(rjeczalik): document
func (cli *CLI) Status(ctx *cli.Context) {
	cli.Init(ctx)
	p, err := cli.c.Projects()
	if err != nil {
		cli.Error(err)
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
				cli.Error(err)
			}
			var max int64
			for i := range b {
				if b[i].ID > max {
					max = b[i].ID
				}
			}
			if cli.n < 0 {
				b, err = cli.c.BuildResult(p, max-n)
			}
			n = max - n
		}
		if err != nil {
			cli.Error(err)
		}
		m[fmt.Sprintf("%s (build %d)", p, n)] = b
	}
	y, err := yaml.Marshal(m)
	if err != nil {
		cli.Error(err)
	}
	fmt.Println(string(y))
}

// Run TODO(rjeczalik): document
func (cli *CLI) Run(args []string) {
	cli.app.Run(args)
}
