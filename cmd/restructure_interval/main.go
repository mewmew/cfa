package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/graphism/exp/cfg"
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
	buf, err := json.MarshalIndent(prims, "", "\t")
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Println(string(buf))
	return nil
}
