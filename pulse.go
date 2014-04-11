// TODO(rjeczalik): Agents.Filter and Messages.Filter share basically the same
//                  implementation, but because of lack of generics it is
//                  duplicated right now. If it turns out it must be duplicated
//                  for even more types it would be nice to find out whether it's
//                  possible to create cheap implementation using reflect package.
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

// Info TODO(rjeczalik): document
var Info = func(m *Message) bool {
	return m.Severity == SeverityInfo
}

// Warning TODO(rjeczalik): document
var Warning = func(m *Message) bool {
	return m.Severity == SeverityWarning
}

// Error TODO(rjeczalik): document
var Error = func(m *Message) bool {
	return m.Severity == SeverityError
}

// Messages TODO(rjeczalik): document
type Messages []Message

// Filter TODO(rjeczalik): document
func (m Messages) Filter(pred ...func(*Message) bool) Messages {
	if len(pred) == 0 {
		panic("pulse: missing predicate")
	}
	n := make(Messages, 0)
	for i := range m {
		if pred[0](&m[i]) {
			n = append(n, m[i])
		}
	}
	if len(pred) == 1 {
		return n
	}
	return n.Filter(pred[1:]...)
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
