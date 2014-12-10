package pulse

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/kolo/xmlrpc"
)

// Client is a RPC client for talking with Pulse Remote API endpoint.
// It is expected that Client holds valid user session, which can be
// terminated by a call to Close method.
type Client interface {
	// Agents returns every machine registred with Pulse server that the user
	// holding the session has an access to.
	Agents() (Agents, error)
	// BuildID gives a build ID associated with given request ID. If a build
	// is queued and not started yet it waits up to 15 seconds before timing out.
	BuildID(reqid string) (int64, error)
	// BuildResults gives full statistics and information for a build with given
	// ID and project name.
	BuildResult(project string, id int64) ([]BuildResult, error)
	// Clear clears a working directories on agents for a given project name.
	Clear(project string) error
	// Close terminates the user session.
	Close() error
	// ConfigStage TODO(rjeczalik): document
	ConfigStage(project, stage string) (ProjectStage, error)
	// Init (re-)initializes the project with a given name. It stops the SCM polling,
	// clears Pulse server's local clone of a repository, configured for
	// a given project, and checks it out again.
	Init(project string) (bool, error)
	// LastestBuildResult returns statistics for a latest completed build of
	// a given project.
	LatestBuildResult(project string) ([]BuildResult, error)
	// Messages returns all info, warning and error messages for a particular
	// build of a given project.
	Messages(project string, id int64) (Messages, error)
	// Projects gives every project name that the user holding the session
	// has an access to.
	Projects() ([]string, error)
	// Stages gives every stage name for a given project.
	Stages(project string) ([]string, error)
	// SetTimeout TODO(rjeczalik): document
	SetTimeout(d time.Duration)
	// SetConfigStage TODO(rjeczalik): document
	SetConfigStage(project string, s ProjectStage) error
	// Trigger triggers a build for a given project returning request IDs
	// of builds caused by that trigger.
	Trigger(project string) ([]string, error)
	// Artifact downloads artifacts for given project and build number
	Artifact(id int64, project, dir, url string) error
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

// NewClient authenticates with Pulse server for a user session, creating
// a RPC client.
func NewClient(url, user, pass string) (Client, error) {
	c, err := &client{d: 15 * time.Second}, (error)(nil)
	if c.rpc, err = xmlrpc.NewClient(url+"/xmlrpc", nil); err != nil {
		return nil, err
	}
	if err = c.rpc.Call("RemoteApi.login", []interface{}{user, pass}, &c.tok); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *client) SetTimeout(d time.Duration) { c.d = d }

func (c *client) Init(project string) (ok bool, err error) {
	err = c.rpc.Call("RemoteApi.initialiseProject", []interface{}{c.tok, project}, &ok)
	return
}

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

func (c *client) ConfigStage(project, stage string) (s ProjectStage, err error) {
	req := []interface{}{c.tok, fmt.Sprintf("projects/%s/stages/%s", project, stage)}
	err = c.rpc.Call("RemoteApi.getConfig", req, &s)
	return
}

func (c *client) SetConfigStage(project string, s ProjectStage) (err error) {
	req := []interface{}{c.tok, fmt.Sprintf("projects/%s/stages/%s", project, s.Name), &s, false}
	err = c.rpc.Call("RemoteApi.saveConfig", req, new(string))
	return
}

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
		return nil, &InvalidBuildError{ID: id, Status: BuildNeverBuilt}
	}
	return res, nil
}

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
		return nil, &InvalidBuildError{Status: BuildNeverBuilt}
	}
	return res, nil
}

func (c *client) Close() error {
	if err := c.rpc.Call("RemoteApi.logout", c.tok, nil); err != nil {
		return err
	}
	return c.rpc.Close()
}

func (c *client) Clear(project string) error {
	return c.rpc.Call("RemoteApi.doConfigAction", []interface{}{c.tok, "projects/" + project, "clean"}, nil)
}

func (c *client) Trigger(project string) (id []string, err error) {
	// TODO(rjeczalik): Use TriggerOptions struct instead after kolo/xmlrpc
	//                  supports maps.
	req := struct {
		R bool `xmlrpc:"rebuild"`
	}{true}
	err = c.rpc.Call("RemoteApi.triggerBuild", []interface{}{c.tok, project, req}, &id)
	return
}

func (c *client) Projects() (s []string, err error) {
	err = c.rpc.Call("RemoteApi.getAllProjectNames", c.tok, &s)
	return
}

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

// Artifact(project string, id int64, dir string, baseUrl string)
func (c *client) Artifact(id int64, project, dir, url string) (err error) {
	var art []BuildArtifact
	if project == ProjectPersonal {
		err = c.rpc.Call("RemoteApi.getArtifactsInPersonalBuild", []interface{}{c.tok, int(id)}, &art)
	} else {
		err = c.rpc.Call("RemoteApi.getArtifactsInBuild", []interface{}{c.tok, project, int(id)}, &art)
	}
	if err != nil {
		return err
	}

	for i := range art {
		if project == ProjectPersonal {
			err = c.rpc.Call("RemoteApi.getArtifactFileListingPersonal", []interface{}{c.tok, int(id), art[i].Stage, art[i].Command, art[i].Name, ""},
				&art[i].Files)
		} else {
			err = c.rpc.Call("RemoteApi.getArtifactFileListing", []interface{}{c.tok, project, int(id), art[i].Stage, art[i].Command, art[i].Name, ""},
				&art[i].Files)
		}
		if err != nil {
			return err
		}
	}

	af := NewArtifactFetcher(url, c.tok, dir)
	for i := range art {
		if err = af.Fetch(&art[i], project); err != nil {
			return err
		}
	}
	return nil
}
