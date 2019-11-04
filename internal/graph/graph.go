package graph

import (
	"log"

	"github.com/rhagenson/relped/internal/csvin"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rs/xid"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

const lenUnknownNames = 6

// Graph has named nodes/vertexes
type Graph struct {
	g     *simple.WeightedUndirectedGraph
	nodes map[string]gonumGraph.Node
}

func NewGraph() *Graph {
	return &Graph{
		g:     simple.NewWeightedUndirectedGraph(0, 0),
		nodes: make(map[string]gonumGraph.Node),
	}
}

func NewGraphFromCsvInput(in csvin.CsvInput, maxDist unit.RelationalDistance) *Graph {
	indvs := in.Indvs()
	g := NewGraph()

	// Add paths from node to node based on relational distance
	for i := 0; i < len(indvs); i++ {
		for j := i + 1; j < len(indvs); j++ {
			i1 := indvs[i]
			i2 := indvs[j]
			g.AddNode(i1)
			g.AddNode(i2)
			graphdist := in.RelDistance(i1, i2).GraphDistance()
			relatedness := in.Relatedness(i1, i2)
			if graphdist <= maxDist.GraphDistance() {
				g.AddRelationalPath(i1, i2, graphdist, relatedness.Weight())
			}
		}
	}
	return g
}

func (self *Graph) PruneToShortest(indvs []string) *Graph {
	g := NewGraph()
	for i := 0; i < len(indvs); i++ {
		if src := self.Node(indvs[i]); src != nil {
			if shortest, ok := path.BellmanFordFrom(src, self.g); ok {
				for j := i + 1; j < len(indvs); j++ {
					if dest := self.Node(indvs[j]); dest != nil {
						nodes, cost := shortest.To(dest.ID())
						if len(nodes) != 0 {
							names := make([]string, len(nodes))
							for i, node := range nodes {
								names[i] = self.NameFromID(node.ID())
							}
							g.AddFractionalWeightPath(names, unit.Weight(cost))
						}
					}
				}
			}
		}
	}
	return g
}

func (self *Graph) Nodes() gonumGraph.Nodes {
	return self.g.Nodes()
}

func (self *Graph) NameFromID(id int64) string {
	for name, node := range self.nodes {
		if node.ID() == id {
			return name
		}
	}
	return ""
}

func (self *Graph) RmDisconnected() {
	for name := range self.nodes {
		nodes := self.From(name)
		if nodes.Len() == 0 {
			self.RemoveNode(name)
		}
	}
}

func (self *Graph) Weight(xid, yid int64) (w unit.Weight, ok bool) {
	weight, ok := self.g.Weight(xid, yid)
	return unit.Weight(weight), ok
}

func (self *Graph) From(name string) gonumGraph.Nodes {
	if node, ok := self.nodes[name]; ok {
		return self.g.From(node.ID())
	}
	return nil
}

func (self *Graph) RemoveNode(name string) {
	if node, ok := self.nodes[name]; ok {
		self.g.RemoveNode(node.ID())
	}
}

func (self *Graph) AddNode(name string) {
	if _, ok := self.nodes[name]; !ok {
		n := self.g.NewNode()
		self.g.AddNode(n)
		self.nodes[name] = n
	}
}

func (self *Graph) Edge(n1, n2 string) gonumGraph.Edge {
	uid := self.nodes[n1].ID()
	vid := self.nodes[n2].ID()
	return self.g.Edge(uid, vid)
}

func (self *Graph) WeightedEdge(n1, n2 string) gonumGraph.WeightedEdge {
	uid := self.nodes[n1].ID()
	vid := self.nodes[n2].ID()
	return self.g.WeightedEdge(uid, vid)
}

func (self *Graph) Node(name string) gonumGraph.Node {
	return self.nodes[name]
}

func (self *Graph) Edges() gonumGraph.Edges {
	return self.g.Edges()
}

func (self *Graph) WeightedEdges() gonumGraph.WeightedEdges {
	return self.g.WeightedEdges()
}

func (self *Graph) NewWeightedEdge(n1, n2 string, weight unit.Weight) gonumGraph.WeightedEdge {
	uid := self.nodes[n1]
	vid := self.nodes[n2]
	e := self.g.NewWeightedEdge(uid, vid, float64(weight))
	self.g.SetWeightedEdge(e)
	return e
}

func (self *Graph) AddPath(names []string, weights []unit.Weight) {
	if len(weights) != len(names)-1 {
		log.Fatalf("Weights along path should be one less than names along path.")
	}
	for i := range weights {
		self.AddNode(names[i])
		self.AddNode(names[i+1])
		self.NewWeightedEdge(names[i], names[i+1], weights[i])
	}
}

func (self *Graph) AddEqualWeightPath(names []string, weight unit.Weight) {
	weights := make([]unit.Weight, len(names)-1)
	for i := range weights {
		weights[i] = weight
	}
	self.AddPath(names, weights)
}

func (self *Graph) AddFractionalWeightPath(names []string, weight unit.Weight) {
	weights := make([]unit.Weight, len(names)-1)
	incWeight := float64(weight) / float64(len(weights))
	for i := range weights {
		weights[i] = unit.Weight(incWeight)
	}
	self.AddPath(names, weights)
}

// AddRelationalPath adds a path from n1 to n2 with dist unknowns separating them
func (self *Graph) AddRelationalPath(n1, n2 string, dist unit.GraphDistance, weight unit.Weight) {
	path := make([]string, dist+2)
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
	weights := make([]unit.Weight, len(path)-1)
	incWeight := float64(weight) / float64(len(weights))
	for i := range weights {
		weights[i] = unit.Weight(incWeight)
	}
	self.AddPath(path, weights)
}
