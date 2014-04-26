package dev

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/x-formation/pulsekit"
)

// Personal TODO(rjeczalik): document
type Personal struct {
	Patch    string
	Project  string
	Stages   []string
	Revision string
}

var ErrTimeout = errors.New("pulsedev: waiting for command to finish has timed out")

// Exec TODO(rjeczalik): document
type Exec interface {
	// LookPath TODO(rjeczalik): document
	LookPath(string) (string, error)
	// CombinedOutput TODO(rjeczalik): document
	CombinedOutput(string, []string) (io.Writer, io.Reader, func(time.Duration) error, error)
}

type cmdExec struct{}

func (cmdExec) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (cmdExec) CombinedOutput(file string, args []string) (stdin io.Writer,
	stdout io.Reader, wait func(time.Duration) error, err error) {
	cmd := exec.Command(file, args...)
	if stdin, err = cmd.StdinPipe(); err != nil {
		return
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	stdout = io.MultiReader(stdout, stderr)
	if err = cmd.Start(); err != nil {
		return
	}
	wait = func(d time.Duration) (err error) {
		done := make(chan error)
		go func() {
			done <- cmd.Wait()
			close(done)
		}()
		select {
		case err, _ = <-done:
		case <-time.After(d):
			err = ErrTimeout
		}
		return
	}
	return
}

// DefaultExec TODO(rjeczalik): document
var DefaultExec Exec = new(cmdExec)

// Tool TODO(rjeczalik): document
type Tool interface {
	// TODO(rjeczalik): document
	Personal(p *Personal) (int64, error)
	SetTimeout(d time.Duration)
}

type tool struct {
	exec Exec
	c    pulse.Client
	url  string
	user string
	pass string
	exe  string
	d    time.Duration
}

func New(c pulse.Client, url, user, pass string) (Tool, error) {
	var err error
	t := &tool{
		exec: DefaultExec,
		c:    c,
		url:  url,
		user: user,
		pass: pass,
		d:    15 * time.Second,
	}
	if t.exe, err = t.exec.LookPath("pulse"); err != nil {
		return nil, err
	}
	if err = t.valid(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *tool) valid() error {
	for _, args := range [][]string{{"help"}, {"help", "personal"}} {
		_, _, wait, err := t.exec.CombinedOutput(t.exe, args)
		if err != nil {
			return err
		}
		if err = wait(t.d); err != nil {
			return err
		}
	}
	return nil
}

func (t *tool) Personal(p *Personal) (id int64, err error) {
	var (
		line   []byte
		stages []pulse.ProjectStage
		delim  = byte('\n')
		rev    = false
	)
	args := []string{
		"personal",
		"-s", t.url,
		"-u", t.user,
		"-p", t.pass,
		"-r", p.Project,
		"-t", "git",
		"-f", p.Patch,
	}
	if p.Revision != "" {
		args = append(args, "-e", p.Revision)
	}
	if p.Stages != nil {
		stages = make([]pulse.ProjectStage, len(p.Stages))
		for i, s := range p.Stages {
			stages[i], err = t.c.ConfigStage(p.Project, s)
			if err != nil {
				return
			}
			stages[i].Enabled = false
			if err = t.c.SetConfigStage(p.Project, stages[i]); err != nil {
				return
			}
		}
	}
	in, out, wait, err := t.exec.CombinedOutput(t.exe, args)
	if err != nil {
		return
	}
	r := bufio.NewReader(out)
INTERACTIVE:
	for {
		line, err = r.ReadBytes(delim)
		if err == io.EOF {
			continue
		}
		if err != nil {
			break INTERACTIVE
		}
		s := string(line)
		switch delim {
		case '\n':
			if strings.Contains(s, "Continue anyway?") || strings.Contains(s, "Synchronise now?") {
				delim = '>'
			}
			if strings.Contains(s, "Choose revision to build against") {
				delim, rev = '>', true
			}
			i := strings.Index(s, "Patch accepted: personal build")
			if i != -1 {
				id, err = strconv.ParseInt(s[31+i:len(s)-2], 10, 64)
				break INTERACTIVE
			}
			if strings.Contains(s, "Error") || strings.Contains(s, "Exception") {
				err = errors.New(s)
				break INTERACTIVE
			}
		case '>':
			res := []byte("Yes\n")
			if rev {
				res = []byte("1!\n")
			}
			if _, err = in.Write(res); err != nil && err != io.EOF {
				break INTERACTIVE
			}
			delim, rev = '\n', false
		}
	}
	if e := wait(t.d); (err == io.EOF || err == nil) && e != nil {
		err = e
	}
	if stages != nil {
		for i := range stages {
			stages[i].Enabled = true
			if e := t.c.SetConfigStage(p.Project, stages[i]); e != nil && err == nil {
				err = e
			}
		}
	}
	return
}

func (t *tool) SetTimeout(d time.Duration) {
	t.d = d
}
