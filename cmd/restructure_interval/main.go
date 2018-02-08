package main

import (
	"flag"
	"log"

	"github.com/graphism/exp/cfg"
	"github.com/kr/pretty"
	"github.com/mewmew/cfa/interval"
	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	for _, dotPath := range flag.Args() {
		if err := restructure(dotPath); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

func restructure(dotPath string) error {
	g, err := cfg.ParseFile(dotPath)
	if err != nil {
		return errors.WithStack(err)
	}
	interval.Analyze(g)
	pretty.Println("nodes:", cfg.SortByRevPost(g.Nodes()))
	return nil
}
