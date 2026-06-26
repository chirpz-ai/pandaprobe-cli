package models

import "encoding/json"

// MonitorResponse is the result of creating, fetching, or mutating an
// evaluation monitor. The same shape covers both trace and session monitors
// (distinguished by TargetType).
type MonitorResponse struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	Name          string        `json:"name"`
	TargetType    string        `json:"target_type"`
	MetricNames   []string      `json:"metric_names"`
	Filters       JSON          `json:"filters"`
	SamplingRate  float64       `json:"sampling_rate"`
	Model         *string       `json:"model"`
	Cadence       string        `json:"cadence"`
	OnlyIfChanged bool          `json:"only_if_changed"`
	Status        MonitorStatus `json:"status"`
	LastRunAt     *string       `json:"last_run_at"`
	LastRunID     *string       `json:"last_run_id"`
	NextRunAt     *string       `json:"next_run_at"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

// CreateMonitorRequest is the body for POST /evaluations/monitors. Filters
// holds the typed per-target filter struct (*EvalRunFilters for TRACE,
// *SessionEvalRunFilters for SESSION).
type CreateMonitorRequest struct {
	Name          string             `json:"name"`
	TargetType    string             `json:"target_type"`
	Metrics       []string           `json:"metrics"`
	Cadence       string             `json:"cadence"`
	Filters       any                `json:"filters,omitempty"`
	SamplingRate  *float64           `json:"sampling_rate,omitempty"`
	Model         *string            `json:"model,omitempty"`
	OnlyIfChanged *bool              `json:"only_if_changed,omitempty"`
	SignalWeights map[string]float64 `json:"signal_weights,omitempty"`
}

// UpdateMonitorRequest is the body for PATCH /evaluations/monitors/{id}. Every
// field is optional; only the fields the user changed are sent. Filters and
// SignalWeights are raw JSON so an explicit null is forwarded verbatim (the API
// treats null as "reset"), which a nil map/struct could not express.
type UpdateMonitorRequest struct {
	Name          *string         `json:"name,omitempty"`
	Metrics       []string        `json:"metrics,omitempty"`
	Filters       json.RawMessage `json:"filters,omitempty"`
	SamplingRate  *float64        `json:"sampling_rate,omitempty"`
	Model         *string         `json:"model,omitempty"`
	Cadence       *string         `json:"cadence,omitempty"`
	OnlyIfChanged *bool           `json:"only_if_changed,omitempty"`
	SignalWeights json.RawMessage `json:"signal_weights,omitempty"`
}
