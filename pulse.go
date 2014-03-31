package pulse

// IsOffline TODO(rjeczalik): document
var IsOffline = func(a *Agent) bool {
	return a.Status == AgentOffline
}

// IsSync TODO(rjeczalik): document
var IsSync = func(a *Agent) bool {
	return a.Status == AgentSync
}

// Filter TODO(rjeczalik): document
func Filter(a []Agent, pred func(*Agent) bool) []Agent {
	filt := make([]Agent, 0)
	for i := range a {
		if pred(&a[i]) {
			filt = append(filt, a[i])
		}
	}
	return filt
}
