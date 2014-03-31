package pulse

import (
	"reflect"
	"strings"
	"testing"
)

func TestFilter(t *testing.T) {
	agents := [][]Agent{
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
		IsOffline,
		IsOffline,
		IsOffline,
		IsSync,
		IsSync,
		IsSync,
		func(a *Agent) bool { return strings.HasPrefix(a.Host, "win-") },
		func(a *Agent) bool { return strings.HasPrefix(a.Host, "dev-") },
		func(a *Agent) bool { return strings.HasPrefix(a.Host, "pulse-") },
	}
	expected := [][]Agent{
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
		filtered := Filter(agents[i], filters[i])
		if !reflect.DeepEqual(filtered, expected[i]) {
			t.Errorf("expected filtered to be equal %v, was %v instead (i=%d)",
				expected[i], filtered, i)
		}
	}
}
