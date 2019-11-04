package graph

import (
	"fmt"

	"github.com/rhagenson/relped/internal/csvin"
	"github.com/rhagenson/relped/internal/unit"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

const lenUnknownNames = 6

var _ gonumGraph.Graph = new(Graph)

// Graph has named nodes/vertexes
type Graph struct {
	wug      *simple.WeightedUndirectedGraph
	nameToID map[string]int64
}

func NewGraph() *Graph {
	return &Graph{
		simple.NewWeightedUndirectedGraph(0, 0),
		make(map[string]int64),
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
			g.AddNodeNamed(i1)
			g.AddNodeNamed(i2)
			graphdist := in.RelDistance(i1, i2).GraphDistance()
			relatedness := in.Relatedness(i1, i2)
			if graphdist <= maxDist.GraphDistance() {
				path := NewRelationalWeightPath(i1, i2, graphdist, relatedness.Weight())
				g.AddPath(path)
			}
		}
	}
	return g
}

func (self *Graph) PruneToShortest(indvs []string) *Graph {
	g := NewGraph()
	for i := 0; i < len(indvs); i++ {
		if src := self.NodeNamed(indvs[i]); src != nil {
			if shortest, ok := path.BellmanFordFrom(src, self.g); ok {
				for j := i + 1; j < len(indvs); j++ {
					if dest := self.NodeNamed(indvs[j]); dest != nil {
						nodes, cost := shortest.To(dest.ID())
						if len(nodes) != 0 {
							names := make([]string, len(nodes))
							for i, node := range nodes {
								if name, ok := self.IDToName(node.ID()); ok {
									names[i] = name
								} else {
									fmt.Printf("Node %q was not found\n", name)
								}
							}
							path := NewFractionalWeightPath(names, unit.Weight(cost))
							g.AddPath(path)
						}
					}
				}
			}
		}
	}
	return g
}

// IDToName converts the id to its corresponding node name
// Returns false if the node does not exist
func (self *Graph) IDToName(id int64) (string, bool) {
	for name, nid := range self.nameToID {
		if nid == id {
			return name, true
		}
	}
	return "", false
}

// NameToID converts the name to its corresponding node ID
// Returns false if the node does not exist
func (self *Graph) NameToID(name string) (int64, bool) {
	id, ok := self.nameToID[name]
	return id, ok
}

func (self *Graph) RmDisconnected() {
	for name := range self.nameToID {
		nodes := self.FromNamed(name)
		if nodes.Len() == 0 {
			self.RemoveNodeNamed(name)
		}
	}
}

func (self *Graph) Weight(xid, yid int64) (w float64, ok bool) {
	return self.wug.Weight(xid, yid)
}

func (self *Graph) WeightNamed(n1, n2 string) (w float64, ok bool) {
	xid := self.nameToID[n1]
	yid := self.nameToID[n2]
	return self.Weight(xid, yid)
}

func (self *Graph) From(id int64) gonumGraph.Nodes {
	return self.wug.From(id)
}

func (self *Graph) FromNamed(name string) gonumGraph.Nodes {
	if id, ok := self.nameToID[name]; ok {
		return self.From(id)
	}
	return gonumGraph.Empty
}

func (self *Graph) RemoveNode(id int64) {
	self.wug.RemoveNode(id)
}

func (self *Graph) RemoveNodeNamed(name string) {
	if id, ok := self.nameToID[name]; ok {
		self.RemoveNode(id)
		delete(self.nameToID, name)
	}
}

func (self *Graph) AddNode(n gonumGraph.Node) {
	self.wug.AddNode(n)
}

func (self *Graph) Nodes() gonumGraph.Nodes {
	return self.wug.Nodes()
}

func (self *Graph) AddNodeNamed(name string) {
	if _, ok := self.nameToID[name]; !ok {
		n := self.NewNode()
		self.AddNode(n)
		self.nameToID[name] = n.ID()
	}
}

func (self *Graph) NewNode() gonumGraph.Node {
	return self.wug.NewNode()
}

func (self *Graph) Edge(uid, vid int64) gonumGraph.Edge {
	return self.wug.Edge(uid, vid)
}

func (self *Graph) EdgeNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := self.NameToID(n1)
	vID, _ := self.NameToID(n2)
	return self.wug.Edge(uID, vID)
}

func (self *Graph) HasEdgeBetween(xid, yid int64) bool {
	return self.wug.HasEdgeBetween(xid, yid)
}

func (self *Graph) HasEdgeBetweenNamed(n1, n2 string) bool {
	uID, _ := self.NameToID(n1)
	vID, _ := self.NameToID(n2)
	return self.HasEdgeBetween(uID, vID)
}

func (self *Graph) WeightedEdge(uid, vid int64) gonumGraph.WeightedEdge {
	return self.wug.WeightedEdge(uid, vid)
}

func (self *Graph) WeightedEdgeNamed(n1, n2 string) gonumGraph.WeightedEdge {
	uID, uOK := self.NameToID(n1)
	vID, vOK := self.NameToID(n2)
	if uOK && vOK {
		return self.wug.WeightedEdge(uID, vID)
	}
	return nil
}

func (self *Graph) Node(id int64) gonumGraph.Node {
	return self.wug.Node(id)
}

func (self *Graph) NodeNamed(name string) gonumGraph.Node {
	if id, ok := self.NameToID(name); ok {
		return self.wug.Node(id)
	}
	return gonumGraph.Empty.Node()
}

func (self *Graph) Edges() gonumGraph.Edges {
	return self.wug.Edges()
}

func (self *Graph) WeightedEdges() gonumGraph.WeightedEdges {
	return self.wug.WeightedEdges()
}

func (self *Graph) NewWeightedEdge(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedEdge {
	return self.wug.NewWeightedEdge(from, to, weight)
}

func (self *Graph) NewWeightedEdgeNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedEdge {
	uID, _ := self.NameToID(n1)
	vID, _ := self.NameToID(n2)
	return self.NewWeightedEdge(self.Node(uID), self.Node(vID), float64(weight))
}

func (self *Graph) SetWeightedEdge(e gonumGraph.WeightedEdge) {
	self.wug.SetWeightedEdge(e)
}
