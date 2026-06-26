package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

const monitorResp = `{"id":"mon1","project_id":"p","name":"m","target_type":"TRACE","metric_names":["accuracy"],"filters":{},"sampling_rate":1,"model":null,"cadence":"daily","only_if_changed":true,"status":"ACTIVE","last_run_at":null,"last_run_id":null,"next_run_at":"2026-01-01T00:00:00Z","created_at":"","updated_at":""}`

func TestMonitorsCreateTraceUppercasesTarget(t *testing.T) {
	var body []byte
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		body = buf.Bytes()
		w.WriteHeader(201)
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "create", "--name", "m", "--metrics", "accuracy", "--cadence", "daily")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Equal(t, "POST", method)
	assert.Equal(t, "/evaluations/monitors", path)
	assert.Contains(t, string(body), `"target_type":"TRACE"`)
	assert.Contains(t, string(body), `"cadence":"daily"`)
	assert.Contains(t, out, `"id": "mon1"`)
}

func TestMonitorsCreateRequiresName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected")
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "create", "--metrics", "accuracy", "--cadence", "daily")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "name")
}

func TestMonitorsCreateRequiresMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected")
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "create", "--name", "m", "--cadence", "daily")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "metrics")
}

func TestMonitorsCreateRejectsBadCadence(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected")
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "create", "--name", "m", "--metrics", "accuracy", "--cadence", "hourly")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "cadence")
}

func TestMonitorsCreateCronCadence(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		body = buf.Bytes()
		w.WriteHeader(201)
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "create", "--name", "m", "--metrics", "accuracy", "--cadence", "cron:0 */6 * * *")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, string(body), `"cadence":"cron:0 */6 * * *"`)
}

func TestMonitorsCreateSessionSignalWeights(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		body = buf.Bytes()
		w.WriteHeader(201)
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "create", "--target", "session",
		"--name", "m", "--metrics", "coherence", "--cadence", "weekly",
		"--signal-weights", `{"coherence":0.5}`)
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, string(body), `"target_type":"SESSION"`)
	assert.Contains(t, string(body), `"signal_weights":{"coherence":0.5}`)
	assert.Contains(t, out, `"id": "mon1"`)
}

func TestMonitorsListSendsStatusAndPagination(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"items":[` + monitorResp + `],"total":1,"limit":50,"offset":0}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "list", "--status", "ACTIVE", "--limit", "50", "--offset", "10")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, gotQuery, "status=ACTIVE")
	assert.Contains(t, gotQuery, "limit=50")
	assert.Contains(t, gotQuery, "offset=10")
	assert.Contains(t, out, `"pagination"`)
}

func TestMonitorsListRejectsBadStatus(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "list", "--status", "BOGUS")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "validation_error")
	assert.False(t, hit, "no request should be made when validation fails")
}

func TestMonitorsListTable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[` + monitorResp + `],"total":1,"limit":50,"offset":0}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, _, code := runCLI(t, "evals", "monitors", "list", "--format", "table", "--no-color")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, "CADENCE")
	assert.Contains(t, out, "mon1")
	assert.NotContains(t, out, "\033[")
}

func TestMonitorsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/evaluations/monitors/mon1", r.URL.Path)
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "get", "mon1")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, out, `"id": "mon1"`)
}

func TestMonitorsUpdateSendsOnlyChangedFields(t *testing.T) {
	var body []byte
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		body = buf.Bytes()
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "update", "mon1", "--name", "renamed")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Equal(t, "PATCH", method)
	assert.Contains(t, string(body), `"name":"renamed"`)
	assert.NotContains(t, string(body), `"cadence"`)
	assert.NotContains(t, string(body), `"sampling_rate"`)
	assert.Contains(t, out, `"id": "mon1"`)
}

func TestMonitorsUpdateRejectsBadCadence(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected")
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "update", "mon1", "--cadence", "nonsense")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "cadence")
}

func TestMonitorsUpdateSendsValidFiltersObject(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		body = buf.Bytes()
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "update", "mon1", "--filters", `{"status":"COMPLETED"}`)
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, string(body), `"filters":{"status":"COMPLETED"}`)
	assert.Contains(t, out, `"id": "mon1"`)
}

func TestMonitorsUpdateRejectsNonObjectFilters(t *testing.T) {
	// Valid JSON that is not an object (array, string, number, bool) must be
	// rejected client-side and never reach the API.
	for _, bad := range []string{`[1,2,3]`, `"oops"`, `42`, `true`} {
		t.Run(bad, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("no request expected")
			}))
			defer srv.Close()
			withAuthEnv(t, srv)

			_, errOut, code := runCLI(t, "evals", "monitors", "update", "mon1", "--filters", bad)
			assert.Equal(t, exitcode.Validation, code)
			assert.Contains(t, errOut, "filters")
		})
	}
}

func TestMonitorsPauseResumePaths(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte(monitorResp))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "monitors", "pause", "mon1")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Equal(t, "/evaluations/monitors/mon1/pause", path)

	_, errOut, code = runCLI(t, "evals", "monitors", "resume", "mon1")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Equal(t, "/evaluations/monitors/mon1/resume", path)
}

func TestMonitorsDeleteRendersConfirmation(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.WriteHeader(204)
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "delete", "mon1")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Equal(t, "DELETE", method)
	assert.Equal(t, "/evaluations/monitors/mon1", path)
	assert.Contains(t, out, `"status": "deleted"`)
	assert.Contains(t, out, `"id": "mon1"`)
}

func TestMonitorsRunsPaginated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/evaluations/monitors/mon1/runs", r.URL.Path)
		_, _ = w.Write([]byte(`{"items":[{"id":"run1","status":"COMPLETED","metric_names":[],"total_targets":1,"evaluated_count":1,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}],"total":1,"limit":50,"offset":0}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "runs", "mon1")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, out, `"id": "run1"`)
	assert.Contains(t, out, `"pagination"`)
}

func TestMonitorsTriggerReturnsRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/evaluations/monitors/mon1/trigger", r.URL.Path)
		w.WriteHeader(202)
		_, _ = w.Write([]byte(`{"id":"run2","status":"PENDING","metric_names":["accuracy"],"total_targets":0,"evaluated_count":0,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "monitors", "trigger", "mon1")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, out, `"id": "run2"`)
	assert.Contains(t, out, `"status": "PENDING"`)
}
