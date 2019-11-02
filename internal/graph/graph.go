package graph

import (
	"log"

	"github.com/rhagenson/relped/internal/csvin"
	"github.com/rs/xid"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

const lenUnknownNames = 6

// Graph has named nodes/vertexes
type Graph struct {
	g *simple.WeightedUndirectedGraph
	m map[string]gonumGraph.Node
}

func NewGraph() *Graph {
	return &Graph{
		g: simple.NewWeightedUndirectedGraph(0, 0),
		m: make(map[string]gonumGraph.Node),
	}
}

func NewGraphFromCsvInput(in csvin.CsvInput, maxDist uint) *Graph {
	indvs := in.Indvs()
	g := NewGraph()

	// Add paths from node to node based on relational distance
	for i := 0; i < len(indvs); i++ {
		for j := i + 1; j < len(indvs); j++ {
			i1 := indvs[i]
			i2 := indvs[j]
			g.AddNode(i1)
			g.AddNode(i2)
			dist := in.RelDistance(i1, i2)
			weight := in.Relatedness(i1, i2)
			if dist <= maxDist {
				g.AddRelationalPath(i1, i2, dist, weight)
			}
		}
	}
	return g
}

func (self *Graph) PruneToShortest(indvs []string) *Graph {
	g := NewGraph()
	for i := 0; i < len(indvs); i++ {
		for j := i + 1; j < len(indvs); j++ {
			src := self.Node(indvs[i])
			dest := self.Node(indvs[j])
			if self.g.HasEdgeBetween(src.ID(), dest.ID()) { // Directly connected
				g.AddEqualWeightPath([]string{indvs[i], indvs[j]}, self.WeightedEdge(indvs[i], indvs[j]).Weight())
			} else { // Perhaps indirectly connected
				shortest, _ := path.BellmanFordFrom(src, self.g)
				nodes, cost := shortest.To(dest.ID())
				names := make([]string, len(nodes))
				for i, node := range nodes {
					names[i] = self.NameFromID(node.ID())
				}
				g.AddEqualWeightPath(names, cost)
			}
		}
	}
	return g
}

func (self *Graph) Nodes() gonumGraph.Nodes {
	return self.g.Nodes()
}

func (self *Graph) NameFromID(id int64) string {
	for name, node := range self.m {
		if node.ID() == id {
			return name
		}
	}
	return ""
}

func (self *Graph) RmDisconnected() {
	for name := range self.m {
		nodes := self.From(name)
		if nodes.Len() == 0 {
			self.RemoveNode(name)
		}
	}
}

func (self *Graph) Weight(xid, yid int64) (w float64, ok bool) {
	return self.g.Weight(xid, yid)
}

func (self *Graph) From(name string) gonumGraph.Nodes {
	if node, ok := self.m[name]; ok {
		return self.g.From(node.ID())
	}
	return nil
}

func (self *Graph) RemoveNode(name string) {
	if node, ok := self.m[name]; ok {
		self.g.RemoveNode(node.ID())
	}
}

func (self *Graph) AddNode(name string) {
	if _, ok := self.m[name]; !ok {
		n := self.g.NewNode()
		self.g.AddNode(n)
		self.m[name] = n
	}
}

func (self *Graph) Edge(n1, n2 string) gonumGraph.Edge {
	uid := self.m[n1].ID()
	vid := self.m[n2].ID()
	return self.g.Edge(uid, vid)
}

func (self *Graph) WeightedEdge(n1, n2 string) gonumGraph.WeightedEdge {
	uid := self.m[n1].ID()
	vid := self.m[n2].ID()
	return self.g.WeightedEdge(uid, vid)
}

func (self *Graph) Node(name string) gonumGraph.Node {
	return self.m[name]
}

func (self *Graph) Edges() gonumGraph.Edges {
	return self.g.Edges()
}

func (self *Graph) WeightedEdges() gonumGraph.WeightedEdges {
	return self.g.WeightedEdges()
}

func (self *Graph) NewWeightedEdge(n1, n2 string, weight float64) gonumGraph.WeightedEdge {
	uid := self.m[n1]
	vid := self.m[n2]
	e := self.g.NewWeightedEdge(uid, vid, weight)
	self.g.SetWeightedEdge(e)
	return e
}

func (self *Graph) AddPath(names []string, weights []float64) {
	if len(weights) != len(names)-1 {
		log.Fatalf("Weights along path should be one less than names along path.")
	}
	for i := range weights {
		self.AddNode(names[i])
		self.AddNode(names[i+1])
		self.NewWeightedEdge(names[i], names[i+1], weights[i])
	}
}

func (self *Graph) AddEqualWeightPath(names []string, weight float64) {
	weights := make([]float64, len(names)-1)
	for i := range weights {
		weights[i] = weight
	}
	self.AddPath(names, weights)
}

// AddRelationalPath adds a path from n1 to n2 distributing the
// weight accordingly between dist number of links
func (self *Graph) AddRelationalPath(n1, n2 string, dist uint, weight float64) {
	incWeight := weight / float64(dist)
	// Path length is one fewer than distance
	// parent-offspring has dist == 1, but are directly linked
	path := make([]string, dist+2-1)
	// Add knowns
	path[0] = n1
	path[len(path)-1] = n2
	// Add unknowns if there are any
	for i := range path {
		if i == 0 || i == len(path)-1 {
			continue
		} else {
			name := xid.New().String()
			path[i] = name[len(name)-lenUnknownNames:]
		}

	}
	weights := make([]float64, len(path)-1)
	for i := range weights {
		weights[i] = incWeight
	}
	self.AddPath(path, weights)
}
