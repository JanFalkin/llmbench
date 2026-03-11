package report

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"testing"
	"time"

	"github.com/JanFalkin/llmbench/internal/stats"
)

func TestRenderBenchmarkCSV_IncludesMetaColumnsAndValues(t *testing.T) {
	var rep stats.BenchmarkReport
	cfg := configValue(t, &rep)

	if !setOneOfStringField(cfg, []string{"Model", "ModelName"}, "gpt-4o-mini") {
		t.Fatalf("config missing model field")
	}
	if !setOneOfStringField(cfg, []string{"URL", "Url", "Endpoint", "BaseURL", "BaseUrl"}, "https://api.example/v1/chat/completions") {
		t.Fatalf("config missing url field")
	}
	expectedLabel := ""
	if setOneOfStringField(cfg, []string{"Label", "RunLabel"}, "nightly-a") {
		expectedLabel = "nightly-a"
	}

	setSingleResult(t, &rep, map[string]any{
		"RequestID":    "req-1",
		"HTTPStatus":   200,
		"InputTokens":  12,
		"OutputTokens": 34,
		"EndToEnd":     150 * time.Millisecond,
		"TTFT":         40 * time.Millisecond,
		"Decode":       90 * time.Millisecond,
		"Error":        "",
	})

	out, err := RenderBenchmarkCSV(rep)
	if err != nil {
		t.Fatalf("RenderBenchmarkCSV error: %v", err)
	}

	records := mustReadCSV(t, out)
	if len(records) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(records))
	}

	header := records[0]
	row := records[1]

	wantPrefix := []string{"model", "url", "label"}
	if !reflect.DeepEqual(header[:3], wantPrefix) {
		t.Fatalf("unexpected header prefix: got %v want %v", header[:3], wantPrefix)
	}
	if row[0] != "gpt-4o-mini" || row[1] != "https://api.example/v1/chat/completions" || row[2] != expectedLabel {
		t.Fatalf("unexpected metadata values: %v", row[:3])
	}
}

func TestRenderSweepCSV_IncludesMetaColumnsAndValues(t *testing.T) {
	var rep stats.BenchmarkReport
	cfg := configValue(t, &rep)

	if !setOneOfStringField(cfg, []string{"Model", "ModelName"}, "gpt-4.1") {
		t.Fatalf("config missing model field")
	}
	if !setOneOfStringField(cfg, []string{"URL", "Url", "Endpoint", "BaseURL", "BaseUrl"}, "https://api.example/v1/chat/completions") {
		t.Fatalf("config missing url field")
	}
	expectedLabel := ""
	if setOneOfStringField(cfg, []string{"Label", "RunLabel"}, "batch-42") {
		expectedLabel = "batch-42"
	}
	setIntField(t, cfg, "Concurrency", 8)

	setIntField(t, reflect.ValueOf(&rep).Elem(), "TotalRequests", 100)
	setIntField(t, reflect.ValueOf(&rep).Elem(), "SuccessfulRequests", 99)
	setIntField(t, reflect.ValueOf(&rep).Elem(), "FailedRequests", 1)
	setDurationField(t, reflect.ValueOf(&rep).Elem(), "Elapsed", 10*time.Second)
	setFloatField(t, reflect.ValueOf(&rep).Elem(), "RequestsPerSecond", 9.9)
	setFloatField(t, reflect.ValueOf(&rep).Elem(), "OutputTokensPerSec", 123.456)
	setDurationField(t, reflect.ValueOf(&rep).Elem(), "AvgLatency", 200*time.Millisecond)
	setDurationField(t, reflect.ValueOf(&rep).Elem(), "LatencyP50", 180*time.Millisecond)
	setDurationField(t, reflect.ValueOf(&rep).Elem(), "LatencyP95", 300*time.Millisecond)
	setDurationField(t, reflect.ValueOf(&rep).Elem(), "TTFTP50", 70*time.Millisecond)
	setDurationField(t, reflect.ValueOf(&rep).Elem(), "TTFTP95", 120*time.Millisecond)

	out, err := RenderSweepCSV([]stats.BenchmarkReport{rep})
	if err != nil {
		t.Fatalf("RenderSweepCSV error: %v", err)
	}

	records := mustReadCSV(t, out)
	if len(records) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(records))
	}

	header := records[0]
	row := records[1]

	wantPrefix := []string{"model", "url", "label"}
	if !reflect.DeepEqual(header[:3], wantPrefix) {
		t.Fatalf("unexpected header prefix: got %v want %v", header[:3], wantPrefix)
	}
	if row[0] != "gpt-4.1" || row[1] != "https://api.example/v1/chat/completions" || row[2] != expectedLabel {
		t.Fatalf("unexpected metadata values: %v", row[:3])
	}
}

func mustReadCSV(t *testing.T, b []byte) [][]string {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(b))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	return records
}

func configValue(t *testing.T, rep *stats.BenchmarkReport) reflect.Value {
	t.Helper()
	v := reflect.ValueOf(rep).Elem().FieldByName("Config")
	if !v.IsValid() {
		t.Fatalf("BenchmarkReport missing Config field")
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		t.Fatalf("Config is not a struct")
	}
	return v
}

func setOneOfStringField(v reflect.Value, names []string, val string) bool {
	for _, name := range names {
		f := v.FieldByName(name)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			f.SetString(val)
			return true
		}
	}
	return false
}

func setIntField(t *testing.T, v reflect.Value, name string, val int) {
	t.Helper()
	f := v.FieldByName(name)
	if !f.IsValid() || !f.CanSet() {
		t.Fatalf("missing/settable int field %q", name)
	}
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.SetInt(int64(val))
	default:
		t.Fatalf("field %q is not int-kind", name)
	}
}

func setFloatField(t *testing.T, v reflect.Value, name string, val float64) {
	t.Helper()
	f := v.FieldByName(name)
	if !f.IsValid() || !f.CanSet() {
		t.Fatalf("missing/settable float field %q", name)
	}
	if f.Kind() != reflect.Float32 && f.Kind() != reflect.Float64 {
		t.Fatalf("field %q is not float-kind", name)
	}
	f.SetFloat(val)
}

func setDurationField(t *testing.T, v reflect.Value, name string, d time.Duration) {
	t.Helper()
	f := v.FieldByName(name)
	if !f.IsValid() || !f.CanSet() {
		t.Fatalf("missing/settable duration field %q", name)
	}
	// time.Duration is int64
	if f.Kind() != reflect.Int64 {
		t.Fatalf("field %q is not duration/int64-kind", name)
	}
	f.SetInt(int64(d))
}

func setSingleResult(t *testing.T, rep *stats.BenchmarkReport, values map[string]any) {
	t.Helper()

	rv := reflect.ValueOf(rep).Elem()
	results := rv.FieldByName("Results")
	if !results.IsValid() || results.Kind() != reflect.Slice {
		t.Fatalf("BenchmarkReport missing Results slice")
	}

	elemType := results.Type().Elem()
	elem := reflect.New(elemType).Elem()

	for k, val := range values {
		f := elem.FieldByName(k)
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		switch x := val.(type) {
		case string:
			if f.Kind() == reflect.String {
				f.SetString(x)
			}
		case int:
			if f.Kind() >= reflect.Int && f.Kind() <= reflect.Int64 {
				f.SetInt(int64(x))
			}
		case time.Duration:
			if f.Kind() == reflect.Int64 {
				f.SetInt(int64(x))
			}
		}
	}

	s := reflect.MakeSlice(results.Type(), 1, 1)
	s.Index(0).Set(elem)
	results.Set(s)
}
