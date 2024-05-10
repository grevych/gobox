//go:build !gobox_e2e

package trace_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grevych/gobox/pkg/app"
	"github.com/grevych/gobox/pkg/differs"
	"github.com/grevych/gobox/pkg/events"
	"github.com/grevych/gobox/pkg/trace"
	"github.com/grevych/gobox/pkg/trace/tracetest"
)

func TestForceTracingByHeader(t *testing.T) {
	defer app.SetName(app.Info().Name)
	app.SetName("gobox")

	recorder := tracetest.NewSpanRecorderWithOptions(tracetest.Options{
		SamplePercent: 0.1,
	})

	state := propagationInitRoundTripperState(t, recorder)
	defer state.Close()

	ctx := trace.StartSpan(context.Background(), "trace-test")

	client := http.Client{Transport: trace.NewTransport(nil)}
	req, err := http.NewRequestWithContext(ctx, "GET", state.Server.URL+"/myendpoint", http.NoBody)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	req.Header.Set(trace.HeaderForceTracing, "true")

	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	defer res.Body.Close()

	trace.End(ctx)

	traceID := trace.ID(ctx)
	rootID := differs.CaptureString()

	expected := []map[string]interface{}{
		{
			"name":                                 "ep",
			"spanContext.traceID":                  traceID,
			"spanContext.spanID":                   differs.AnyString(),
			"spanContext.traceFlags":               "01",
			"parent.traceID":                       traceID,
			"parent.spanID":                        rootID,
			"parent.traceFlags":                    "00",
			"parent.remote":                        true,
			"spanKind":                             "server",
			"startTime":                            differs.AnyString(),
			"endTime":                              differs.AnyString(),
			"attributes.force_trace":               "true",
			"attributes.net.host.name":             "127.0.0.1",
			"attributes.net.host.port":             differs.AnyInt64(),
			"attributes.net.sock.peer.addr":        "127.0.0.1",
			"attributes.net.sock.peer.port":        differs.AnyInt64(),
			"attributes.app.name":                  "gobox",
			"attributes.service_name":              "gobox",
			"attributes.app.version":               "testing",
			"attributes.duration":                  differs.FloatRange(0, 30),
			"attributes.http.method":               "GET",
			"attributes.http.flavor":               "1.1",
			"attributes.http.referer":              "",
			"attributes.http.request_id":           "",
			"attributes.http.scheme":               "http",
			"attributes.http.status_code":          int64(200),
			"attributes.http.url_details.endpoint": "ep",
			"attributes.http.url_details.path":     "/myendpoint",
			"attributes.http.url_details.uri":      "/myendpoint",
			"attributes.http.user_agent":           "Go-http-client/1.1",
			"attributes.network.bytes_read":        int64(0),
			"attributes.network.bytes_written":     int64(2),
			"attributes.network.client.ip":         "",
			"attributes.network.destination.ip":    "",
			"attributes.timing.dequeued_at":        differs.RFC3339NanoTime(),
			"attributes.timing.finished_at":        differs.RFC3339NanoTime(),
			"attributes.timing.scheduled_at":       differs.RFC3339NanoTime(),
			"attributes.timing.service_time":       differs.FloatRange(0, 30),
			"attributes.timing.total_time":         differs.FloatRange(0, 30),
			"attributes.timing.wait_time":          differs.FloatRange(0, 30),
			"SampleRate":                           int64(1),
		},
	}

	ev := recorder.Ended()
	if diff := cmp.Diff(expected, ev, differs.Custom()); diff != "" {
		t.Fatal("unexpected events", diff)
	}
}

func TestHeadersForceTracingByHeader(t *testing.T) {
	defer app.SetName(app.Info().Name)
	app.SetName("gobox")

	recorder := tracetest.NewSpanRecorderWithOptions(tracetest.Options{
		SamplePercent: 0.1,
	})
	defer recorder.Close()

	header := http.Header{}

	header.Set(trace.HeaderForceTracing, "true")

	ctx := trace.FromHeaders(context.Background(), header, "trace-test")

	traceID := trace.ID(ctx)
	rootID := differs.CaptureString()

	inner := trace.StartSpan(ctx, "inner")

	trace.End(inner)
	trace.End(ctx)

	expected := []map[string]interface{}{
		{
			"SampleRate":              int64(1),
			"name":                    "inner",
			"spanContext.traceID":     traceID,
			"spanContext.spanID":      differs.AnyString(),
			"spanContext.traceFlags":  "01",
			"parent.traceID":          traceID,
			"parent.spanID":           rootID,
			"parent.traceFlags":       "01",
			"parent.remote":           false,
			"spanKind":                "internal",
			"startTime":               differs.AnyString(),
			"endTime":                 differs.AnyString(),
			"attributes.app.name":     "gobox",
			"attributes.app.version":  "testing",
			"attributes.service_name": "gobox",
		},
		{
			"SampleRate":              int64(1),
			"name":                    "trace-test",
			"spanContext.traceID":     traceID,
			"spanContext.spanID":      rootID,
			"spanContext.traceFlags":  "01",
			"parent.traceID":          "00000000000000000000000000000000",
			"parent.spanID":           "0000000000000000",
			"parent.traceFlags":       "00",
			"parent.remote":           false,
			"spanKind":                "internal",
			"startTime":               differs.AnyString(),
			"endTime":                 differs.AnyString(),
			"attributes.app.name":     "gobox",
			"attributes.app.version":  "testing",
			"attributes.service_name": "gobox",
		},
	}

	ev := recorder.Ended()
	if diff := cmp.Diff(expected, ev, differs.Custom()); diff != "" {
		t.Fatal("unexpected events", diff)
	}
}

