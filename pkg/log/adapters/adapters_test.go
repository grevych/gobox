//go:build !gobox_e2e

package adapters_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/grevych/gobox/pkg/log"
	"github.com/grevych/gobox/pkg/log/adapters"
	"github.com/grevych/gobox/pkg/log/logtest"
)

func printEntries(entries []log.F) {
	for _, entry := range entries {
		entry["@timestamp"] = "2021-12-21T14:19:20.0424249-08:00"
		bytes, err := json.Marshal(entry)
		if err != nil {
			fmt.Println("unexpected", err)
		} else {
			fmt.Println(string(bytes))
		}
	}
}

func ExampleNewLogrLogger() {
	logs := logtest.NewLogRecorder(nil)
	defer logs.Close()

	logger := adapters.NewLogrLogger(context.Background())

	logger.Info(strconv.FormatBool(logger.Enabled()))

	logger.Info("hello, world", "a", 1)

	childLogger := logger.WithValues("c", 1, "b", "hello, world!")
	childLogger.Info("info!!")
	childLogger.Error(fmt.Errorf("bad thing"), "end of the world!")

	printEntries(logs.Entries())

	//nolint:lll // Why: testing output
	// Output:
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","app.version":"testing","level":"INFO","message":"true","module":"github.com/grevych/gobox"}
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","a":1,"app.version":"testing","level":"INFO","message":"hello, world","module":"github.com/grevych/gobox"}
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","app.version":"testing","b":"hello, world!","c":1,"level":"INFO","message":"info!!","module":"github.com/grevych/gobox"}
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","app.version":"testing","b":"hello, world!","c":1,"error.error":"bad thing","error.kind":"error","error.message":"bad thing","level":"ERROR","message":"end of the world!","module":"github.com/grevych/gobox"}
}

func ExampleNewRetryableHTTPLogger() {
	logs := logtest.NewLogRecorder(nil)
	defer logs.Close()

	logger := adapters.NewRetryableHTTPLogger(context.Background())

	logger.Info("hello, info", "a", 1)
	logger.Debug("hello, debug", "a", 1)
	logger.Error("hello, error", "a", 1)
	logger.Warn("hello, warn", "a", 1)

	printEntries(logs.Entries())

	//nolint:lll // Why: testing output
	// Output:
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","a":1,"app.version":"testing","level":"INFO","message":"hello, info","module":"github.com/grevych/gobox"}
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","a":1,"app.version":"testing","level":"DEBUG","message":"hello, debug","module":"github.com/grevych/gobox"}
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","a":1,"app.version":"testing","level":"ERROR","message":"hello, error","module":"github.com/grevych/gobox"}
	// {"@timestamp":"2021-12-21T14:19:20.0424249-08:00","a":1,"app.version":"testing","level":"WARN","message":"hello, warn","module":"github.com/grevych/gobox"}
}
