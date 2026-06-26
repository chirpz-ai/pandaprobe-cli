package models

// Enum is a string-backed enum that can validate itself and report its members.
type Enum interface {
	~string
	Valid() bool
}

// --- Trace ---

// TraceStatus is the lifecycle status of a trace.
type TraceStatus string

const (
	TraceStatusPending   TraceStatus = "PENDING"
	TraceStatusRunning   TraceStatus = "RUNNING"
	TraceStatusCompleted TraceStatus = "COMPLETED"
	TraceStatusError     TraceStatus = "ERROR"
)

// TraceStatusValues lists the valid trace statuses.
func TraceStatusValues() []string {
	return []string{"PENDING", "RUNNING", "COMPLETED", "ERROR"}
}

// Valid reports whether s is a known trace status.
func (s TraceStatus) Valid() bool { return contains(TraceStatusValues(), string(s)) }

// TraceSortBy enumerates the fields traces can be sorted by.
func TraceSortByValues() []string {
	return []string{"started_at", "ended_at", "name", "latency", "status"}
}

// --- Span ---

// SpanKind classifies a span's operation type.
type SpanKind string

const (
	SpanKindAgent     SpanKind = "AGENT"
	SpanKindTool      SpanKind = "TOOL"
	SpanKindLLM       SpanKind = "LLM"
	SpanKindRetriever SpanKind = "RETRIEVER"
	SpanKindChain     SpanKind = "CHAIN"
	SpanKindEmbedding SpanKind = "EMBEDDING"
	SpanKindOther     SpanKind = "OTHER"
)

// SpanKindValues lists the valid span kinds.
func SpanKindValues() []string {
	return []string{"AGENT", "TOOL", "LLM", "RETRIEVER", "CHAIN", "EMBEDDING", "OTHER"}
}

// Valid reports whether k is a known span kind.
func (k SpanKind) Valid() bool { return contains(SpanKindValues(), string(k)) }

// SpanStatusCode is a span's outcome.
type SpanStatusCode string

const (
	SpanStatusUnset SpanStatusCode = "UNSET"
	SpanStatusOK    SpanStatusCode = "OK"
	SpanStatusError SpanStatusCode = "ERROR"
)

// SpanStatusValues lists the valid span statuses.
func SpanStatusValues() []string { return []string{"UNSET", "OK", "ERROR"} }

// Valid reports whether s is a known span status.
func (s SpanStatusCode) Valid() bool { return contains(SpanStatusValues(), string(s)) }

// --- Sort order (shared) ---

// SortOrderValues lists the valid sort directions.
func SortOrderValues() []string { return []string{"asc", "desc"} }

// SessionSortByValues lists the valid session sort fields.
func SessionSortByValues() []string {
	return []string{"recent", "trace_count", "latency", "cost"}
}

// --- Evaluation ---

// EvaluationStatus is the status of an evaluation run.
type EvaluationStatus string

const (
	EvalStatusPending   EvaluationStatus = "PENDING"
	EvalStatusRunning   EvaluationStatus = "RUNNING"
	EvalStatusCompleted EvaluationStatus = "COMPLETED"
	EvalStatusFailed    EvaluationStatus = "FAILED"
)

// EvaluationStatusValues lists the valid evaluation-run statuses.
func EvaluationStatusValues() []string {
	return []string{"PENDING", "RUNNING", "COMPLETED", "FAILED"}
}

// Valid reports whether s is a known evaluation status.
func (s EvaluationStatus) Valid() bool { return contains(EvaluationStatusValues(), string(s)) }

// MonitorStatus is the lifecycle state of an evaluation monitor.
type MonitorStatus string

const (
	MonitorStatusActive MonitorStatus = "ACTIVE"
	MonitorStatusPaused MonitorStatus = "PAUSED"
)

// MonitorStatusValues lists the valid monitor statuses.
func MonitorStatusValues() []string {
	return []string{"ACTIVE", "PAUSED"}
}

// Valid reports whether s is a known monitor status.
func (s MonitorStatus) Valid() bool { return contains(MonitorStatusValues(), string(s)) }

// ScoreStatus is the status of an individual score.
type ScoreStatus string

// ScoreStatusValues lists the valid score statuses.
func ScoreStatusValues() []string { return []string{"SUCCESS", "FAILED", "PENDING"} }

// Valid reports whether s is a known score status.
func (s ScoreStatus) Valid() bool { return contains(ScoreStatusValues(), string(s)) }

// ScoreSource identifies who/what produced a score.
type ScoreSource string

const (
	ScoreSourceAutomated    ScoreSource = "AUTOMATED"
	ScoreSourceAnnotation   ScoreSource = "ANNOTATION"
	ScoreSourceProgrammatic ScoreSource = "PROGRAMMATIC"
)

// ScoreSourceValues lists the valid score sources.
func ScoreSourceValues() []string { return []string{"AUTOMATED", "ANNOTATION", "PROGRAMMATIC"} }

// Valid reports whether s is a known score source.
func (s ScoreSource) Valid() bool { return contains(ScoreSourceValues(), string(s)) }

// ScoreDataType is the value domain of a score.
type ScoreDataType string

const (
	ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
	ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"
)

// ScoreDataTypeValues lists the valid score data types.
func ScoreDataTypeValues() []string { return []string{"NUMERIC", "BOOLEAN", "CATEGORICAL"} }

// Valid reports whether d is a known score data type.
func (d ScoreDataType) Valid() bool { return contains(ScoreDataTypeValues(), string(d)) }

// EvalTargetValues lists the valid evaluation targets (CLI-only concept).
func EvalTargetValues() []string { return []string{"trace", "session"} }

func contains(set []string, s string) bool {
	for _, v := range set {
		if v == s {
			return true
		}
	}
	return false
}
