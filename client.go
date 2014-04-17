package pulse

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/kolo/xmlrpc"
)

// Client TODO(rjeczalik): document
type Client interface {
	Agents() (Agents, error)
	BuildID(reqid string) (int64, error)
	BuildResult(project string, id int64) ([]BuildResult, error)
	Clear(project string) error
	Close() error
	ConfigStage(project, stage string) (ProjectStage, error)
	Init(project string) (bool, error)
	LatestBuildResult(project string) ([]BuildResult, error)
	Messages(project string, id int64) (Messages, error)
	Projects() ([]string, error)
	Stages(project string) ([]string, error)
	SetTimeout(d time.Duration)
	SetConfigStage(project string, s ProjectStage) error
	Trigger(project string) ([]string, error)
}

var ErrTimeout = errors.New("pulse: request has timed out")

// InvalidBuildError TODO(rjeczalik): document
type InvalidBuildError struct {
	ID     int64
	Status BuildStatus
	ReqID  string
}

func (e InvalidBuildError) Error() string {
	return fmt.Sprintf("pulse: invalid build: id=%d, status=%s, reqid=%s", e.ID,
		e.Status, e.ReqID)
}

type client struct {
	rpc *xmlrpc.Client
	tok string
	d   time.Duration
}

// NewClient TODO(rjeczalik): document
func NewClient(url, user, pass string) (Client, error) {
	c, err := &client{d: 15 * time.Second}, (error)(nil)
	if c.rpc, err = xmlrpc.NewClient(url, nil); err != nil {
		return nil, err
	}
	if err = c.rpc.Call("RemoteApi.login", []interface{}{user, pass}, &c.tok); err != nil {
		return nil, err
	}
	return c, nil
}

// SetTimeout TODO(rjeczalik): document
func (c *client) SetTimeout(d time.Duration) { c.d = d }

// Init TODO(rjeczalik): document
func (c *client) Init(project string) (ok bool, err error) {
	err = c.rpc.Call("RemoteApi.initialiseProject", []interface{}{c.tok, project}, &ok)
	return
}

// Messages TODO(rjeczalik): document
func (c *client) Messages(project string, id int64) (Messages, error) {
	var (
		m, warn, info Messages
		req           = []interface{}{c.tok, project, int(id)}
	)
	if err := c.rpc.Call("RemoteApi.getErrorMessagesInBuild", req, &m); err != nil {
		return nil, err
	}
	if err := c.rpc.Call("RemoteApi.getWarningMessagesInBuild", req, &warn); err != nil {
		return nil, err
	}
	if err := c.rpc.Call("RemoteApi.getInfoMessagesInBuild", req, &info); err != nil {
		return nil, err
	}
	return append(append(m, warn...), info...), nil
}

// ConfigStage TODO(rjeczalik): document
func (c *client) ConfigStage(project, stage string) (s ProjectStage, err error) {
	req := []interface{}{c.tok, fmt.Sprintf("projects/%s/stages/%s", project, stage)}
	err = c.rpc.Call("RemoteApi.getConfig", req, &s)
	return
}

// SetConfigStage TODO(rjeczalik): document
func (c *client) SetConfigStage(project string, s ProjectStage) (err error) {
	req := []interface{}{c.tok, fmt.Sprintf("projects/%s/stages/%s", project, s.Name), &s, false}
	err = c.rpc.Call("RemoteApi.saveConfig", req, new(string))
	return
}

// BuildID TODO(rjeczalik): document
func (c *client) BuildID(reqid string) (int64, error) {
	timeout, rep := int(c.d.Seconds())*1000, &BuildRequestStatus{}
	err := c.rpc.Call("RemoteApi.waitForBuildRequestToBeActivated",
		[]interface{}{c.tok, reqid, timeout}, &rep)
	if err != nil {
		return 0, err
	}
	if rep.Status == BuildUnknown {
		return 0, &InvalidBuildError{Status: BuildUnknown, ReqID: reqid}
	}
	if rep.Status == BuildUnhandled || rep.Status == BuildQueued {
		return 0, ErrTimeout
	}
	return strconv.ParseInt(rep.ID, 10, 64)
}

// BuildResult TODO(rjeczalik): document
func (c *client) BuildResult(project string, id int64) (res []BuildResult, err error) {
	if project == ProjectPersonal {
		err = c.rpc.Call("RemoteApi.getPersonalBuild", []interface{}{c.tok, int(id)}, &res)
	} else {
		err = c.rpc.Call("RemoteApi.getBuild", []interface{}{c.tok, project, int(id)}, &res)
	}
	if err != nil {
		return nil, err
	}
	if res == nil || len(res) == 0 {
		return nil, &InvalidBuildError{ID: id, Status: BuildUnknown}
	}
	return res, nil
}

// Stages TODO(rjeczalik): document
func (c *client) Stages(project string) ([]string, error) {
	// TODO(rjeczalik): It would be better to get stages list from project's configuration.
	//                  I ran away screaming while trying to get that information from
	//                  the Remote API spec.
	b, err := c.LatestBuildResult(project)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("pulse: error requesting latest build status")
	}
	if len(b[0].Stages) == 0 {
		return nil, errors.New("pulse: stage list is empty")
	}
	s := make([]string, 0, len(b[0].Stages))
	for i := range b[0].Stages {
		s = append(s, b[0].Stages[i].Name)
	}
	return s, nil
}

// LatestBuildResult TODO(rjeczalik): document
func (c *client) LatestBuildResult(project string) (res []BuildResult, err error) {
	if project == ProjectPersonal {
		err = c.rpc.Call("RemoteApi.getLatestPersonalBuildForProject", []interface{}{c.tok, true}, &res)
	} else {
		err = c.rpc.Call("RemoteApi.getLatestBuildForProject", []interface{}{c.tok, project, true}, &res)
	}
	if err != nil {
		return nil, err
	}
	if res == nil || len(res) == 0 {
		return nil, &InvalidBuildError{Status: BuildUnknown}
	}
	return res, nil
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
func (c *client) Agents() (Agents, error) {
	var names []string
	if err := c.rpc.Call("RemoteApi.getAllAgentNames", c.tok, &names); err != nil {
		return nil, err
	}
	a := make(Agents, len(names))
	for i := range names {
		if err := c.rpc.Call("RemoteApi.getAgentDetails", []interface{}{c.tok, names[i]}, &a[i]); err != nil {
			return nil, err
		}
		a[i].Name = names[i]
	}
	return a, nil
}
