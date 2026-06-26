package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

const monitorBody = `{"id":"mon1","project_id":"p","name":"m","target_type":"TRACE","metric_names":["accuracy"],"filters":{},"sampling_rate":1,"model":null,"cadence":"daily","only_if_changed":true,"status":"ACTIVE","last_run_at":null,"last_run_id":null,"next_run_at":"2026-01-01T00:00:00Z","created_at":"","updated_at":""}`

func TestMonitorsLifecycle(t *testing.T) {
	var lastPath, lastMethod string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		lastPath, lastMethod = r.URL.Path, r.Method
		switch {
		case r.URL.Path == "/evaluations/monitors" && r.Method == "POST":
			w.WriteHeader(201)
			_, _ = w.Write([]byte(monitorBody))
		case r.URL.Path == "/evaluations/monitors" && r.Method == "GET":
			_, _ = w.Write([]byte(`{"items":[` + monitorBody + `],"total":1,"limit":50,"offset":0}`))
		case r.URL.Path == "/evaluations/monitors/mon1" && r.Method == "GET":
			_, _ = w.Write([]byte(monitorBody))
		case r.URL.Path == "/evaluations/monitors/mon1" && r.Method == "PATCH":
			_, _ = w.Write([]byte(monitorBody))
		case r.URL.Path == "/evaluations/monitors/mon1" && r.Method == "DELETE":
			w.WriteHeader(204)
		case r.URL.Path == "/evaluations/monitors/mon1/pause":
			_, _ = w.Write([]byte(monitorBody))
		case r.URL.Path == "/evaluations/monitors/mon1/resume":
			_, _ = w.Write([]byte(monitorBody))
		case r.URL.Path == "/evaluations/monitors/mon1/runs":
			_, _ = w.Write([]byte(`{"items":[{"id":"run1","status":"COMPLETED","metric_names":[],"total_targets":1,"evaluated_count":1,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}],"total":1,"limit":50,"offset":0}`))
		case r.URL.Path == "/evaluations/monitors/mon1/trigger":
			w.WriteHeader(202)
			_, _ = w.Write([]byte(`{"id":"run2","status":"PENDING","metric_names":["accuracy"],"total_targets":0,"evaluated_count":0,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	ctx := context.Background()

	created, err := c.CreateMonitor(ctx, models.CreateMonitorRequest{Name: "m", TargetType: "TRACE", Metrics: []string{"accuracy"}, Cadence: "daily"})
	require.NoError(t, err)
	assert.Equal(t, "mon1", created.ID)
	assert.Equal(t, models.MonitorStatusActive, created.Status)

	list, err := c.ListMonitors(ctx, nil)
	require.NoError(t, err)
	require.Len(t, list.Items, 1)

	got, err := c.GetMonitor(ctx, "mon1")
	require.NoError(t, err)
	assert.Equal(t, "mon1", got.ID)

	name := "renamed"
	updated, err := c.UpdateMonitor(ctx, "mon1", models.UpdateMonitorRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "mon1", updated.ID)
	assert.Equal(t, "PATCH", lastMethod)

	paused, err := c.PauseMonitor(ctx, "mon1")
	require.NoError(t, err)
	assert.Equal(t, "mon1", paused.ID)
	assert.Equal(t, "/evaluations/monitors/mon1/pause", lastPath)

	resumed, err := c.ResumeMonitor(ctx, "mon1")
	require.NoError(t, err)
	assert.Equal(t, "mon1", resumed.ID)
	assert.Equal(t, "/evaluations/monitors/mon1/resume", lastPath)

	runs, err := c.MonitorRuns(ctx, "mon1", nil)
	require.NoError(t, err)
	require.Len(t, runs.Items, 1)
	assert.Equal(t, "run1", runs.Items[0].ID)

	triggered, err := c.TriggerMonitor(ctx, "mon1")
	require.NoError(t, err)
	assert.Equal(t, "run2", triggered.ID)
	assert.Equal(t, models.EvalStatusPending, triggered.Status)

	err = c.DeleteMonitor(ctx, "mon1")
	require.NoError(t, err)
	assert.Equal(t, "/evaluations/monitors/mon1", lastPath)
	assert.Equal(t, "DELETE", lastMethod)
}
