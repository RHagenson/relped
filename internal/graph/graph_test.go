package graph_test

import (
	"testing"

	"github.com/rhagenson/relped/internal/graph"
)

func TestGraph(t *testing.T) {
	t.Run("Self-loop does not panic", func(t *testing.T) {
		g := graph.NewGraph([]string{"I1", "I2"})
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Adding a self-loop paniced with: %s", r)
			}
		}()
		g.AddPath(graph.NewEqualWeightPath([]string{"I1", "I1"}, 1))
	})
	t.Run("Bowtie pattern is removed", func(t *testing.T) {
		// Bowtie:
		//     Dam->O1
		//     Sire->O1
		//     Dam->O2
		//     Sire->O2
		//     O1<->O1 // This connection should not persist after pruning
		g := graph.NewGraph([]string{"Dam", "Sire", "O1", "O2"})
		g.AddPath(graph.NewEqualWeightPath([]string{"Dam", "O1"}, 1))
		g.AddPath(graph.NewEqualWeightPath([]string{"Sire", "O1"}, 1))
		g.AddPath(graph.NewEqualWeightPath([]string{"Dam", "O2"}, 1))
		g.AddPath(graph.NewEqualWeightPath([]string{"Sire", "O2"}, 1))
		g.AddPath(graph.NewEqualWeightPath([]string{"Dam", "O1"}, 1))
		g.AddPath(graph.NewEqualWeightPath([]string{"O1", "O2"}, 1)) // Should be removed via g.Prune()

		g.AddDam("O1", "Dam")
		g.AddDam("O2", "Dam")
		g.AddSire("O1", "Sire")
		g.AddSire("O2", "Sire")
		g.Prune()

		o1, _ := g.NameToID("O1")
		o2, _ := g.NameToID("O2")
		if g.HasEdgeBetween(o1, o2) {
			t.Errorf("Bowtie was not removed. Offspring with the same parents remained connected:\n%s", g.String())
		}
	})
}
