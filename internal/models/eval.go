package models

// MetricSummary describes an available evaluation metric.
type MetricSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// EvalRunResponse is the result of creating or fetching an evaluation run. The
// same shape is used for both trace and session runs (distinguished by
// TargetType).
type EvalRunResponse struct {
	ID             string           `json:"id"`
	Name           *string          `json:"name"`
	Status         EvaluationStatus `json:"status"`
	MetricNames    []string         `json:"metric_names"`
	TotalTargets   int              `json:"total_targets"`
	EvaluatedCount int              `json:"evaluated_count"`
	FailedCount    int              `json:"failed_count"`
	CreatedAt      string           `json:"created_at"`
	CompletedAt    *string          `json:"completed_at"`
	ProjectID      string           `json:"project_id"`
	TargetType     string           `json:"target_type"`
	Filters        JSON             `json:"filters"`
	SamplingRate   float64          `json:"sampling_rate"`
	Model          *string          `json:"model"`
	MonitorID      *string          `json:"monitor_id"`
	ErrorMessage   *string          `json:"error_message"`
}

// TraceScoreResponse is a single trace-level score.
type TraceScoreResponse struct {
	ID           string        `json:"id"`
	TraceID      string        `json:"trace_id"`
	Name         string        `json:"name"`
	Value        *string       `json:"value"`
	Status       ScoreStatus   `json:"status"`
	Source       ScoreSource   `json:"source"`
	CreatedAt    string        `json:"created_at"`
	ProjectID    string        `json:"project_id"`
	DataType     ScoreDataType `json:"data_type"`
	EvalRunID    *string       `json:"eval_run_id"`
	AuthorUserID *string       `json:"author_user_id"`
	Reason       *string       `json:"reason"`
	Environment  *string       `json:"environment"`
	ConfigID     *string       `json:"config_id"`
	Metadata     JSON          `json:"metadata"`
	UpdatedAt    string        `json:"updated_at"`
}

// SessionScoreResponse is a single session-level score.
type SessionScoreResponse struct {
	ID           string  `json:"id"`
	SessionID    string  `json:"session_id"`
	ProjectID    string  `json:"project_id"`
	Name         string  `json:"name"`
	DataType     string  `json:"data_type"`
	Value        *string `json:"value"`
	Source       string  `json:"source"`
	Status       string  `json:"status"`
	EvalRunID    *string `json:"eval_run_id"`
	AuthorUserID *string `json:"author_user_id"`
	Reason       *string `json:"reason"`
	Metadata     JSON    `json:"metadata"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// --- Request bodies ---

// EvalRunFilters selects traces for a filter-based trace evaluation run.
type EvalRunFilters struct {
	DateFrom  *string  `json:"date_from,omitempty"`
	DateTo    *string  `json:"date_to,omitempty"`
	Status    *string  `json:"status,omitempty"`
	SessionID *string  `json:"session_id,omitempty"`
	UserID    *string  `json:"user_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Name      *string  `json:"name,omitempty"`
}

// CreateEvalRunRequest is the body for POST /evaluations/trace-runs.
type CreateEvalRunRequest struct {
	Name         *string         `json:"name,omitempty"`
	Metrics      []string        `json:"metrics"`
	Filters      *EvalRunFilters `json:"filters,omitempty"`
	SamplingRate *float64        `json:"sampling_rate,omitempty"`
	Model        *string         `json:"model,omitempty"`
}

// CreateBatchEvalRunRequest is the body for POST /evaluations/trace-runs/batch.
type CreateBatchEvalRunRequest struct {
	TraceIDs []string `json:"trace_ids"`
	Metrics  []string `json:"metrics"`
	Name     *string  `json:"name,omitempty"`
	Model    *string  `json:"model,omitempty"`
}

// SessionEvalRunFilters selects sessions for a filter-based session run.
type SessionEvalRunFilters struct {
	DateFrom      *string  `json:"date_from,omitempty"`
	DateTo        *string  `json:"date_to,omitempty"`
	UserID        *string  `json:"user_id,omitempty"`
	HasError      *bool    `json:"has_error,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	MinTraceCount *int     `json:"min_trace_count,omitempty"`
}

// CreateSessionEvalRunRequest is the body for POST /evaluations/session-runs.
type CreateSessionEvalRunRequest struct {
	Name          *string                `json:"name,omitempty"`
	Metrics       []string               `json:"metrics"`
	Filters       *SessionEvalRunFilters `json:"filters,omitempty"`
	SamplingRate  *float64               `json:"sampling_rate,omitempty"`
	Model         *string                `json:"model,omitempty"`
	SignalWeights map[string]float64     `json:"signal_weights,omitempty"`
}

// CreateBatchSessionEvalRunRequest is the body for POST /evaluations/session-runs/batch.
type CreateBatchSessionEvalRunRequest struct {
	SessionIDs    []string           `json:"session_ids"`
	Metrics       []string           `json:"metrics"`
	Name          *string            `json:"name,omitempty"`
	Model         *string            `json:"model,omitempty"`
	SignalWeights map[string]float64 `json:"signal_weights,omitempty"`
}

// CreateTraceScoreRequest is the body for POST /evaluations/trace-scores.
type CreateTraceScoreRequest struct {
	TraceID  string  `json:"trace_id"`
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	DataType string  `json:"data_type,omitempty"`
	Source   string  `json:"source,omitempty"`
	Reason   *string `json:"reason,omitempty"`
	Metadata JSON    `json:"metadata,omitempty"`
}
