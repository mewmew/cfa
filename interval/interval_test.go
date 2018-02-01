package interval

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/graphism/exp/cfg"
)

func TestIntervals(t *testing.T) {
	golden := []struct {
		path string
		want [][]string
	}{
		{
			path: "testdata/structuring_decompiled_graphs_figure_2.dot",
			want: [][]string{
				[]string{"B1", "B2", "B3", "B4", "B5"},
				[]string{"B6", "B7", "B8", "B9", "B10", "B11", "B12"},
				[]string{"B13", "B14", "B15"},
			},
		},
		{
			path: "testdata/control_flow_analysis_figure_2.dot",
			want: [][]string{
				[]string{"1"},
				[]string{"2"},
				[]string{"3", "4", "5", "6"},
				[]string{"7", "8"},
			},
		},
	}
	for _, gold := range golden {
		// Parse input.
		in, err := cfg.ParseFile(gold.path)
		if err != nil {
			t.Errorf("%q; unable to parse file; %v", gold.path, err)
			continue
		}
		// Locate intervals.
		intervals := Intervals(in)
		if len(intervals) != len(gold.want) {
			t.Errorf("%q: number of intervals mismatch; expected %d, got %d", gold.path, len(gold.want), len(intervals))
			continue
		}
		for i, want := range gold.want {
			var got []string
			// TODO: Update test to randomize node order. Then make sure the
			// intervals are calculated independent of what g.Nodes() returns. Use
			// reverse post-order.
			for _, n := range intervals[i].Nodes() {
				nn, ok := n.(*cfg.Node)
				if !ok {
					panic(fmt.Errorf("invalid node type; expected *cfg.Node, got %T", n))
				}
				got = append(got, nn.DOTID())
			}
			sort.Strings(got)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%q; output mismatch; expected `%s`, got `%s`", gold.path, want, got)
				continue
			}
		}
	}
}

func TestDerivedSeq(t *testing.T) {
	golden := []struct {
		path string
		want []string
	}{
		{
			path: "testdata/structuring_decompiled_graphs_figure_2.dot",
			want: []string{
				"testdata/structuring_decompiled_graphs_figure_2.dot.G1.golden",
				"testdata/structuring_decompiled_graphs_figure_2.dot.G2.golden",
				"testdata/structuring_decompiled_graphs_figure_2.dot.G3.golden",
				"testdata/structuring_decompiled_graphs_figure_2.dot.G4.golden",
			},
		},
	}
	for _, gold := range golden {
		in, err := cfg.ParseFile(gold.path)
		if err != nil {
			t.Errorf("%q; unable to parse file; %v", gold.path, err)
			continue
		}
		gs, _ := DerivedSeq(in)
		if len(gs) != len(gold.want) {
			t.Errorf("%q: number of derived graphs mismatch; expected %d, got %d", gold.path, len(gold.want), len(gs))
			continue
		}
		for i, g := range gs {
			buf, err := ioutil.ReadFile(gold.want[i])
			if err != nil {
				t.Errorf("%q; unable to read file; %v", gold.path, err)
				continue
			}
			want := strings.TrimSpace(string(buf))
			got := g.String()
			if got != want {
				t.Errorf("%q; output mismatch; expected `%s`, got `%s`", gold.path, want, got)
				continue
			}
		}
	}
}
