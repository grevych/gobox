package main

import "github.com/grevych/gobox/pkg/cfg"

type config struct {
}

func main() {
	c := &config{}
	cfg.Load("config.yaml", c)
}
