// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Provides static app info

// Package app has the static app info
package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

// defaultVersion is the default version string
const defaultVersion = "Please see http://github.com/grevych/gobox/blob/main/docs/version.md"

// Version needs to be set at build time using
// -ldflags "-X github.com/grevych/gobox/pkg/app.Version=something"
//
//nolint:gochecknoglobals // Why: For linking
var Version = defaultVersion

// appName is the name of the app
//
//nolint:gochecknoglobals // Why: For linking
var appName = "unknown"

// appInfo contains information about the app
var appInfo struct {
	mu sync.Mutex // guarding Data to be set initialized concurrently
	*Data
}

// Info returns the static app info
//
// This struct is used mainly to provide tags to append to logs.  It's also used
// by a handful of infrastructure-y packages like Mint or orgservice that have
// special needs.  Most services will never need to access it directly.
func Info() *Data {
	appInfo.mu.Lock()
	defer appInfo.mu.Unlock()

	if appInfo.Data == nil {
		appInfo.Data = info()
	}
	return appInfo.Data
}

// info returns the static app info
//
//nolint:funlen // Why: cleaner to keep everything together
func info() *Data {
	const unknown = "unknown"
	mainModule := ""
	namespace := ""
	serviceAccount := ""
	bento := ""

	ver := Version
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		mainModule = buildInfo.Main.Path

		// allow people to still link the version
		if ver == defaultVersion {
			ver = buildInfo.Main.Version
		}
	}

	// The namespace and service account env vars are set by bootstrap
	// generated deployment scripts.
	if ns := os.Getenv("MY_NAMESPACE"); ns != "" {
		namespace = ns
	}

	if sa := os.Getenv("MY_POD_SERVICE_ACCOUNT"); sa != "" {
		serviceAccount = sa
	}

	// There is no guarantee that this correlation between `app.Name` and
	// ServiceID will exist forever.  For example, in the future we could
	// have several apps sharing the same ServiceID.  But that's not
	// supportd by bootstrap yet and so this hard-coded assumption works
	// well enough for now.
	serviceID := fmt.Sprintf("%s@gobox.cloud", appName)

	parts := strings.Split(namespace, "--")
	if len(parts) == 2 {
		bento = parts[1]
	}

	environment := unknown
	if env := os.Getenv("MY_ENVIRONMENT"); env != "" {
		environment = env
	}

	clusterName := unknown
	if cn := os.Getenv("MY_CLUSTER"); cn != "" {
		clusterName = cn
	}

	region := unknown
	if r := os.Getenv("MY_REGION"); r != "" {
		region = r
	} else if rps := strings.Split(clusterName, "."); len(rps) == 2 {
		// e.g. production.us-west-2
		region = rps[1]
	}

	podID := unknown
	if pi := os.Getenv("MY_POD_NAME"); pi != "" {
		podID = pi
	}

	nodeID := unknown
	if ni := os.Getenv("MY_NODE_NAME"); ni != "" {
		nodeID = ni
	}

	deployment := unknown
	if d := os.Getenv("MY_DEPLOYMENT"); d != "" {
		deployment = d
	}

	return &Data{
		Name:    appName,
		Version: ver,

		MainModule: mainModule,

		Environment:    environment,
		Namespace:      namespace,
		ServiceAccount: serviceAccount,
		ClusterName:    clusterName,
		Region:         region,
		PodID:          podID,
		NodeID:         nodeID,
		Deployment:     deployment,

		ServiceID: serviceID,

		Bento: bento,
	}
}

// SetName sets the app name
//
// Should only be called from tests and app initialization
func SetName(name string) {
	appInfo.mu.Lock()
	defer appInfo.mu.Unlock()

	appName = name
	appInfo.Data = info()
}

// Data provides the global app info
type Data struct {
	mu sync.Mutex // Just for the log marshaler

	Name    string
	Version string

	MainModule string

	Environment    string
	Namespace      string
	ServiceAccount string
	ClusterName    string
	Region         string
	PodID          string
	NodeID         string
	Deployment     string

	// ServiceID is a unique identifier for the service used primarily
	// within the `authn` framework.
	ServiceID string

	Bento string
}

// MarshalLog marshals the struct for logging
func (d *Data) MarshalLog(addField func(key string, v interface{})) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Name != "unknown" {
		addField("app.name", d.Name)
		addField("service_name", d.Name)
	}
	if d.Version != "" {
		addField("app.version", d.Version)
	}
	if d.Namespace != "" {
		addField("deployment.namespace", d.Namespace)
	}
}

// LogValue implements the log/slog package's LogValuer interface (found
// here: https://pkg.go.dev/log/slog#LogValuer). Returns a subset of the
// App.Info data as a map.
func (d *Data) LogValue() slog.Value {
	d.mu.Lock()
	defer d.mu.Unlock()

	attrs := make([]slog.Attr, 0, 4)

	// App prefixes are removed as AppInfo func nests data under app key
	// already.
	if d.Name != "unknown" {
		attrs = append(
			attrs,
			slog.String("name", d.Name),
			slog.String("service_name", d.Name),
		)
	}
	if d.Version != "" {
		attrs = append(attrs, slog.String("version", d.Version))
	}
	if d.Namespace != "" {
		attrs = append(attrs, slog.String("deployment.namespace", d.Namespace))
	}

	return slog.GroupValue(attrs...)
}

// LogHook provides an olog compatible hook func which extracts and returns
// the app Data as a nested attribute on log record.
// nolint:gocritic // Why: this signature is inline with the olog pkg hook type
func LogHook(ctx context.Context, r slog.Record) ([]slog.Attr, error) {
	info := Info()
	if info == nil {
		return []slog.Attr{}, nil
	}

	return []slog.Attr{
		// Manually assign the LogValue to an attribute. The slog.Group
		// func doesn't really take an already existing GroupValue as
		// Values are more meant to be used directly in log calls.
		{Key: "app", Value: info.LogValue()},
	}, nil
}
