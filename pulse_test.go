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
		a := agents[i].Filter(filters[i])
		if !reflect.DeepEqual(a, expected[i]) {
			t.Errorf("expected a to be equal %v, was %v instead (i=%d)",
				expected[i], a, i)
		}
	}
}

func TestMessagesFilter(t *testing.T) {
	messages := []Messages{
		{Message{Severity: SeverityInfo}, Message{Severity: SeverityInfo}, Message{Severity: SeverityInfo}},
		{Message{Severity: SeverityWarning}, Message{Severity: SeverityInfo}, Message{Severity: SeverityError}},
		{Message{Severity: SeverityWarning}, Message{Severity: SeverityWarning}, Message{Severity: SeverityError}},
		{Message{Severity: SeverityWarning}, Message{Severity: SeverityWarning}, Message{Severity: SeverityWarning}},
		{Message{Severity: SeverityError}, Message{Severity: SeverityWarning}, Message{Severity: SeverityInfo}},
		{Message{Severity: SeverityError}, Message{Severity: SeverityError}, Message{Severity: SeverityInfo}},
		{Message{Severity: SeverityError}, Message{Severity: SeverityError}, Message{Severity: SeverityError}},
		{Message{Severity: SeverityInfo}, Message{Severity: SeverityError}, Message{Severity: SeverityWarning}},
		{Message{Severity: SeverityInfo}, Message{Severity: SeverityInfo}, Message{Severity: SeverityWarning}},
	}
	filters := []func(*Message) bool{
		Info, Info, Info,
		Warning, Warning, Warning,
		Error, Error, Error,
	}
	expected := []Messages{
		{Message{Severity: SeverityInfo}, Message{Severity: SeverityInfo}, Message{Severity: SeverityInfo}},
		{Message{Severity: SeverityInfo}},
		{},
		{Message{Severity: SeverityWarning}, Message{Severity: SeverityWarning}, Message{Severity: SeverityWarning}},
		{Message{Severity: SeverityWarning}},
		{},
		{Message{Severity: SeverityError}, Message{Severity: SeverityError}, Message{Severity: SeverityError}},
		{Message{Severity: SeverityError}},
		{},
	}
	for i := range messages {
		m := messages[i].Filter(filters[i])
		if !reflect.DeepEqual(m, expected[i]) {
			t.Errorf("expected m to be equal %v, was %v instead (i=%d)",
				expected[i], m, i)
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
