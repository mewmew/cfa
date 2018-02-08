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
	prims := interval.Analyze(g)
	for _, prim := range prims {
		pretty.Println("prim:", prim)
	}
	/*
		for _, n := range cfg.SortByRevPost(g.Nodes()) {
			pretty.Println("node:", n)
			if n.LoopHead != nil {
				pretty.Println("   LoopHead:", n.LoopHead.DOTID())
			}
			if n.Latch != nil {
				pretty.Println("   Latch:", n.Latch.DOTID())
			}
			if n.LoopFollow != nil {
				pretty.Println("   LoopFollow:", n.LoopFollow.DOTID())
			}
		}
	*/
	return nil
}
