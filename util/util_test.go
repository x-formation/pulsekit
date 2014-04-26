package util

import (
	"errors"
	"fmt"
	"testing"

	"github.com/x-formation/pulsekit"
	"github.com/x-formation/pulsekit/mock"
)

func TestPending(t *testing.T) {
	pending := []interface{}{
		&pulse.StageResult{Agent: pulse.AgentPending},
		&[]pulse.StageResult{{}, {}, {Agent: pulse.AgentPending}},
		&pulse.BuildResult{
			Stages: []pulse.StageResult{{}, {}, {Agent: pulse.AgentPending}},
		},
		&[]pulse.BuildResult{
			{Stages: []pulse.StageResult{{}, {}, {}}},
			{Stages: []pulse.StageResult{{}, {}, {}}},
			{Stages: []pulse.StageResult{{}, {}, {Agent: pulse.AgentPending}}},
		},
	}
	notpending := []interface{}{
		&pulse.StageResult{},
		&[]pulse.StageResult{{}, {}, {}},
		&pulse.BuildResult{
			Stages: []pulse.StageResult{{}, {}, {}},
		},
		&[]pulse.BuildResult{
			{Stages: []pulse.StageResult{{}, {}, {}}},
			{Stages: []pulse.StageResult{{}, {}, {}}},
			{Stages: []pulse.StageResult{{}, {}, {}}},
		},
	}
	for i := 0; i < 4; i++ {
		if !Pending(pending[i]) {
			t.Errorf("expected %+v to be pending (i=%d)", pending[i], i)
		}
		if Pending(notpending[i]) {
			t.Errorf("expected %+v to be not pending (i=%d)", notpending[i], i)
		}
	}
}

func TestWaitOK(t *testing.T) {
	mc := mock.NewClient()
	mc.Err = []error{nil}
	mc.BR = []pulse.BuildResult{
		{Complete: true},
		{Complete: true},
	}
	if _, ok := <-Wait(mc, 0, "LM-X - Tier 1", 1024); ok {
		t.Error("expected ok to be false")
	}
	mc.Check(t)
}

func TestWaitErr(t *testing.T) {
	mc := mock.NewClient()
	mc.Err = []error{errors.New("err")}
	if err, ok := <-Wait(mc, 0, "LM-X - Tier 1", 1024); err == nil || !ok {
		t.Errorf("expected err!=nil and ok=true, was err=%q, ok=%v", err, ok)
	}
	mc.Check(t)
}

var errInvalidBuild = &pulse.InvalidBuildError{Status: pulse.BuildUnknown}

type fixture struct {
	BI          int64
	L           []pulse.BuildResult
	BR          []pulse.BuildResult
	Err         []error
	ID          int64
	ExpectedID  int64
	ExpectedErr error
}

func (f *fixture) MockClient() *mock.Client {
	mc := mock.NewClient()
	mc.BI, mc.L, mc.BR, mc.Err = f.BI, f.L, f.BR, f.Err
	return mc
}

func (f *fixture) String() string {
	return fmt.Sprintf("reqid=%d, id=%d, expected=%d", f.BI, f.ID, f.ExpectedID)
}

func check(t *testing.T, err, expected error) {
	if expected != nil && err == nil {
		t.Error("expected err to be non-nil")
	}
	if expected == nil && err != nil {
		t.Errorf("expected err to be nil, was %q instead", err)
	}
}

func TestNormalizeBuildID(t *testing.T) {
	f := []fixture{{
		ID:         1,
		ExpectedID: 1,
	}, {
		ID:         0,
		L:          []pulse.BuildResult{{ID: 12}, {ID: 15}},
		Err:        []error{nil},
		ExpectedID: 15,
	}, {
		ID:          0,
		Err:         []error{pulse.ErrTimeout},
		ExpectedID:  0,
		ExpectedErr: pulse.ErrTimeout,
	}, {
		ID:         -10,
		L:          []pulse.BuildResult{{ID: 12}, {ID: 15}},
		Err:        []error{nil},
		ExpectedID: 5,
	}, {
		ID:          -20,
		L:           []pulse.BuildResult{{ID: 12}, {ID: 15}},
		Err:         []error{nil},
		ExpectedID:  0,
		ExpectedErr: errInvalidBuild,
	}, {
		ID:          -20,
		Err:         []error{pulse.ErrTimeout},
		ExpectedID:  0,
		ExpectedErr: pulse.ErrTimeout,
	}}
	for _, f := range f {
		mc := f.MockClient()
		id, err := NormalizeBuildID(mc, "License Statistics", f.ID)
		if id != f.ExpectedID {
			t.Errorf("expected id to be %d, was %d instead (%s)", f.ExpectedID, id, f.String())
		}
		check(t, err, f.ExpectedErr)
		mc.Check(t)
	}
}

func TestNormalizeBuildOrRequestId(t *testing.T) {
	f := []fixture{{
		ID:         2,
		Err:        []error{nil},
		ExpectedID: 2,
	}, {
		ID:         2,
		BI:         10,
		Err:        []error{errInvalidBuild, nil},
		ExpectedID: 10,
	}, {
		ID:          2,
		Err:         []error{errInvalidBuild, errInvalidBuild},
		ExpectedID:  0,
		ExpectedErr: errInvalidBuild,
	}, {
		ID:          2,
		Err:         []error{pulse.ErrTimeout},
		ExpectedID:  0,
		ExpectedErr: pulse.ErrTimeout,
	}}
	for _, f := range f {
		mc := f.MockClient()
		id, err := NormalizeBuildOrRequestID(mc, "License Statistics", f.ID)
		if id != f.ExpectedID {
			t.Errorf("expected id to be %d, was %d instead (%s)", f.ExpectedID, id, f.String())
		}
		check(t, err, f.ExpectedErr)
		mc.Check(t)
	}

}
