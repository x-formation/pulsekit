package pulse

import (
	"errors"
	"strconv"
	"time"

	"github.com/kolo/xmlrpc"
)

// Client TODO(rjeczalik): document
type Client interface {
	Agents() ([]Agent, error)
	BuildID(reqid string) (int64, error)
	BuildResult(project string, id int64) ([]BuildResult, error)
	Clear(project string) error
	Close() error
	Projects() ([]string, error)
	Trigger(project string) ([]string, error)
	WaitBuild(project string, id int64) <-chan struct{}
}

type client struct {
	rpc *xmlrpc.Client
	tok string
}

// NewClient TODO(rjeczalik): document
func NewClient(url, user, pass string) (Client, error) {
	c, err := &client{}, (error)(nil)
	if c.rpc, err = xmlrpc.NewClient(url, nil); err != nil {
		return nil, err
	}
	if err = c.rpc.Call("RemoteApi.login", []interface{}{user, pass}, &c.tok); err != nil {
		return nil, err
	}
	return c, nil
}

// BuildID TODO(rjeczalik): document
func (c *client) BuildID(reqid string) (int64, error) {
	// TODO(rjeczalik): Extend Client interface with SetDeadline() method which
	//                  will configure both timeouts - for Pulse API and for
	//                  *rpc.Client from net/rpc.
	timeout, rep := 15*1000, &BuildRequestStatus{}
	err := c.rpc.Call("RemoteApi.waitForBuildRequestToBeActivated",
		[]interface{}{c.tok, reqid, timeout}, &rep)
	if err != nil {
		return 0, err
	}
	if rep.Status == BuildUnhandled || rep.Status == BuildQueued {
		return 0, errors.New("pulse: requesting build ID has timed out")
	}
	return strconv.ParseInt(rep.ID, 10, 64)
}

// BuildResult TODO(rjeczalik): document
func (c *client) BuildResult(project string, id int64) ([]BuildResult, error) {
	var build []BuildResult
	err := c.rpc.Call("RemoteApi.getBuild", []interface{}{c.tok, project, int(id)}, &build)
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (c *client) WaitBuild(project string, id int64) <-chan struct{} {
	done, sleep := make(chan struct{}), 250*time.Millisecond
	go func() {
		build, retry := make([]BuildResult, 0), 3
	WaitLoop:
		for {
			build = build[:0]
			err := c.rpc.Call("RemoteApi.getBuild", []interface{}{c.tok, project,
				int(id)}, &build)
			if err != nil {
				if retry > 0 {
					retry -= 1
					time.Sleep(sleep)
					continue WaitLoop
				}
				close(done)
				return
			}
			for i := range build {
				if !build[i].Complete {
					time.Sleep(sleep)
					continue WaitLoop
				}
			}
			close(done)
			return
		}
	}()
	return done
}

// Close TODO(rjeczalik): document
func (c *client) Close() error {
	if err := c.rpc.Call("RemoteApi.logout", c.tok, nil); err != nil {
		return err
	}
	return c.rpc.Close()
}

// Clear TODO(rjeczalik): document
func (c *client) Clear(project string) error {
	return c.rpc.Call("RemoteApi.doConfigAction", []interface{}{c.tok, "projects/" + project, "clean"}, nil)
}

// Trigger TODO(rjeczalik): document
func (c *client) Trigger(project string) (id []string, err error) {
	// TODO(rjeczalik): Use TriggerOptions struct instead after kolo/xmlrpc
	//                  supports maps.
	req := struct {
		R bool `xmlrpc:"rebuild"`
	}{true}
	err = c.rpc.Call("RemoteApi.triggerBuild", []interface{}{c.tok, project, req}, &id)
	return
}

// Projects TODO(rjeczalik): document
func (c *client) Projects() (s []string, err error) {
	err = c.rpc.Call("RemoteApi.getAllProjectNames", c.tok, &s)
	return
}

// Agents TODO(rjeczalik): document
func (c *client) Agents() ([]Agent, error) {
	var names []string
	if err := c.rpc.Call("RemoteApi.getAllAgentNames", c.tok, &names); err != nil {
		return nil, err
	}
	agents := make([]Agent, len(names))
	for i := range names {
		if err := c.rpc.Call("RemoteApi.getAgentDetails", []interface{}{c.tok, names[i]}, &agents[i]); err != nil {
			return nil, err
		}
		agents[i].Name = names[i]
	}
	return agents, nil
}
