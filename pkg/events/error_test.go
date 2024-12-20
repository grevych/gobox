//go:build !gobox_e2e

package events_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grevych/gobox/pkg/differs"
	"github.com/grevych/gobox/pkg/events"
	"github.com/grevych/gobox/pkg/log"
	"github.com/pkg/errors"
)

func TestErrorInfo(t *testing.T) {
	err := errors.New("test error")
	info := events.NewErrorInfo(err)
	got := map[string]interface{}{}
	info.MarshalLog(func(key string, v interface{}) {
		got[key] = v
	})
	if got["error.message"] != "test error" {
		t.Fatal("Unexpected message", got["error.message"])
	}

	expectedStack := []string{
		"gobox/pkg/events/error_test.go:18 `events_test.TestErrorInfo`",
	}

	errorStack, _ := got["error.stack"].(string)
	stack := []string{errorStack}
	if diff := differs.StackTrace(strings.Join(expectedStack, "\n"), strings.Join(stack, "\n")); diff != "" {
		t.Fatal("got unexpected stack", stack, "\n", diff)
	}
}

func TestErrorInfoCollapse(t *testing.T) {
	err := errors.Wrap(errors.New("test error"), "context")
	info := events.NewErrorInfo(err)
	got := map[string]interface{}{}
	info.Cause.MarshalLog(func(key string, v interface{}) {
		got[key] = v
	})

	// with collapse behavior, both message and stack should be set.
	want := map[string]interface{}{
		"kind":    "cause",
		"stack":   differs.StackLike("gobox/pkg/events/error_test.go:39 `events_test.TestErrorInfoCollapse`"),
		"message": "test error",
	}
	if diff := cmp.Diff(want, got, differs.Custom()); diff != "" {
		t.Error("custom error mismatched", diff)
	}
}

func TestErrorRecoveryPanicNonError(t *testing.T) {
	func() {
		defer func() {
			info := events.NewErrorInfoFromPanic(recover())
			got := map[string]interface{}{}
			info.MarshalLog(func(key string, v interface{}) {
				got[key] = v
			})
			if got["error.message"] != "42" {
				t.Fatal("Unexpected message", got["error.message"])
			}
			if s, _ := got["error.stack"].(string); s == "" {
				t.Fatal("no stack?", got)
			}
		}()

		panic(42)
	}()
}

func TestErrorRecoveryPanicError(t *testing.T) {
	err := errors.New("test error")
	func() {
		defer func() {
			info := events.NewErrorInfoFromPanic(recover())
			if info.Kind != "error" || info.Message != "test error" {
				t.Fatal("Unexpected info", info)
			}
		}()

		panic(err)
	}()
}

type customError struct{}

func (c customError) Error() string {
	return "custom error!"
}
func (c customError) MarshalLog(addField func(k string, v interface{})) {
	addField("error.kind", "custom kind")
	addField("error.stack", "custom stack")
	addField("error.message", "custom message")
	addField("other_field", "custom field")
}

func TestCustomErrorLoginfo(t *testing.T) {
	info := events.NewErrorInfo(customError{})
	if info.Custom == nil {
		t.Fatal("no custom marshaler")
	}

	got := map[string]interface{}{}
	info.MarshalLog(func(key string, v interface{}) {
		got[key] = v
	})
	want := map[string]interface{}{
		"error.kind":    "custom kind",
		"error.stack":   "custom stack",
		"error.error":   "custom error!",
		"error.message": "custom message",
		"other_field":   "custom field",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error("custom error mismatched", diff)
	}
}

func TestNestedErrorLogInfo(t *testing.T) {
	inner := errors.New("inner error")
	outer := fmt.Errorf("outer error: %w", inner)

	info := events.NewErrorInfo(outer)

	got := map[string]interface{}{}
	var addField func(key string, v interface{})
	addField = func(key string, v interface{}) {
		if v == nil {
			return
		}

		m, ok := v.(log.Marshaler)
		if !ok {
			got[key] = v
			return
		}
		m.MarshalLog(func(inner string, v interface{}) {
			addField(key+"."+inner, v)
		})
	}
	info.MarshalLog(addField)

	want := map[string]interface{}{
		"error.kind":          "error",
		"error.error":         "outer error: inner error",
		"error.message":       "outer error",
		"error.cause.kind":    "cause",
		"error.cause.message": "inner error",
		"error.cause.stack":   differs.StackLike("gobox/pkg/events/error_test.go:000 `events_test.TestNestedErrorLogInfo`"),
	}
	if diff := cmp.Diff(want, got, differs.Custom()); diff != "" {
		t.Error("custom error mismatched", diff)
	}
}
