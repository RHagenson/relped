package graph

import (
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/path"
)

var _ gonumGraph.Graph = new(UndirectedGraph)
var _ gonumGraph.Undirected = new(UndirectedGraph)
var _ gonumGraph.Weighted = new(UndirectedGraph)

var _ Graph = new(UndirectedGraph)

// UndirectedGraph has named nodes/vertexes
type UndirectedGraph struct {
	wug        *multi.WeightedUndirectedGraph
	nameToInfo map[string]Info
	knowns     []string
}

func NewUndirectedGraph(indvs []string) *UndirectedGraph {
	return &UndirectedGraph{
		wug:        multi.NewWeightedUndirectedGraph(),
		nameToInfo: make(map[string]Info, len(indvs)),
		knowns:     indvs,
	}
}

func newUndirectedGraph(in relatedness.CsvInput, maxDist relational.Degree) *UndirectedGraph {
	indvs := in.Indvs()
	g := NewUndirectedGraph(indvs)

	for i := range indvs {
		for j := range indvs {
			if i == j {
				continue
			} else {
				from := indvs[i]
				to := indvs[j]
				degree := in.RelDistance(from, to)
				relatedness := in.Relatedness(from, to)
				if degree <= maxDist {
					if path, err := NewRelationalWeightPath(from, to, degree, relatedness.Weight()); err == nil {
						g.AddPath(path)
					}
				}
			}
		}
	}
	return g
}

func (graph *UndirectedGraph) Info(n string) Info {
	return graph.nameToInfo[n]
}

func (graph *UndirectedGraph) AddInfo(name string, info Info) {
	graph.nameToInfo[name] = info
}

func (graph *UndirectedGraph) PruneToShortest(indvs []string) Graph {
	g := NewUndirectedGraph(indvs)
	for i := 0; i < len(indvs); i++ {
		if src := graph.NodeNamed(indvs[i]); src != nil {
			if shortest, ok := path.BellmanFordFrom(src, graph); ok {
				for j := i + 1; j < len(indvs); j++ {
					if dest := graph.NodeNamed(indvs[j]); dest != nil {
						nodes, cost := shortest.To(dest.ID())
						names := make([]string, len(nodes))
						if len(nodes) != 0 {
							for i := range nodes {
								from := nodes[i]
								if g.Node(from.ID()) == nil {
									if name, ok := graph.IDToName(from.ID()); ok {
										names[i] = name
										g.AddNodeNamed(name)
										info := g.Info(name)
										g.AddInfo(name, info)
									}
								}
							}
							g.AddPath(NewFractionalWeightPath(names, unit.Weight(cost)))
						}
					}
				}
			}
		}
	}
	return g
}

func (graph *UndirectedGraph) IsKnown(name string) bool {
	for i := range graph.knowns {
		if name == graph.knowns[i] {
			return true
		}
	}
	return false
}

func (graph *UndirectedGraph) AddPath(p Path) {
	names := p.Names()
	weights := p.Weights()

	from := graph.NodeNamed(names[0])
	to := graph.NodeNamed(names[len(names)-1])

	// Maintain only the shortest path
	if from != nil && to != nil {
		shortest, _ := path.AStar(from, to, graph, nil)
		nodes, _ := shortest.To(to.ID())
		if len(nodes) != 0 && len(nodes) < len(names) {
			return
		}
	}

	for i := range weights {
		from := names[i]
		to := names[i+1]
		weight := weights[i]
		graph.AddNodeNamed(from)
		graph.AddNodeNamed(to)
		graph.SetWeightedLine(graph.NewWeightedLineNamed(from, to, weight))
	}
}

// IDToName converts the id to its corresponding node name
// Returns false if the node does not exist
func (graph *UndirectedGraph) IDToName(id int64) (string, bool) {
	for name, info := range graph.nameToInfo {
		if info.ID == id {
			return name, true
		}
	}
	return "", false
}

// NameToID converts the name to its corresponding node ID
// Returns false if the node does not exist
func (graph *UndirectedGraph) NameToID(name string) (int64, bool) {
	info, ok := graph.nameToInfo[name]
	return info.ID, ok
}

func (graph *UndirectedGraph) RmDisconnected() {
	for name := range graph.nameToInfo {
		nodes := graph.FromNamed(name)
		if nodes.Len() == 0 {
			graph.RemoveNodeNamed(name)
		}
	}
}

func (graph *UndirectedGraph) Weight(xid, yid int64) (w float64, ok bool) {
	return graph.wug.Weight(xid, yid)
}

