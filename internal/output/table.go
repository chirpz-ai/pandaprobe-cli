package output

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
	"github.com/chirpz-ai/pandaprobe-cli/internal/version"
)

// ANSI color codes (emitted only when color is enabled).
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorDim   = "\033[2m"
)

func (w *Writer) color(code string) string {
	if w.noColor {
		return ""
	}
	return code
}

// renderTable dispatches on concrete type. Unknown types fall back to pretty
// JSON so table mode never fails to produce output.
func (w *Writer) renderTable(v any) error {
	switch t := v.(type) {
	case models.ListResult[models.TraceListItem]:
		return w.traceListTable(t)
	case models.ListResult[models.SessionSummary]:
		return w.sessionListTable(t)
	case models.ListResult[models.EvalRunResponse]:
		return w.evalRunListTable(t.Items, &t.Pagination)
	case models.ListResult[models.TraceScoreResponse]:
		return w.traceScoreTable(t.Items, &t.Pagination)
	case models.ListResult[models.SessionScoreResponse]:
		return w.sessionScoreTable(t.Items, &t.Pagination)
	case []models.SpanResponse:
		return w.spanTable(t)
	case []models.MetricSummary:
		return w.metricTable(t)
	case []models.TraceScoreResponse:
		return w.traceScoreTable(t, nil)
	case []models.SessionScoreResponse:
		return w.sessionScoreTable(t, nil)
	case []models.EvalRunResponse:
		return w.evalRunListTable(t, nil)
	case *models.TraceResponse:
		return w.traceDetail(t)
	case *models.SessionDetail:
		return w.sessionDetail(t)
	case *models.EvalRunResponse:
		return w.evalRunDetail(t)
	case *models.TraceScoreResponse:
		return w.kvTable(traceScoreRows(*t))
	case version.Info:
		return w.kvTable([][2]string{
			{"version", t.Version}, {"commit", t.Commit}, {"build_date", t.BuildDate},
			{"go_version", t.GoVersion}, {"os", t.OS}, {"arch", t.Arch},
		})
	case map[string]string:
		rows := make([][2]string, 0, len(t))
		for _, k := range sortedKeys(t) {
			rows = append(rows, [2]string{k, t[k]})
		}
		return w.kvTable(rows)
	default:
		// Robust fallback: pretty JSON.
		return w.renderJSON(w.out, v)
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// simple insertion sort to avoid importing sort for tiny maps
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return keys
}

func (w *Writer) tw() *tabwriter.Writer {
	return tabwriter.NewWriter(w.out, 0, 4, 2, ' ', 0)
}

func (w *Writer) rowTable(headers []string, rows [][]string) error {
	tw := w.tw()
	if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(r, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (w *Writer) kvTable(rows [][2]string) error {
	tw := w.tw()
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", r[0], r[1]); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (w *Writer) footer(p *models.Pagination) error {
	if p == nil {
		return nil
	}
	_, err := fmt.Fprintf(w.out, "%s(total %d, limit %d, offset %d)%s\n",
		w.color(colorDim), p.Total, p.Limit, p.Offset, w.color(colorReset))
	return err
}

// statusCell colorizes well-known status values.
func (w *Writer) statusCell(s string) string {
	switch s {
	case "OK", "COMPLETED", "SUCCESS":
		return w.color(colorGreen) + s + w.color(colorReset)
	case "ERROR", "FAILED":
		return w.color(colorRed) + s + w.color(colorReset)
	default:
		return s
	}
}

func (w *Writer) traceListTable(t models.ListResult[models.TraceListItem]) error {
	rows := make([][]string, 0, len(t.Items))
	for _, it := range t.Items {
		rows = append(rows, []string{
			it.TraceID,
			truncate(it.Name, 40),
			w.statusCell(string(it.Status)),
			it.StartedAt,
			fmtFloatPtr(it.LatencyMs),
			strconv.Itoa(it.SpanCount),
			strconv.Itoa(it.TotalTokens),
		})
	}
	if err := w.rowTable([]string{"TRACE_ID", "NAME", "STATUS", "STARTED_AT", "LATENCY_MS", "SPANS", "TOKENS"}, rows); err != nil {
		return err
	}
	return w.footer(&t.Pagination)
}

func (w *Writer) sessionListTable(t models.ListResult[models.SessionSummary]) error {
	rows := make([][]string, 0, len(t.Items))
	for _, it := range t.Items {
		rows = append(rows, []string{
			it.SessionID,
			strconv.Itoa(it.TraceCount),
			it.FirstTraceAt,
			boolCell(it.HasError),
			fmtFloatPtr(it.TotalLatencyMs),
			strconv.Itoa(it.TotalTokens),
		})
	}
	if err := w.rowTable([]string{"SESSION_ID", "TRACES", "FIRST_TRACE_AT", "HAS_ERROR", "LATENCY_MS", "TOKENS"}, rows); err != nil {
		return err
	}
	return w.footer(&t.Pagination)
}

func (w *Writer) spanTable(spans []models.SpanResponse) error {
	rows := make([][]string, 0, len(spans))
	for _, s := range spans {
		rows = append(rows, []string{
			s.SpanID,
			truncate(s.Name, 36),
			string(s.Kind),
			w.statusCell(string(s.Status)),
			deref(s.Model),
			fmtFloatPtr(s.LatencyMs),
		})
	}
	return w.rowTable([]string{"SPAN_ID", "NAME", "KIND", "STATUS", "MODEL", "LATENCY_MS"}, rows)
}

func (w *Writer) metricTable(ms []models.MetricSummary) error {
	rows := make([][]string, 0, len(ms))
	for _, m := range ms {
		rows = append(rows, []string{m.Name, m.Category, truncate(m.Description, 60)})
	}
	return w.rowTable([]string{"NAME", "CATEGORY", "DESCRIPTION"}, rows)
}

func (w *Writer) evalRunListTable(runs []models.EvalRunResponse, p *models.Pagination) error {
	rows := make([][]string, 0, len(runs))
	for _, r := range runs {
		rows = append(rows, []string{
			r.ID,
			truncate(deref(r.Name), 30),
			r.TargetType,
			w.statusCell(string(r.Status)),
			strconv.Itoa(r.EvaluatedCount) + "/" + strconv.Itoa(r.TotalTargets),
			strconv.Itoa(r.FailedCount),
			r.CreatedAt,
		})
	}
	if err := w.rowTable([]string{"RUN_ID", "NAME", "TARGET", "STATUS", "EVALUATED", "FAILED", "CREATED_AT"}, rows); err != nil {
		return err
	}
	return w.footer(p)
}

func (w *Writer) traceScoreTable(scores []models.TraceScoreResponse, p *models.Pagination) error {
	rows := make([][]string, 0, len(scores))
	for _, s := range scores {
		rows = append(rows, []string{
			s.TraceID,
			s.Name,
			deref(s.Value),
			string(s.DataType),
			string(s.Source),
			w.statusCell(string(s.Status)),
		})
	}
	if err := w.rowTable([]string{"TRACE_ID", "NAME", "VALUE", "DATA_TYPE", "SOURCE", "STATUS"}, rows); err != nil {
		return err
	}
	return w.footer(p)
}

func (w *Writer) sessionScoreTable(scores []models.SessionScoreResponse, p *models.Pagination) error {
	rows := make([][]string, 0, len(scores))
	for _, s := range scores {
		rows = append(rows, []string{
			s.SessionID,
			s.Name,
			deref(s.Value),
			s.DataType,
			s.Source,
			w.statusCell(s.Status),
		})
	}
	if err := w.rowTable([]string{"SESSION_ID", "NAME", "VALUE", "DATA_TYPE", "SOURCE", "STATUS"}, rows); err != nil {
		return err
	}
	return w.footer(p)
}

func (w *Writer) traceDetail(t *models.TraceResponse) error {
	rows := [][2]string{
		{"trace_id", t.TraceID},
		{"name", t.Name},
		{"status", w.statusCell(string(t.Status))},
		{"started_at", t.StartedAt},
		{"ended_at", deref(t.EndedAt)},
		{"session_id", deref(t.SessionID)},
		{"environment", deref(t.Environment)},
		{"tags", strings.Join(t.Tags, ",")},
		{"spans", strconv.Itoa(len(t.Spans))},
		{"total_tokens", strconv.Itoa(t.TotalTokens)},
		{"total_cost", strconv.FormatFloat(t.TotalCost, 'f', -1, 64)},
	}
	if err := w.kvTable(rows); err != nil {
		return err
	}
	if len(t.Spans) > 0 {
		if _, err := fmt.Fprintln(w.out); err != nil {
			return err
		}
		return w.spanTable(t.Spans)
	}
	return nil
}

func (w *Writer) sessionDetail(s *models.SessionDetail) error {
	rows := [][2]string{
		{"session_id", s.SessionID},
		{"trace_count", strconv.Itoa(s.TraceCount)},
		{"first_trace_at", s.FirstTraceAt},
		{"last_trace_at", deref(s.LastTraceAt)},
		{"has_error", boolCell(s.HasError)},
		{"total_tokens", strconv.Itoa(s.TotalTokens)},
		{"traces_returned", strconv.Itoa(len(s.Traces))},
	}
	if err := w.kvTable(rows); err != nil {
		return err
	}
	if len(s.Traces) > 0 {
		if _, err := fmt.Fprintln(w.out); err != nil {
			return err
		}
		items := make([]models.TraceListItem, 0, len(s.Traces))
		for _, tr := range s.Traces {
			items = append(items, models.TraceListItem{
				TraceID: tr.TraceID, Name: tr.Name, Status: tr.Status,
				StartedAt: tr.StartedAt, EndedAt: tr.EndedAt, SpanCount: len(tr.Spans),
				TotalTokens: tr.TotalTokens,
			})
		}
		return w.traceListTable(models.ListResult[models.TraceListItem]{Items: items})
	}
	return nil
}

func (w *Writer) evalRunDetail(r *models.EvalRunResponse) error {
	return w.kvTable([][2]string{
		{"id", r.ID},
		{"name", deref(r.Name)},
		{"status", w.statusCell(string(r.Status))},
		{"target_type", r.TargetType},
		{"metrics", strings.Join(r.MetricNames, ",")},
		{"total_targets", strconv.Itoa(r.TotalTargets)},
		{"evaluated_count", strconv.Itoa(r.EvaluatedCount)},
		{"failed_count", strconv.Itoa(r.FailedCount)},
		{"sampling_rate", strconv.FormatFloat(r.SamplingRate, 'f', -1, 64)},
		{"model", deref(r.Model)},
		{"created_at", r.CreatedAt},
		{"completed_at", deref(r.CompletedAt)},
		{"error_message", deref(r.ErrorMessage)},
	})
}

func traceScoreRows(s models.TraceScoreResponse) [][2]string {
	return [][2]string{
		{"id", s.ID},
		{"trace_id", s.TraceID},
		{"name", s.Name},
		{"value", deref(s.Value)},
		{"data_type", string(s.DataType)},
		{"source", string(s.Source)},
		{"status", string(s.Status)},
		{"reason", deref(s.Reason)},
		{"created_at", s.CreatedAt},
	}
}

func fmtFloatPtr(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}

func boolCell(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
