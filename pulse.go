package pulse

// Offline TODO(rjeczalik): document
var Offline = func(a *Agent) bool {
	return a.Status == AgentOffline
}

// Sync TODO(rjeczalik): document
var Sync = func(a *Agent) bool {
	return a.Status == AgentSync
}

// Agents TODO(rjeczalik): document
type Agents []Agent

// Filter TODO(rjeczalik): document
func (a Agents) Filter(pred ...func(*Agent) bool) Agents {
	if len(pred) == 0 {
		panic("pulse: missing predicate")
	}
	b := make(Agents, 0)
	for i := range a {
		if pred[0](&a[i]) {
			b = append(b, a[i])
		}
	}
	if len(pred) == 1 {
		return b
	}
	return b.Filter(pred[1:]...)
}

// Pending TODO(rjeczalik): document
func Pending(v interface{}) bool {
	switch v := v.(type) {
	case *BuildResult:
		return Pending(&v.Stages)
	case *[]BuildResult:
		for i := range *v {
			if Pending(&(*v)[i].Stages) {
				return true
			}
		}
	case *StageResult:
		return v.Agent == AgentPending
	case *[]StageResult:
		for i := range *v {
			if Pending(&(*v)[i]) {
				return true
			}
		}
	}
	return false
}
