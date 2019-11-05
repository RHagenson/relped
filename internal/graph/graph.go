package graph

import (
	"fmt"

	"github.com/rhagenson/relped/internal/csvin"
	"github.com/rhagenson/relped/internal/unit"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/path"
)

const lenUnknownNames = 6

var _ gonumGraph.Graph = new(Graph)
var _ gonumGraph.Undirected = new(Graph)
var _ gonumGraph.Weighted = new(Graph)

// Graph has named nodes/vertexes
type Graph struct {
	wug      *multi.WeightedUndirectedGraph
	nameToID map[string]int64
}

func NewGraph() *Graph {
	return &Graph{
		wug:      multi.NewWeightedUndirectedGraph(),
		nameToID: make(map[string]int64),
	}
}

func NewGraphFromCsvInput(in csvin.CsvInput, maxDist unit.RelationalDistance) *Graph {
	indvs := in.Indvs()
	g := NewGraph()

	for i := range indvs {
		for j := range indvs {
			if i == j {
				continue
			} else {
				from := indvs[i]
				to := indvs[j]
				graphdist := in.RelDistance(from, to).GraphDistance()
				fmt.Println(graphdist)
				relatedness := in.Relatedness(from, to)
				if graphdist <= maxDist.GraphDistance() {
					path := NewRelationalWeightPath(from, to, graphdist, relatedness.Weight())
					g.AddPath(path)
				}
			}
		}
	}
	// Add paths from node to node based on relational distance
	for i := 0; i < len(indvs); i++ {
		for j := i + 1; j < len(indvs); j++ {
			i1 := indvs[i]
			i2 := indvs[j]
			graphdist := in.RelDistance(i1, i2).GraphDistance()
			relatedness := in.Relatedness(i1, i2)
			switch graphdist {
			case 0:
				g.AddNodeNamed(i1)
				g.AddNodeNamed(i2)
				g.SetWeightedLine(g.NewWeightedLineNamed(i1, i2, relatedness.Weight()))
			}
			// if graphdist <= maxDist.GraphDistance() {
			// 	path := NewRelationalWeightPath(i1, i2, graphdist, relatedness.Weight())
			// 	g.AddPath(path)
			// }
		}
	}
	return g
}

