package graph

import (
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/parentage"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/path"
)

const lenUnknownNames = 6

var _ gonumGraph.Graph = new(UndirectedGraph)
var _ gonumGraph.Undirected = new(UndirectedGraph)
var _ gonumGraph.Weighted = new(UndirectedGraph)

var _ Graph = new(UndirectedGraph)
var _ Graph = new(DirectedGraph)

type Graph interface {
	gonumGraph.Weighted
	NewNode() gonumGraph.Node
	EdgeBetweenNamed(n1, n2 string) gonumGraph.Edge
	EdgeNamed(n1, n2 string) gonumGraph.Edge
	HasEdgeBetweenNamed(n1, n2 string) bool
	WeightedEdgeNamed(n1, n2 string) gonumGraph.WeightedEdge
	NodeNamed(name string) gonumGraph.Node
	NewWeightedLineNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedLine
	AddNodeAge(n string, age demographics.Age)
	AddNodeSex(n string, sex demographics.Sex)
	AddNodeParentage(n, dam, sire string)
	IDToName(int64) (string, bool)
	Edges() gonumGraph.Edges
	NameToID(string) (int64, bool)
	IsKnown(name string) bool
	PruneToShortest(indvs []string) *UndirectedGraph
}

// UndirectedGraph has named nodes/vertexes
type UndirectedGraph struct {
	wug        *multi.WeightedUndirectedGraph
	nameToInfo map[string]info
	knowns     []string
}

type info struct {
	ID        int64
	sex       demographics.Sex
	age       demographics.Age
	dam, sire string
}

type DirectedGraph struct {
	wug        *multi.WeightedDirectedGraph
	nameToInfo map[string]info
	knowns     []string
}

func NewUndirectedGraph(indvs []string) *UndirectedGraph {
	return &UndirectedGraph{
		wug:        multi.NewWeightedUndirectedGraph(),
		nameToInfo: make(map[string]info, len(indvs)),
		knowns:     indvs,
	}
}