func TestForceTracing(t *testing.T) {
	defer app.SetName(app.Info().Name)
	app.SetName("gobox")

	recorder := tracetest.NewSpanRecorderWithOptions(tracetest.Options{
		SamplePercent: 0.1,
	})

	state := propagationInitRoundTripperState(t, recorder)
	defer state.Close()

	ctx := trace.ForceTracing(context.Background())
	ctx = trace.StartSpan(ctx, "trace-test")

	client := http.Client{Transport: trace.NewTransport(nil)}
	req, err := http.NewRequestWithContext(ctx, "GET", state.Server.URL+"/myendpoint", http.NoBody)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	defer res.Body.Close()

	trace.End(ctx)

	traceID := trace.ID(ctx)

	expected := []map[string]interface{}{
		{
			"name":                                 "ep",
			"spanContext.traceID":                  traceID,
			"spanContext.spanID":                   differs.AnyString(),
			"spanContext.traceFlags":               "01",
			"parent.traceID":                       traceID,
			"parent.spanID":                        differs.AnyString(),
			"parent.traceFlags":                    "01",
			"parent.remote":                        true,
			"spanKind":                             "server",
			"startTime":                            differs.AnyString(),
			"endTime":                              differs.AnyString(),
			"attributes.force_trace":               "true",
			"attributes.net.sock.peer.addr":        "127.0.0.1",
			"attributes.net.sock.peer.port":        differs.AnyInt64(),
			"attributes.net.host.name":             "127.0.0.1",
			"attributes.net.host.port":             differs.AnyInt64(),
			"attributes.app.name":                  "gobox",
			"attributes.service_name":              "gobox",
			"attributes.app.version":               "testing",
			"attributes.duration":                  differs.FloatRange(0, 30),
			"attributes.http.flavor":               "1.1",
			"attributes.http.method":               "GET",
			"attributes.http.referer":              "",
			"attributes.http.request_id":           "",
			"attributes.http.scheme":               "http",
			"attributes.http.status_code":          int64(200),
			"attributes.http.url_details.endpoint": "ep",
			"attributes.http.url_details.path":     "/myendpoint",
			"attributes.http.url_details.uri":      "/myendpoint",
			"attributes.http.user_agent":           "Go-http-client/1.1",
			"attributes.network.bytes_read":        int64(0),
			"attributes.network.bytes_written":     int64(2),
			"attributes.network.client.ip":         "",
			"attributes.network.destination.ip":    "",
			"attributes.timing.dequeued_at":        differs.RFC3339NanoTime(),
			"attributes.timing.finished_at":        differs.RFC3339NanoTime(),
			"attributes.timing.scheduled_at":       differs.RFC3339NanoTime(),
			"attributes.timing.service_time":       differs.FloatRange(0, 30),
			"attributes.timing.total_time":         differs.FloatRange(0, 30),
			"attributes.timing.wait_time":          differs.FloatRange(0, 30),
			"SampleRate":                           int64(1),
		},
		{
			"name":                    "trace-test",
			"spanContext.traceID":     traceID,
			"spanContext.spanID":      differs.AnyString(),
			"spanContext.traceFlags":  "01",
			"parent.traceID":          "00000000000000000000000000000000",
			"parent.spanID":           "0000000000000000",
			"parent.traceFlags":       "00",
			"parent.remote":           false,
			"spanKind":                "internal",
			"startTime":               differs.AnyString(),
			"endTime":                 differs.AnyString(),
			"attributes.app.name":     "gobox",
			"attributes.service_name": "gobox",
			"attributes.app.version":  "testing",
			"SampleRate":              int64(1),
		},
	}

	ev := recorder.Ended()
	if diff := cmp.Diff(expected, ev, differs.Custom()); diff != "" {
		t.Fatal("unexpected events", diff)
	}
}

func propagationInitRoundTripperState(t *testing.T, recorder *tracetest.SpanRecorder) *roundtripperState {
	t.Helper()
	server := httptest.NewServer(
		trace.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace.StartSpan(r.Context(), "ep")
			defer trace.End(r.Context())

			var info events.HTTPRequest
			info.FillFieldsFromRequest(r)
			info.Endpoint = "ep"
			if n, err := w.Write([]byte("OK")); err != nil {
				t.Fatal("Got error", err)
			} else {
				info.FillResponseInfo(n, 200)
			}
			trace.AddInfo(r.Context(), &info)
		}), "ep"))

	return &roundtripperState{recorder, server}
}
