// TODO(rjeczalik): Agents.Filter{,Out} and Messages.Filter{,Out} share the same
//                  implementation, but because of lack of generics it is
//                  duplicated right now. If it turns out it must be duplicated
//                  for even more types it would be nice to find out whether it's
//                  possible to create cheap implementation using reflect package.
package pulse

// Offline predicate returns true when the agent has an offline state.
var Offline = func(agent *Agent) bool {
	return agent.Status == AgentOffline
}

// Sync predicate returns true when the agent has a synchronizing state.
var Sync = func(agent *Agent) bool {
	return agent.Status == AgentSync
}

// Agents is an utility wrapper for a slice of agents, which extends it with
// a filtering functionality.
type Agents []Agent

// Filter returns a slice which is a subset of Agents. Every agent
// in the subset fulfills every predicate. A predicate must not modify
// the Agent struct. The method returns nil as soon as resulting set becomes
// empty, which may cause that not all the predicates might get called.
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
	if len(b) == 0 {
		return nil
	}
	if len(pred) == 1 {
		return b
	}
	return b.Filter(pred[1:]...)
}

// FilterOut behaves exacly like Filter with the only exception, that resulting
// subset contains only elements, that do not fulfill all the predicates.
func (a Agents) FilterOut(pred ...func(*Agent) bool) Agents {
	if len(pred) == 0 {
		panic("pulse: missing predicate")
	}
	notpred := make([]func(*Agent) bool, 0, len(pred))
	for _, pred := range pred {
		pred := pred
		notpred = append(notpred, func(a *Agent) bool { return !pred(a) })
	}
	return a.Filter(notpred...)
}

// Info predicate returns true when the message is of an information kind.
var Info = func(m *Message) bool {
	return m.Severity == SeverityInfo
}

// Warning predicate returns true when the message is of a warning kind.
var Warning = func(m *Message) bool {
	return m.Severity == SeverityWarning
}

// Error predicate returns true when the message is of an error kind.
var Error = func(m *Message) bool {
	return m.Severity == SeverityError
}

// Messages is an utility wrapper for a slice of messages, which extends it with
// a filtering functionality.
type Messages []Message

// Filter returns a slice which is a subset of Messages. Every message
// in the subset fulfills every predicate. A predicate must not modify
// the Message struct. The method returns nil as soon as resulting set becomes
// empty, which may cause that not all the predicates might get called.
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
	if len(n) == 0 {
		return nil
	}
	if len(pred) == 1 {
		return n
	}
	return n.Filter(pred[1:]...)
}

// FilterOut behaves exacly like Filter with the only exception, that resulting
// subset contains only elements, that do not fulfill all the predicates.
func (m Messages) FilterOut(pred ...func(*Message) bool) Messages {
	if len(pred) == 0 {
		panic("pulse: missing predicate")
	}
	notpred := make([]func(*Message) bool, 0, len(pred))
	for _, pred := range pred {
		pred := pred
		notpred = append(notpred, func(m *Message) bool { return !pred(m) })
	}
	return m.Filter(notpred...)
}
