package pulse

import "time"

// AgentStatus TODO(rjeczalik): document
type AgentStatus string

const (
	AgentOffline  AgentStatus = "offline"
	AgentSync     AgentStatus = "Synchronizing"
	AgentIdle     AgentStatus = "idle"
	AgentBuilding AgentStatus = "building"
	AgentDisabled AgentStatus = "disabled"
)

// Agent TODO(rjeczalik): document
type Agent struct {
	Name   string
	Status AgentStatus `xmlrpc:"status"`
	Host   string      `xmlrpc:"location"`
}

// String TODO(rjeczalik): document
func (a Agent) String() string { return a.Name + "@" + a.Host }

// TriggerOptions TODO(rjeczalik): document
// TODO(rjeczalik): TriggerOptions should be a map, because of the way Pulse
//                  handles optional fields - if a option is intended to not
//                  be overwritten it must not be sent in the request - Pulse
//                  does not treat empty values as a default ones.
//                  A map[string]interface{} is ideal to model this, but
//                  kolo/xmlrpc does not support maps.
type TriggerOptions struct {
	Force bool `xmlrpc:"force"`
	// TODO(rjeczalik): github.com/kolo/xmlrpc/issues/17
	Properties interface{} `xmlrpc:"properties"`
	Rebuild    bool        `xmlrpc:"rebuild"`
	Replace    bool        `xmlrpc:"replaceable"`
	Revision   string      `xmlrpc:"revision"`
	Status     string      `xmlrpc:"status"`
}

// BuildStatus TODO(rjeczalik): document
type BuildStatus string

const (
	BuildUnknown     BuildStatus = "UNKNOWN"
	BuildUnhandled   BuildStatus = "UNHANDLED"
	BuildRejected    BuildStatus = "REJECTED"
	BuildAssimilated BuildStatus = "ASSIMILATED"
	BuildQueued      BuildStatus = "QUEUED"
	BuildCancelled   BuildStatus = "CANCELLED"
	BuildActivated   BuildStatus = "ACTIVATED"
)

// BuildRequestStatus TODO(rjeczalik): document
type BuildRequestStatus struct {
	Status BuildStatus `xmlrpc:"status"`
	// TODO(rjeczalik): According to the API documentation ID and AssimID must
	//                  both be int, but Pulse sends it as strings.
	ID           string `xmlrpc:"buildId"`
	AssimID      string `xmlrpc:"assimilatedId"`
	RejectReason string `xmlrpc:"rejectionReason"`
}

// CommandResult TODO(rjeczalik): document
type CommandResult struct {
	Complete   bool                    `xmlrpc:"completed"`
	End        time.Time               `xmlrpc:"endTime"`
	Errors     int                     `xmlrpc:"errorCount"`
	Name       string                  `xmlrpc:"name"`
	Progress   int                     `xmlrpc:"progress"`
	Start      time.Time               `xmlrpc:"startTime"`
	Status     BuildStatus             `xmlrpc:"status"`
	Success    bool                    `xmlrpc:"succeeded"`
	Properties CommandResultProperties `xmlrpc:"properties"`
	Warnings   int                     `xmlrpc:"warningCount"`
}

// CommandResultProperties TODO(rjeczalik): document
type CommandResultProperties struct {
	// TODO(rjeczalik): According to the API documentation Exit must be int,
	//                  but Pulse sends it as string.
	Exit    string `xmlrpc:"exit code"`
	CmdLine string `xmlrpc:"command line"`
	WorkDir string `xmlrpc:"working directory"`
}

const (
	// AgentPending TODO(rjeczalik): document
	AgentPending = "[pending]"
)

// BuildState TODO(rjeczalik): document
type BuildState string

const (
	BuildCancelling  BuildState = "cancelling"
	BuildError       BuildState = "error"
	BuildFailure     BuildState = "failure"
	BuildInProgress  BuildState = "in progress"
	BuildPending     BuildState = "pending"
	BuildSkipped     BuildState = "skipped"
	BuildSuccess     BuildState = "success"
	BuildTerminating BuildState = "terminating"
	BuildTerminated  BuildState = "terminated"
	BuildWarnings    BuildState = "warnings"
)

// StageResult TODO(rjeczalik): document
type StageResult struct {
	Agent    string          `xmlrpc:"agent"`
	Complete bool            `xmlrpc:"completed"`
	End      time.Time       `xmlrpc:"endTime"`
	Errors   int             `xmlrpc:"errorCount"`
	Name     string          `xmlrpc:"name"`
	Progress int             `xmlrpc:"progress"`
	Start    time.Time       `xmlrpc:"startTime"`
	State    BuildState      `xmlrpc:"status"`
	Success  bool            `xmlrpc:"succeeded"`
	Test     TestSummary     `xmlrpc:"tests"`
	Command  []CommandResult `xmlrpc:"commands"`
	Warnings int             `xmlrpc:"warningCount"`
}

// TestSummary TODO(rjeczalik): document
type TestSummary struct {
	Total            int `xmlrpc:"total"`
	Errors           int `xmlrpc:"errors"`
	ExpectedFailures int `xmlrpc:"expectedFailures"`
	Failures         int `xmlrpc:"failures"`
	Passed           int `xmlrpc:"passed"`
	Skipped          int `xmlrpc:"skipped"`
}

// BuildResult TODO(rjeczalik): document
type BuildResult struct {
	ID        int64         `xmlrpc:"id"`
	Complete  bool          `xmlrpc:"completed"`
	End       time.Time     `xmlrpc:"endTime"`
	EndUnix   string        `xmlrpc:"endTimeMillis"`
	Errors    int           `xmlrpc:"errorCount"`
	Maturity  string        `xmlrpc:"maturity"`
	Owner     string        `xmlrpc:"owner"`
	Personal  bool          `xmlrpc:"personal"`
	Pinned    bool          `xmlrpc:"pinned"`
	Progress  int           `xmlrpc:"progress"`
	Project   string        `xmlrpc:"project"`
	Revision  string        `xmlrpc:"revision"`
	Reason    string        `xmlrpc:"reason"`
	Stages    []StageResult `xmlrpc:"stages"`
	Start     time.Time     `xmlrpc:"startTime"`
	StartUnix string        `xmlrpc:"startTimeMillis"`
	State     BuildState    `xmlrpc:"status"`
	Test      TestSummary   `xmlrpc:"tests"`
	Success   bool          `xmlrpc:"succeeded"`
	Version   string        `xmlrpc:"version"`
	Warnings  int           `xmlrpc:"warningCount"`
}

// Severity TODO(rjeczalik): document
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Message TODO(rjeczalik): document
type Message struct {
	Severity     Severity `xmlrpc:"level"`
	Message      string   `xmlrpc:"message"`
	StageName    string   `xmlrpc:"stage"`
	CommandName  string   `xmlrpc:"command"`
	ArtifactName string   `xmlrpc:"artifact"`
	Path         string   `xmlrpc:"path"`
}

// ProjectStage TODO(rjeczalik): document
// 'projects/$PROJECT/stages'
type ProjectStage struct {
	Meta      string `xmlrpc:"meta.symbolicName"`
	Name      string `xmlrpc:"name"`
	Recipe    string `xmlrpc:"recipe"`
	Agent     string `xmlrpc:"agent"`
	Enabled   bool   `xmlrpc:"enabled"`
	Terminate bool   `xmlrpc:"terminateBuildOnFailure"`
}