func NewDirectedGraph(indvs []string) *DirectedGraph {
	return &DirectedGraph{
		wug:        multi.NewWeightedDirectedGraph(),
		nameToInfo: make(map[string]info, len(indvs)),
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

func newDirectedGraph(in relatedness.CsvInput, maxDist relational.Degree, pars parentage.CsvInput, dems demographics.CsvInput) *DirectedGraph {
	indvs := in.Indvs()
	g := NewDirectedGraph(indvs)

	// Add known parentage
	for _, indv := range pars.Indvs() {
		degree := relational.First
		relatedness := unit.Relatedness(1.0)
		if degree <= maxDist {
			if sire, ok := pars.Sire(indv); ok {
				if path, err := NewRelationalWeightPath(sire, indv, degree, relatedness.Weight()); err == nil {
					g.AddPath(path)
				}
			}
			if dam, ok := pars.Dam(indv); ok {
				if path, err := NewRelationalWeightPath(dam, indv, degree, relatedness.Weight()); err == nil {
					g.AddPath(path)
				}
			}
		}
	}

	// Add unknowns
	for i := range indvs {
		for j := range indvs {
			if i == j {
				continue
			} else {
				from := indvs[i]
				fromAge, fromAgeOk := dems.Age(from)
				to := indvs[j]
				toAge, toAgeOk := dems.Age(to)
				degree := in.RelDistance(from, to)
				relatedness := in.Relatedness(from, to)
				if degree <= maxDist {
					if fromAgeOk && toAgeOk {
						if fromAge >= toAge {
							if path, err := NewRelationalWeightPath(from, to, degree, relatedness.Weight()); err == nil {
								g.AddPath(path)
							}
						} else {
							if path, err := NewRelationalWeightPath(to, from, degree, relatedness.Weight()); err == nil {
								g.AddPath(path)
							}
						}
					} else {
						if path, err := NewRelationalWeightPath(to, from, degree, relatedness.Weight()); err == nil {
							g.AddPath(path)
						}
					}
				}

				// Add demographics
				g.AddNodeAge(from, fromAge)
				g.AddNodeAge(to, toAge)
				if sex, ok := dems.Sex(from); ok {
					g.AddNodeSex(from, sex)
				}
				if sex, ok := dems.Sex(to); ok {
					g.AddNodeSex(to, sex)
				}
			}
		}
	}

	return g
}

func NewGraphFromCsvInput(in relatedness.CsvInput, maxDist relational.Degree, pars parentage.CsvInput, dems demographics.CsvInput) Graph {
	if pars == nil && dems == nil {
		return newUndirectedGraph(in, maxDist)
	}
	return newDirectedGraph(in, maxDist, pars, dems)
}

func (graph *UndirectedGraph) PruneToShortest(indvs []string) *UndirectedGraph {
	return PruneToShortest(graph, indvs)
}

func (graph *DirectedGraph) PruneToShortest(indvs []string) *UndirectedGraph {
	return PruneToShortest(graph, indvs)
}

func PruneToShortest(graph Graph, indvs []string) *UndirectedGraph {
	g := NewUndirectedGraph(indvs)
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

func (graph *UndirectedGraph) IsKnown(name string) bool {
	for i := range graph.knowns {
		if name == graph.knowns[i] {
			return true
		}
	}
	return false
}

func (graph *DirectedGraph) IsKnown(name string) bool {
	for i := range graph.knowns {
		if name == graph.knowns[i] {
			return true
		}
	}
	return false
}

func (self *UndirectedGraph) AddPath(p Path) {
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

func (self *DirectedGraph) AddPath(p Path) {
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
func (graph *UndirectedGraph) IDToName(id int64) (string, bool) {
	for name, info := range graph.nameToInfo {
		if info.ID == id {
			return name, true
		}
	}
	return "", false
}

// IDToName converts the id to its corresponding node name
// Returns false if the node does not exist
func (graph *DirectedGraph) IDToName(id int64) (string, bool) {
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

// NameToID converts the name to its corresponding node ID
// Returns false if the node does not exist
func (graph *DirectedGraph) NameToID(name string) (int64, bool) {
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

func (graph *DirectedGraph) Weight(xid, yid int64) (w float64, ok bool) {
	return graph.wug.Weight(xid, yid)
}

func (graph *UndirectedGraph) AddNodeParentage(n, dam, sire string) {
	info := graph.nameToInfo[n]
	info.dam = dam
	info.sire = sire
	graph.nameToInfo[n] = info
}

func (graph *DirectedGraph) AddNodeParentage(n, dam, sire string) {
	info := graph.nameToInfo[n]
	info.dam = dam
	info.sire = sire
	graph.nameToInfo[n] = info
}

func (graph *UndirectedGraph) AddNodeAge(n string, age demographics.Age) {
	info := graph.nameToInfo[n]
	info.age = age
	graph.nameToInfo[n] = info
}
func (graph *UndirectedGraph) AddNodeSex(n string, sex demographics.Sex) {
	info := graph.nameToInfo[n]
	info.sex = sex
	graph.nameToInfo[n] = info
}

func (graph *DirectedGraph) AddNodeAge(n string, age demographics.Age) {
	info := graph.nameToInfo[n]
	info.age = age
	graph.nameToInfo[n] = info
}
func (graph *DirectedGraph) AddNodeSex(n string, sex demographics.Sex) {
	info := graph.nameToInfo[n]
	info.sex = sex
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

func (graph *DirectedGraph) From(id int64) gonumGraph.Nodes {
	return graph.wug.From(id)
}

func (graph *UndirectedGraph) FromNamed(name string) gonumGraph.Nodes {
	if info, ok := graph.nameToInfo[name]; ok {
		return graph.From(info.ID)
	}
	return gonumGraph.Empty
}

func (graph *DirectedGraph) FromNamed(name string) gonumGraph.Nodes {
	if info, ok := graph.nameToInfo[name]; ok {
		return graph.From(info.ID)
	}
	return gonumGraph.Empty
}

func (graph *UndirectedGraph) RemoveNode(id int64) {
	graph.wug.RemoveNode(id)
}

func (graph *DirectedGraph) RemoveNode(id int64) {
	graph.wug.RemoveNode(id)
}

func (graph *UndirectedGraph) RemoveNodeNamed(name string) {
	if info, ok := graph.nameToInfo[name]; ok {
		graph.RemoveNode(info.ID)
		delete(graph.nameToInfo, name)
	}
}

func (graph *DirectedGraph) RemoveNodeNamed(name string) {
	if info, ok := graph.nameToInfo[name]; ok {
		graph.RemoveNode(info.ID)
		delete(graph.nameToInfo, name)
	}
}

func (graph *UndirectedGraph) AddNode(n gonumGraph.Node) {
	graph.wug.AddNode(n)
}

func (graph *DirectedGraph) AddNode(n gonumGraph.Node) {
	graph.wug.AddNode(n)
}

func (graph *UndirectedGraph) Nodes() gonumGraph.Nodes {
	return graph.wug.Nodes()
}

func (graph *DirectedGraph) Nodes() gonumGraph.Nodes {
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

func (graph *DirectedGraph) AddNodeNamed(name string) {
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

func (graph *DirectedGraph) NewNode() gonumGraph.Node {
	return graph.wug.NewNode()
}

func (graph *UndirectedGraph) Edge(uid, vid int64) gonumGraph.Edge {
	return graph.wug.Edge(uid, vid)
}

func (graph *DirectedGraph) Edge(uid, vid int64) gonumGraph.Edge {
	return graph.wug.Edge(uid, vid)
}

func (graph *UndirectedGraph) EdgeNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.Edge(uID, vID)
}

func (graph *DirectedGraph) EdgeNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.Edge(uID, vID)
}

func (graph *UndirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	return graph.wug.HasEdgeBetween(xid, yid)
}

func (graph *DirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	return graph.wug.HasEdgeBetween(xid, yid)
}

func (graph *UndirectedGraph) EdgeBetween(xid, yid int64) gonumGraph.Edge {
	return graph.wug.EdgeBetween(xid, yid)
}

func (graph *DirectedGraph) EdgeBetween(xid, yid int64) gonumGraph.Edge {
	return graph.wug.Edge(xid, yid)
}

func (graph *UndirectedGraph) EdgeBetweenNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.EdgeBetween(uID, vID)
}

func (graph *DirectedGraph) EdgeBetweenNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.EdgeBetween(uID, vID)
}

func (graph *UndirectedGraph) HasEdgeBetweenNamed(n1, n2 string) bool {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.HasEdgeBetween(uID, vID)
}

func (graph *DirectedGraph) HasEdgeBetweenNamed(n1, n2 string) bool {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.HasEdgeBetween(uID, vID)
}

func (graph *UndirectedGraph) WeightedEdge(uid, vid int64) gonumGraph.WeightedEdge {
	return graph.wug.WeightedEdge(uid, vid)
}

func (graph *DirectedGraph) WeightedEdge(uid, vid int64) gonumGraph.WeightedEdge {
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

func (graph *DirectedGraph) WeightedEdgeNamed(n1, n2 string) gonumGraph.WeightedEdge {
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

func (graph *DirectedGraph) Node(id int64) gonumGraph.Node {
	return graph.wug.Node(id)
}

func (graph *UndirectedGraph) NodeNamed(name string) gonumGraph.Node {
	if id, ok := graph.NameToID(name); ok {
		return graph.wug.Node(id)
	}
	return gonumGraph.Empty.Node()
}

func (graph *DirectedGraph) NodeNamed(name string) gonumGraph.Node {
	if id, ok := graph.NameToID(name); ok {
		return graph.wug.Node(id)
	}
	return gonumGraph.Empty.Node()
}

func (graph *UndirectedGraph) Edges() gonumGraph.Edges {
	return graph.wug.Edges()
}

func (graph *DirectedGraph) Edges() gonumGraph.Edges {
	return graph.wug.Edges()
}

func (graph *UndirectedGraph) WeightedEdges() gonumGraph.WeightedEdges {
	return graph.wug.WeightedEdges()
}

func (graph *UndirectedGraph) NewWeightedLine(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedLine {
	return graph.wug.NewWeightedLine(from, to, weight)
}

func (graph *DirectedGraph) NewWeightedLine(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedLine {
	return graph.wug.NewWeightedLine(from, to, weight)
}

func (graph *UndirectedGraph) NewWeightedLineNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedLine {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.NewWeightedLine(graph.Node(uID), graph.Node(vID), float64(weight))
}

func (graph *DirectedGraph) NewWeightedLineNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedLine {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.NewWeightedLine(graph.Node(uID), graph.Node(vID), float64(weight))
}

func (graph *UndirectedGraph) SetWeightedLine(e gonumGraph.WeightedLine) {
	graph.wug.SetWeightedLine(e)
}

func (graph *DirectedGraph) SetWeightedLine(e gonumGraph.WeightedLine) {
	graph.wug.SetWeightedLine(e)
}
