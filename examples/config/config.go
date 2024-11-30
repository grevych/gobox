package main

import "github.com/grevych/gobox/pkg/cfg"

type config struct {
}

func main() {
	c := &config{}
	// config.yaml is expected to be at cwd level
	if err := cfg.Load("config.yaml", c); err != nil {
		panic(err)
	}
}
