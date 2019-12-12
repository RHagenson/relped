package graph_test

import (
	"github.com/rhagenson/relped/internal/graph"
	"testing"
)

func TestGraph(t *testing.T) {
	g := graph.NewGraph([]string{"I1", "I2"})
	t.Run("Self-loop does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Adding a self-loop paniced with: %s", r)
			}
		}()
		g.AddPath(graph.NewEqualWeightPath([]string{"I1", "I1"}, 1))
	})
}
