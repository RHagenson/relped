package graph

import (
	"log"
	"strings"

	"github.com/rhagenson/relped/internal/csvin"
	"github.com/rs/xid"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

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
	for i := range indvs {
		for j := i + 1; j < len(indvs); j++ {
			i1 := indvs[i]
			i2 := indvs[j]
			if i1 != i2 {
				dist := in.RelDistance(i1, i2)
				if dist <= maxDist {
					weight := in.Relatedness(i1, i2)
					g.AddUnknownPath(i1, i2, dist, weight)
				}
			}
		}
	}
	return g
}

func (self *Graph) PruneToShortest() *Graph {
	g := NewGraph()

	for name1, node1 := range self.m {
		if strings.Contains(name1, "Unknown") {
			continue
		}
		for name2, node2 := range self.m {
			if strings.Contains(name2, "Unknown") {
				continue
			}
			if name1 == name2 {
				continue
			}
			paths := path.YenKShortestPaths(self.g, 10, node1, node2)
			for i := range paths {
				names := make([]string, len(paths[i]))
				weights := make([]float64, len(names)-1)
				for j := range paths[i] {
					names[j] = self.NameFromID(paths[i][j].ID())
				}
				for i := 1; i < len(names); i++ {
					weights[i-1] = self.WeightedEdge(names[i-1], names[i]).Weight()
				}
				g.AddPath(names, weights)
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
	return self.g.Node(self.m[name].ID())
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
	for i := 1; i < len(names); i++ {
		self.AddNode(names[i-1])
		self.AddNode(names[i])
		self.NewWeightedEdge(names[i-1], names[i], weights[i-1])
	}
}

func (self *Graph) AddEqualWeightPath(names []string, weight float64) {
	weights := make([]float64, len(names)-1)
	for i := range weights {
		weights[i] = weight
	}
	self.AddPath(names, weights)
}

// AddUnknownPath adds a path from n1 through n "unknowns" to n2 distributing the
// weight accordingly
func (self *Graph) AddUnknownPath(n1, n2 string, n uint, weight float64) {
	incWeight := weight / float64(n)
	path := make([]string, n+2)
	// Add knowns
	path[0] = n1
	path[len(path)-1] = n2
	// Add unknowns
	for i := 1; i < len(path)-1; i++ {
		path[i] = "Unknown" + xid.New().String()
	}
	weights := make([]float64, len(path)-1)
	for i := range weights {
		weights[i] = incWeight
	}
	self.AddPath(path, weights)
}
