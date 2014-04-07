package pulse

import "testing"

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
