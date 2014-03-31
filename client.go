package pulse

import "github.com/kolo/xmlrpc"

// AgentStatus TODO(rjeczalik): document
type AgentStatus string

const (
	AgentOffline  AgentStatus = "offline"
	AgentSync     AgentStatus = "Synchronizing"
	AgentIdle     AgentStatus = "idle"
	AgentBuilding AgentStatus = "building"
	AgentDisabled AgentStatus = "disabled"
)

// Agent TODO(rjeczalik): document
type Agent struct {
	Name   string
	Status AgentStatus `xmlrpc:"status"`
	Host   string      `xmlrpc:"location"`
}

// String TODO(rjeczalik): document
func (a Agent) String() string { return a.Name + "@" + a.Host }

// Client TODO(rjeczalik): document
type Client interface {
	Close() error
	Clear(project string) error
	Trigger(project string) error
	Projects() ([]string, error)
	Agents() ([]Agent, error)
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
func (c *client) Trigger(project string) error {
	return c.rpc.Call("RemoteApi.triggerBuild", []interface{}{c.tok, project}, nil)
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