func (graph *UndirectedGraph) AddNodeParentage(n, dam, sire string) {
	info := graph.nameToInfo[n]
	info.Dam = dam
	info.Sire = sire
	graph.nameToInfo[n] = info
}

func (graph *UndirectedGraph) AddNodeAge(n string, age demographics.Age) {
	info := graph.nameToInfo[n]
	info.Age = age
	graph.nameToInfo[n] = info
}
func (graph *UndirectedGraph) AddNodeSex(n string, sex demographics.Sex) {
	info := graph.nameToInfo[n]
	info.Sex = sex
	graph.nameToInfo[n] = info
}

func (graph *UndirectedGraph) WeightNamed(n1, n2 string) (w float64, ok bool) {
	xinfo := graph.nameToInfo[n1]
	yinfo := graph.nameToInfo[n2]
	return graph.Weight(xinfo.ID, yinfo.ID)
}

func (graph *UndirectedGraph) From(id int64) gonumGraph.Nodes {
	return graph.wug.From(id)
}

func (graph *UndirectedGraph) FromNamed(name string) gonumGraph.Nodes {
	if info, ok := graph.nameToInfo[name]; ok {
		return graph.From(info.ID)
	}
	return gonumGraph.Empty
}

func (graph *UndirectedGraph) RemoveNode(id int64) {
	graph.wug.RemoveNode(id)
}

func (graph *UndirectedGraph) RemoveNodeNamed(name string) {
	if info, ok := graph.nameToInfo[name]; ok {
		graph.RemoveNode(info.ID)
		delete(graph.nameToInfo, name)
	}
}

func (graph *UndirectedGraph) AddNode(n gonumGraph.Node) {
	graph.wug.AddNode(n)
}

func (graph *UndirectedGraph) Nodes() gonumGraph.Nodes {
	return graph.wug.Nodes()
}

func (graph *UndirectedGraph) AddNodeNamed(name string) {
	if _, ok := graph.nameToInfo[name]; !ok {
		n := graph.NewNode()
		graph.AddNode(n)
		info := graph.nameToInfo[name]
		info.ID = n.ID()
		graph.nameToInfo[name] = info
	}
}

func (graph *UndirectedGraph) NewNode() gonumGraph.Node {
	return graph.wug.NewNode()
}

func (graph *UndirectedGraph) Edge(uid, vid int64) gonumGraph.Edge {
	return graph.wug.Edge(uid, vid)
}

func (graph *UndirectedGraph) EdgeNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.Edge(uID, vID)
}

func (graph *UndirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	return graph.wug.HasEdgeBetween(xid, yid)
}

func (graph *UndirectedGraph) EdgeBetween(xid, yid int64) gonumGraph.Edge {
	return graph.wug.EdgeBetween(xid, yid)
}

func (graph *UndirectedGraph) EdgeBetweenNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.EdgeBetween(uID, vID)
}

func (graph *UndirectedGraph) HasEdgeBetweenNamed(n1, n2 string) bool {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.HasEdgeBetween(uID, vID)
}

func (graph *UndirectedGraph) WeightedEdge(uid, vid int64) gonumGraph.WeightedEdge {
	return graph.wug.WeightedEdge(uid, vid)
}

func (graph *UndirectedGraph) WeightedEdgeNamed(n1, n2 string) gonumGraph.WeightedEdge {
	uID, uOK := graph.NameToID(n1)
	vID, vOK := graph.NameToID(n2)
	if uOK && vOK {
		return graph.wug.WeightedEdge(uID, vID)
	}
	return nil
}

func (graph *UndirectedGraph) Node(id int64) gonumGraph.Node {
	return graph.wug.Node(id)
}

func (graph *UndirectedGraph) NodeNamed(name string) gonumGraph.Node {
	if id, ok := graph.NameToID(name); ok {
		return graph.wug.Node(id)
	}
	return gonumGraph.Empty.Node()
}

func (graph *UndirectedGraph) Edges() gonumGraph.Edges {
	return graph.wug.Edges()
}

func (graph *UndirectedGraph) WeightedEdges() gonumGraph.WeightedEdges {
	return graph.wug.WeightedEdges()
}

func (graph *UndirectedGraph) NewWeightedLine(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedLine {
	return graph.wug.NewWeightedLine(from, to, weight)
}

func (graph *UndirectedGraph) NewWeightedLineNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedLine {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.NewWeightedLine(graph.Node(uID), graph.Node(vID), float64(weight))
}

func (graph *UndirectedGraph) SetWeightedLine(e gonumGraph.WeightedLine) {
	graph.wug.SetWeightedLine(e)
}
