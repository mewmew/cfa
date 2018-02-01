package interval

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/graphism/exp/cfg"
)

func TestIntervals(t *testing.T) {
	golden := []struct {
		path string
		want [][]string
	}{
		{
			path: "testdata/sample.dot",
			want: [][]string{
				[]string{"B1", "B2", "B3", "B4", "B5"},
				[]string{"B6", "B7", "B8", "B9", "B10", "B11", "B12"},
				[]string{"B13", "B14", "B15"},
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
