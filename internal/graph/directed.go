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

var _ gonumGraph.Graph = new(DirectedGraph)
var _ gonumGraph.Undirected = new(DirectedGraph)
var _ gonumGraph.Weighted = new(DirectedGraph)

var _ Graph = new(DirectedGraph)

type DirectedGraph struct {
	wug        *multi.WeightedDirectedGraph
	nameToInfo map[string]Info
	knowns     []string
}

func NewDirectedGraph(indvs []string) *DirectedGraph {
	return &DirectedGraph{
		wug:        multi.NewWeightedDirectedGraph(),
		nameToInfo: make(map[string]Info, len(indvs)),
		knowns:     indvs,
	}
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

func (graph *DirectedGraph) Info(n string) Info {
	return graph.nameToInfo[n]
}

func (graph *DirectedGraph) AddInfo(name string, info Info) {
	graph.nameToInfo[name] = info
}

func (graph *DirectedGraph) PruneToShortest(indvs []string) Graph {
	g := NewDirectedGraph(indvs)
	for fromI := range indvs {
		if src := graph.NodeNamed(indvs[fromI]); src != nil {
			if shortest, ok := path.BellmanFordFrom(src, graph); ok {
				for toI := range indvs {
					if fromI == toI {
						continue
					}
					if dest := graph.NodeNamed(indvs[toI]); dest != nil {
						nodes, cost := shortest.To(dest.ID())
						names := make([]string, len(nodes)+2)
						names[0] = indvs[fromI]
						names[len(names)-1] = indvs[toI]
						if len(nodes) != 0 {
							for i := 1; i < len(names)-1; i++ {
								from := nodes[i-1]
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

func (graph *DirectedGraph) IsKnown(name string) bool {
	for i := range graph.knowns {
		if name == graph.knowns[i] {
			return true
		}
	}
	return false
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
func (graph *DirectedGraph) NameToID(name string) (int64, bool) {
	info, ok := graph.nameToInfo[name]
	return info.ID, ok
}

func (graph *DirectedGraph) Weight(xid, yid int64) (w float64, ok bool) {
	return graph.wug.Weight(xid, yid)
}

func (graph *DirectedGraph) AddNodeParentage(n, dam, sire string) {
	info := graph.nameToInfo[n]
	info.Dam = dam
	info.Sire = sire
	graph.nameToInfo[n] = info
}

func (graph *DirectedGraph) AddNodeAge(n string, age demographics.Age) {
	info := graph.nameToInfo[n]
	info.Age = age
	graph.nameToInfo[n] = info
}

func (graph *DirectedGraph) AddNodeSex(n string, sex demographics.Sex) {
	info := graph.nameToInfo[n]
	info.Sex = sex
	graph.nameToInfo[n] = info
}

func (graph *DirectedGraph) From(id int64) gonumGraph.Nodes {
	return graph.wug.From(id)
}

func (graph *DirectedGraph) FromNamed(name string) gonumGraph.Nodes {
	if info, ok := graph.nameToInfo[name]; ok {
		return graph.From(info.ID)
	}
	return gonumGraph.Empty
}

func (graph *DirectedGraph) RemoveNode(id int64) {
	graph.wug.RemoveNode(id)
}

func (graph *DirectedGraph) RemoveNodeNamed(name string) {
	if info, ok := graph.nameToInfo[name]; ok {
		graph.RemoveNode(info.ID)
		delete(graph.nameToInfo, name)
	}
}

func (graph *DirectedGraph) AddNode(n gonumGraph.Node) {
	graph.wug.AddNode(n)
}

func (graph *DirectedGraph) Nodes() gonumGraph.Nodes {
	return graph.wug.Nodes()
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

func (graph *DirectedGraph) NewNode() gonumGraph.Node {
	return graph.wug.NewNode()
}

func (graph *DirectedGraph) Edge(uid, vid int64) gonumGraph.Edge {
	return graph.wug.Edge(uid, vid)
}

func (graph *DirectedGraph) EdgeNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.Edge(uID, vID)
}

func (graph *DirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	return graph.wug.HasEdgeBetween(xid, yid)
}

func (graph *DirectedGraph) EdgeBetween(xid, yid int64) gonumGraph.Edge {
	return graph.wug.Edge(xid, yid)
}

func (graph *DirectedGraph) EdgeBetweenNamed(n1, n2 string) gonumGraph.Edge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.EdgeBetween(uID, vID)
}

func (graph *DirectedGraph) HasEdgeBetweenNamed(n1, n2 string) bool {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.HasEdgeBetween(uID, vID)
}

func (graph *DirectedGraph) WeightedEdge(uid, vid int64) gonumGraph.WeightedEdge {
	return graph.wug.WeightedEdge(uid, vid)
}

func (graph *DirectedGraph) WeightedEdgeNamed(n1, n2 string) gonumGraph.WeightedEdge {
	uID, uOK := graph.NameToID(n1)
	vID, vOK := graph.NameToID(n2)
	if uOK && vOK {
		return graph.wug.WeightedEdge(uID, vID)
	}
	return nil
}

func (graph *DirectedGraph) Node(id int64) gonumGraph.Node {
	return graph.wug.Node(id)
}

func (graph *DirectedGraph) NodeNamed(name string) gonumGraph.Node {
	if id, ok := graph.NameToID(name); ok {
		return graph.wug.Node(id)
	}
	return gonumGraph.Empty.Node()
}

func (graph *DirectedGraph) Edges() gonumGraph.Edges {
	return graph.wug.Edges()
}

func (graph *DirectedGraph) NewWeightedLine(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedLine {
	return graph.wug.NewWeightedLine(from, to, weight)
}

func (graph *DirectedGraph) NewWeightedLineNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedLine {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.NewWeightedLine(graph.Node(uID), graph.Node(vID), float64(weight))
}
func (graph *DirectedGraph) SetWeightedLine(e gonumGraph.WeightedLine) {
	graph.wug.SetWeightedLine(e)
}
