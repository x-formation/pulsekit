package pulse

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestAgentsFilter(t *testing.T) {
	agents := []Agents{
		{Agent{Status: AgentIdle}, Agent{Status: AgentIdle}, Agent{Status: AgentIdle}},
		{Agent{Status: AgentIdle}, Agent{Status: AgentOffline}, Agent{Status: AgentBuilding}},
		{Agent{Status: AgentOffline}, Agent{Status: AgentOffline}, Agent{Status: AgentOffline}},
		{Agent{Status: AgentBuilding}, Agent{Status: AgentBuilding}, Agent{Status: AgentDisabled}},
		{Agent{Status: AgentSync}, Agent{Status: AgentIdle}, Agent{Status: AgentDisabled}},
		{Agent{Status: AgentSync}, Agent{Status: AgentSync}, Agent{Status: AgentSync}},
		{Agent{Host: "win-1"}, Agent{Host: "win-2"}, Agent{Host: "win-3"}},
		{Agent{Host: "solx86-1"}, Agent{Host: "macosx9-1"}, Agent{Host: "solx86-2"}},
		{Agent{Host: "win2012test-1"}, Agent{Host: "pulse-x64-1"}, Agent{Host: "pulse-x86-2"}},
	}
	filters := []func(*Agent) bool{
		Offline, Offline, Offline, Sync, Sync, Sync,
		func(a *Agent) bool { return strings.HasPrefix(a.Host, "win-") },
		func(a *Agent) bool { return strings.HasPrefix(a.Host, "dev-") },
		func(a *Agent) bool { return strings.HasPrefix(a.Host, "pulse-") },
	}
	expected := []Agents{
		{},
		{Agent{Status: AgentOffline}},
		{Agent{Status: AgentOffline}, Agent{Status: AgentOffline}, Agent{Status: AgentOffline}},
		{},
		{Agent{Status: AgentSync}},
		{Agent{Status: AgentSync}, Agent{Status: AgentSync}, Agent{Status: AgentSync}},
		{Agent{Host: "win-1"}, Agent{Host: "win-2"}, Agent{Host: "win-3"}},
		{},
		{Agent{Host: "pulse-x64-1"}, Agent{Host: "pulse-x86-2"}},
	}
	for i := range agents {
		filtered := agents[i].Filter(filters[i])
		if !reflect.DeepEqual(filtered, expected[i]) {
			t.Errorf("expected filtered to be equal %v, was %v instead (i=%d)",
				expected[i], filtered, i)
		}
	}
}

func TestPending(t *testing.T) {
	pending := []interface{}{
		&StageResult{Agent: AgentPending},
		&[]StageResult{{}, {}, {Agent: AgentPending}},
		&BuildResult{
			Stages: []StageResult{{}, {}, {Agent: AgentPending}},
		},
		&[]BuildResult{
			{Stages: []StageResult{{}, {}, {}}},
			{Stages: []StageResult{{}, {}, {}}},
			{Stages: []StageResult{{}, {}, {Agent: AgentPending}}},
		},
	}
	notpending := []interface{}{
		&StageResult{},
		&[]StageResult{{}, {}, {}},
		&BuildResult{
			Stages: []StageResult{{}, {}, {}},
		},
		&[]BuildResult{
			{Stages: []StageResult{{}, {}, {}}},
			{Stages: []StageResult{{}, {}, {}}},
			{Stages: []StageResult{{}, {}, {}}},
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
