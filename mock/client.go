package mock

import (
	"testing"
	"time"

	"github.com/x-formation/int-tools/pulseutil"
)

type Client struct {
	Err []error
	A   pulse.Agents
	BI  int64
	BR  []pulse.BuildResult
	I   bool
	L   []pulse.BuildResult
	M   pulse.Messages
	P   []string
	S   []string
	T   []string
	D   time.Duration
	i   int
}

func (c *Client) Check(t *testing.T) {
	if c.i != len(c.Err) {
		t.Errorf("mock: expected to be called %d times, was called %d times instead",
			len(c.Err), c.i)
	}
}

func (c *Client) err() error {
	i := c.i
	c.i++
	if c.Err == nil || len(c.Err) <= i {
		return nil
	}
	return c.Err[i]
}

func (c *Client) Agents() (pulse.Agents, error) {
	return c.A, c.err()
}

func (c *Client) BuildID(reqid string) (int64, error) {
	return c.BI, c.err()
}

func (c *Client) BuildResult(project string, id int64) ([]pulse.BuildResult, error) {
	return c.BR, c.err()
}

func (c *Client) Clear(project string) error {
	return c.err()
}

func (c *Client) Close() error {
	return c.err()
}

func (c *Client) Init(project string) (bool, error) {
	return c.I, c.err()
}

func (c *Client) LatestBuildResult(project string) ([]pulse.BuildResult, error) {
	return c.L, c.err()
}

func (c *Client) Messages(project string, id int64) (pulse.Messages, error) {
	return c.M, c.err()
}

func (c *Client) Projects() ([]string, error) {
	return c.P, c.err()
}

func (c *Client) SetTimeout(d time.Duration) {
	c.D = d
}

func (c *Client) Stages(project string) ([]string, error) {
	return c.S, c.err()
}

func (c *Client) Trigger(project string) ([]string, error) {
	return c.T, c.err()
}

func NewClient() *Client {
	return &Client{}
}
