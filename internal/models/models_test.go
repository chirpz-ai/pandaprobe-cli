package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginatedDecode(t *testing.T) {
	raw := `{"items":[{"trace_id":"t1","name":"n","status":"COMPLETED","started_at":"2025-01-15T10:30:00Z","ended_at":null,"session_id":null,"user_id":null,"tags":["a"],"environment":"prod","release":null,"latency_ms":12.5,"span_count":3,"total_tokens":100,"total_cost":0.01}],"total":42,"limit":50,"offset":0}`
	var p Paginated[TraceListItem]
	require.NoError(t, json.Unmarshal([]byte(raw), &p))
	assert.Equal(t, 42, p.Total)
	require.Len(t, p.Items, 1)
	assert.Equal(t, TraceStatusCompleted, p.Items[0].Status)
	assert.Nil(t, p.Items[0].EndedAt)
	require.NotNil(t, p.Items[0].Environment)
	assert.Equal(t, "prod", *p.Items[0].Environment)
}

func TestAsListNilItems(t *testing.T) {
	p := &Paginated[SessionSummary]{Items: nil, Total: 0, Limit: 50, Offset: 0}
	out := AsList(p)
	assert.NotNil(t, out.Items)
	assert.Len(t, out.Items, 0)

	b, err := json.Marshal(out)
	require.NoError(t, err)
	// Empty items must serialize as [] not null for stable agent parsing.
	assert.Contains(t, string(b), `"items":[]`)
	assert.Contains(t, string(b), `"pagination":{"total":0,"limit":50,"offset":0}`)
}

func TestEnumValidity(t *testing.T) {
	assert.True(t, TraceStatusCompleted.Valid())
	assert.False(t, TraceStatus("BOGUS").Valid())
	assert.True(t, SpanKindLLM.Valid())
	assert.True(t, ScoreSourceProgrammatic.Valid())
	assert.False(t, ScoreDataType("x").Valid())
}

func TestCreateTraceScoreRequestOmitsEmpty(t *testing.T) {
	req := CreateTraceScoreRequest{TraceID: "t1", Name: "acc", Value: "0.9"}
	b, err := json.Marshal(req)
	require.NoError(t, err)
	s := string(b)
	assert.NotContains(t, s, "reason")
	assert.NotContains(t, s, "metadata")
	assert.NotContains(t, s, "data_type")
}
