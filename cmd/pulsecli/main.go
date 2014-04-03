package main

import (
	"flag"
	"regexp"

	"github.com/x-formation/int-tools/pulse"
	"github.com/x-formation/int-tools/pulse/prtg"
)

var (
	cli     pulse.Client
	trigger *regexp.Regexp
	health  bool
)

func init() {
	var err error
	url := flag.String("url", "http://pulse", "Pulse Remote API endpoint")
	user := flag.String("user", "", "Pulse Remote user name")
	pass := flag.String("pass", "", "Pulse Remote user password")
	trig := flag.String("trigger", "", "Triggers projects that matches the pattern")
	flag.BoolVar(&health, "health", false, "Performs a health check against Pulse server")
	flag.Parse()
	if health && *trig != "" {
		prtg.Error("-health/-trigger: the flags are exclusive!")
	}
	if !health && *trig == "" {
		prtg.Error("-health/-trigger: at least one flag required!")
	}
	if *trig != "" {
		if trigger, err = regexp.Compile(*trig); err != nil {
			prtg.Error(err)
		}
	}
	if cli, err = pulse.NewClient(*url+"/xmlrpc", *user, *pass); err != nil {
		prtg.Error(err)
	}
}

// TODO main should be refactored into separate file and tested.
func main() {
	var (
		projects []string
		agents   []pulse.Agent
		err      error
	)
	switch {
	case health:
		if agents, err = cli.Agents(); err != nil {
			prtg.Error(err)
		}
		if len(pulse.Filter(agents, pulse.IsSync)) >= len(agents)/2 {
			prtg.Error("Pulse Agents are hanging again!")
		}
		if agents = pulse.Filter(agents, pulse.IsOffline); len(agents) > 0 {
			args := make([]interface{}, 0, len(agents))
			for i := range agents {
				args = append(args, agents[i])
			}
			prtg.Error(args...)
		}
	case trigger != nil:
		if projects, err = cli.Projects(); err != nil {
			prtg.Error(err)
		}
		for _, project := range projects {
			if !trigger.MatchString(project) {
				continue
			}
			if err = cli.Clear(project); err != nil {
				prtg.Error(err)
			}
			if _, err = cli.Trigger(project); err != nil {
				prtg.Error(err)
			}
		}
	}
	prtg.OK()
}
