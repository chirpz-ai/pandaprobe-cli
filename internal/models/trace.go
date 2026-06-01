package models

// TraceListItem is a row in the GET /traces list response.
type TraceListItem struct {
	TraceID     string      `json:"trace_id"`
	Name        string      `json:"name"`
	Status      TraceStatus `json:"status"`
	StartedAt   string      `json:"started_at"`
	EndedAt     *string     `json:"ended_at"`
	SessionID   *string     `json:"session_id"`
	UserID      *string     `json:"user_id"`
	Tags        []string    `json:"tags"`
	Environment *string     `json:"environment"`
	Release     *string     `json:"release"`
	LatencyMs   *float64    `json:"latency_ms"`
	SpanCount   int         `json:"span_count"`
	TotalTokens int         `json:"total_tokens"`
	TotalCost   float64     `json:"total_cost"`
}

// TraceResponse is the full GET /traces/{trace_id} response, with spans inlined.
type TraceResponse struct {
	TraceID     string         `json:"trace_id"`
	ProjectID   string         `json:"project_id"`
	Name        string         `json:"name"`
	Status      TraceStatus    `json:"status"`
	Input       JSON           `json:"input"`
	Output      JSON           `json:"output"`
	Metadata    JSON           `json:"metadata"`
	StartedAt   string         `json:"started_at"`
	EndedAt     *string        `json:"ended_at"`
	SessionID   *string        `json:"session_id"`
	UserID      *string        `json:"user_id"`
	Tags        []string       `json:"tags"`
	Environment *string        `json:"environment"`
	Release     *string        `json:"release"`
	Spans       []SpanResponse `json:"spans"`
	TotalTokens int            `json:"total_tokens"`
	TotalCost   float64        `json:"total_cost"`
}

// SpanResponse is a single span within a trace.
type SpanResponse struct {
	SpanID              string         `json:"span_id"`
	TraceID             string         `json:"trace_id"`
	ParentSpanID        *string        `json:"parent_span_id"`
	Name                string         `json:"name"`
	Kind                SpanKind       `json:"kind"`
	Status              SpanStatusCode `json:"status"`
	Input               JSON           `json:"input"`
	Output              JSON           `json:"output"`
	Model               *string        `json:"model"`
	TokenUsage          JSON           `json:"token_usage"`
	Metadata            JSON           `json:"metadata"`
	StartedAt           string         `json:"started_at"`
	EndedAt             *string        `json:"ended_at"`
	Error               *string        `json:"error"`
	CompletionStartTime *string        `json:"completion_start_time"`
	ModelParameters     JSON           `json:"model_parameters"`
	Cost                JSON           `json:"cost"`
	LatencyMs           *float64       `json:"latency_ms"`
	TimeToFirstTokenMs  *float64       `json:"time_to_first_token_ms"`
}
