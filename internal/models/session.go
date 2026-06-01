package models

// SessionSummary is a row in the GET /sessions list response: aggregates only,
// no traces.
type SessionSummary struct {
	SessionID      string   `json:"session_id"`
	TraceCount     int      `json:"trace_count"`
	FirstTraceAt   string   `json:"first_trace_at"`
	LastTraceAt    *string  `json:"last_trace_at"`
	TotalLatencyMs *float64 `json:"total_latency_ms"`
	HasError       bool     `json:"has_error"`
	UserID         *string  `json:"user_id"`
	Tags           []string `json:"tags"`
	TotalSpanCount int      `json:"total_span_count"`
	TotalTokens    int      `json:"total_tokens"`
	TotalCost      float64  `json:"total_cost"`
}

// SessionDetail is the GET /sessions/{session_id} response: the summary plus the
// session's traces (each with spans inlined).
type SessionDetail struct {
	SessionSummary
	Traces []TraceResponse `json:"traces"`
}