func (graph *Graph) PruneToShortest(indvs []string) *Graph {
	g := NewGraph()
	for i := 0; i < len(indvs); i++ {
		if src := graph.NodeNamed(indvs[i]); src != nil {
			if shortest, ok := path.BellmanFordFrom(src, graph); ok {
				for j := i + 1; j < len(indvs); j++ {
					if dest := graph.NodeNamed(indvs[j]); dest != nil {
						nodes, cost := shortest.To(dest.ID())
						if len(nodes) != 0 {
							names := make([]string, len(nodes))
							for i, node := range nodes {
								if name, ok := graph.IDToName(node.ID()); ok {
									names[i] = name
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

func (self *Graph) AddPath(p Path) {
	names := p.Names()
	weights := p.Weights()

	for i := range weights {
		from := names[i]
		to := names[i+1]
		weight := weights[i]
		self.AddNodeNamed(from)
		self.AddNodeNamed(to)
		self.SetWeightedLine(self.NewWeightedLineNamed(from, to, weight))
	}
}

// IDToName converts the id to its corresponding node name
// Returns false if the node does not exist
func (graph *Graph) IDToName(id int64) (string, bool) {
	for name, nid := range graph.nameToID {
		if nid == id {
			return name, true
		}
	}
	return "", false
}

// NameToID converts the name to its corresponding node ID
// Returns false if the node does not exist
func (graph *Graph) NameToID(name string) (int64, bool) {
	id, ok := graph.nameToID[name]
	return id, ok
}

func (graph *Graph) RmDisconnected() {
	for name := range graph.nameToID {
		nodes := graph.FromNamed(name)
		if nodes.Len() == 0 {
			graph.RemoveNodeNamed(name)
		}
	}
}

func (graph *Graph) Weight(xid, yid int64) (w float64, ok bool) {
	return graph.wug.Weight(xid, yid)
}

func (graph *Graph) WeightNamed(n1, n2 string) (w float64, ok bool) {
	xid := graph.nameToID[n1]
	yid := graph.nameToID[n2]
	return graph.Weight(xid, yid)
}

func (graph *Graph) From(id int64) gonumGraph.Nodes {
	return graph.wug.From(id)
}

func (graph *Graph) FromNamed(name string) gonumGraph.Nodes {
	if id, ok := graph.nameToID[name]; ok {
		return graph.From(id)
	}
	return gonumGraph.Empty
}

func (graph *Graph) RemoveNode(id int64) {
	graph.wug.RemoveNode(id)
}

func (graph *Graph) RemoveNodeNamed(name string) {
	if id, ok := graph.nameToID[name]; ok {
		graph.RemoveNode(id)
		delete(graph.nameToID, name)
	}
}

func (graph *Graph) AddNode(n gonumGraph.Node) {
	graph.wug.AddNode(n)
}

func (graph *Graph) Nodes() gonumGraph.Nodes {
	return graph.wug.Nodes()
}

func (graph *Graph) AddNodeNamed(name string) {
	if _, ok := graph.nameToID[name]; !ok {
		n := graph.NewNode()
		graph.AddNode(n)
		graph.nameToID[name] = n.ID()
	}
}

func (graph *Graph) NewNode() gonumGraph.Node {
	return graph.wug.NewNode()
}

func (graph *Graph) Edge(uid, vid int64) gonumGraph.Edge {
	return graph.wug.Edge(uid, vid)
}

func (graph *Graph) EdgeNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.Edge(uID, vID)
}

func (graph *Graph) HasEdgeBetween(xid, yid int64) bool {
	return graph.wug.HasEdgeBetween(xid, yid)
}

func (graph *Graph) EdgeBetween(xid, yid int64) gonumGraph.Edge {
	return graph.wug.EdgeBetween(xid, yid)
}

func (graph *Graph) EdgeBetweenNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.EdgeBetween(uID, vID)
}

func (graph *Graph) HasEdgeBetweenNamed(n1, n2 string) bool {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.HasEdgeBetween(uID, vID)
}

func (graph *Graph) WeightedEdge(uid, vid int64) gonumGraph.WeightedEdge {
	return graph.wug.WeightedEdge(uid, vid)
}

func (graph *Graph) WeightedEdgeNamed(n1, n2 string) gonumGraph.WeightedEdge {
	uID, uOK := graph.NameToID(n1)
	vID, vOK := graph.NameToID(n2)
	if uOK && vOK {
		return graph.wug.WeightedEdge(uID, vID)
	}
	return nil
}

func (graph *Graph) Node(id int64) gonumGraph.Node {
	return graph.wug.Node(id)
}

func (graph *Graph) NodeNamed(name string) gonumGraph.Node {
	if id, ok := graph.NameToID(name); ok {
		return graph.wug.Node(id)
	}
	return gonumGraph.Empty.Node()
}

func (graph *Graph) Edges() gonumGraph.Edges {
	return graph.wug.Edges()
}

func (graph *Graph) WeightedEdges() gonumGraph.WeightedEdges {
	return graph.wug.WeightedEdges()
}

func (graph *Graph) NewWeightedLine(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedLine {
	return graph.wug.NewWeightedLine(from, to, weight)
}

func (graph *Graph) NewWeightedLineNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedLine {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.NewWeightedLine(graph.Node(uID), graph.Node(vID), float64(weight))
}

func (graph *Graph) SetWeightedLine(e gonumGraph.WeightedLine) {
	graph.wug.SetWeightedLine(e)
}

func (graph *Graph) String() string {
	nodes := graph.wug.Nodes()
	edges := graph.wug.Edges()
	return fmt.Sprintf("Graph:\n\tNodes(%d):\n\t%v\n\tEdges(%d):\n\t%t\nMap:\n%v\nmap[527]==%n\n", nodes.Len(), nodes, edges.Len(), edges, graph.nameToID, graph.nameToID["527"])
}
