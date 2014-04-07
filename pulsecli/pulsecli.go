package pulsecli

import (
	"fmt"
	"regexp"

	"github.com/x-formation/int-tools/pulse"
	"github.com/x-formation/int-tools/pulse/prtg"

	"github.com/codegangsta/cli"
)

// New TODO(rjeczalik): document
type CLI struct {
	Error func(...interface{})
	app   *cli.App
	c     pulse.Client
	a     *regexp.Regexp
	p     *regexp.Regexp
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
	a, p := ctx.GlobalString("agent"), ctx.GlobalString("project")
	if cli.a, err = regexp.Compile(a); err != nil {
		cli.Error(err)
	}
	if cli.p, err = regexp.Compile(p); err != nil {
		cli.Error(err)
	}
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

// Run TODO(rjeczalik): document
func (cli *CLI) Run(args []string) {
	cli.app.Run(args)
}
