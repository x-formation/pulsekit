package util

import (
	"strconv"
	"time"

	"github.com/x-formation/pulsekit"
)

// Pending TODO(rjeczalik): document
func Pending(v interface{}) bool {
	switch v := v.(type) {
	case *pulse.BuildResult:
		return Pending(&v.Stages)
	case *[]pulse.BuildResult:
		for i := range *v {
			if Pending(&(*v)[i].Stages) {
				return true
			}
		}
	case *pulse.StageResult:
		return v.Agent == pulse.AgentPending
	case *[]pulse.StageResult:
		for i := range *v {
			if Pending(&(*v)[i]) {
				return true
			}
		}
	}
	return false
}

// Wait TODO(rjeczalik): document
func Wait(c pulse.Client, d time.Duration, project string, id int64) <-chan error {
	done := make(chan error)
	go func() {
	WaitLoop:
		for {
			b, err := c.BuildResult(project, id)
			if err != nil {
				done <- err
				close(done)
				return
			}
			for i := range b {
				if !b[i].Complete {
					time.Sleep(d)
					continue WaitLoop
				}
			}
			close(done)
			return
		}
	}()
	return done
}

// NormalizeBuildID TODO(rjeczalik): document
func NormalizeBuildID(c pulse.Client, p string, id int64) (int64, error) {
	// Regular build ID.
	if id > 0 {
		return id, nil
	}
	// Special build ID:
	//  * 0 means the latest build
	//  * < 0 means relative offset to the latest build
	b, err := c.LatestBuildResult(p)
	if err != nil {
		return 0, err
	}
	var max int64
	for i := range b {
		if b[i].ID > max {
			max = b[i].ID
		}
	}
	if max+id <= 0 {
		return 0, &pulse.InvalidBuildError{ID: id, Status: pulse.BuildUnknown}
	}
	return max + id, nil
}

// NormalizeBuildOrRequestID TODO(rjeczalik): document
func NormalizeBuildOrRequestID(c pulse.Client, p string, reqorid int64) (int64, error) {
	// In case reqorid is relative build offset.
	id, err := NormalizeBuildID(c, p, reqorid)
	if err != nil {
		if _, ok := err.(*pulse.InvalidBuildError); !ok {
			return 0, err
		}
    if err.(*pulse.InvalidBuildError).Status == pulse.BuildNeverBuilt {
      return 0, err
    }
		id = reqorid
	}
	// In case reqorid is request ID.
	_, err = c.BuildResult(p, id)
	if err == nil {
		return id, nil
	}
	if _, ok := err.(*pulse.InvalidBuildError); !ok {
		return 0, err
	}
	// Get build ID from the request ID one.
	if id, err = c.BuildID(strconv.FormatInt(id, 10)); err != nil {
		return 0, err
	}
	return id, nil
}
