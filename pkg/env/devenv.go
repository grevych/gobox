//go:build gobox_dev

package env

import (
	"github.com/grevych/gobox/pkg/cfg"
)

func ApplyOverrides() {
	cfg.SetDefaultReader(devReader(cfg.DefaultReader()))
}
