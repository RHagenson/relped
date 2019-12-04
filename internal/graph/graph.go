package graph

import (
	"fmt"
	"math"

	mapset "github.com/deckarep/golang-set"
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/parentage"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

const lenUnknownNames = 6

var _ gonumGraph.Graph = new(Graph)
var _ gonumGraph.Undirected = new(Graph)
var _ gonumGraph.Weighted = new(Graph)

// Graph has named nodes/vertexes
type Graph struct {
	wug        *simple.WeightedUndirectedGraph
	nameToInfo map[string]Info
	knowns     []string
}

type Info struct {
	ID        int64
	Sex       demographics.Sex
	Age       demographics.Age
	Dam, Sire string
}

func NewGraph(indvs []string) *Graph {
	return &Graph{
		wug:        simple.NewWeightedUndirectedGraph(math.MaxFloat64, math.MaxFloat64),
		nameToInfo: make(map[string]Info, len(indvs)),
		knowns:     indvs,
	}
}

func NewGraphFromCsvInput(in relatedness.CsvInput, maxDist relational.Degree, pars parentage.CsvInput, dems demographics.CsvInput) *Graph {
	indvs := in.Indvs()
	strIndvs := make([]string, 0, indvs.Cardinality())
	for _, indv := range indvs.ToSlice() {
		strIndvs = append(strIndvs, indv.(string))
	}
	g := NewGraph(strIndvs)

	// Add any unknowns to link knowns by relational distance
	for i := range strIndvs {
		for j := range strIndvs {
			if i == j {
				continue
			} else {
				from := strIndvs[i]
				to := strIndvs[j]
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

	// Add parentage
	if pars != nil {
		children := pars.Indvs()
		for _, child := range children {
			degree := relational.First
			relatedness := unit.Relatedness(1.0)
			if degree <= maxDist {
				if sire, ok := pars.Sire(child); ok {
					g.AddParentage(child, "", sire)
					edge := g.NewWeightedEdgeNamed(sire, child, relatedness.Weight())
					g.SetWeightedEdge(edge)
				}
				if dam, ok := pars.Dam(child); ok {
					g.AddParentage(child, "", dam)
					edge := g.NewWeightedEdgeNamed(dam, child, relatedness.Weight())
					g.SetWeightedEdge(edge)
				}
			}
		}
	}

	// Add demographics
	if dems != nil {
		for _, indv := range strIndvs {
			if age, ok := dems.Age(indv); ok {
				g.AddAge(indv, age)
			}
			if sex, ok := dems.Sex(indv); ok {
				g.AddSex(indv, sex)
			}
		}
	}

	return g
}

func (graph *Graph) AddAge(name string, age demographics.Age) {
	info := graph.nameToInfo[name]
	info.Age = age
	graph.nameToInfo[name] = info
}
func (graph *Graph) AddSex(name string, sex demographics.Sex) {
	info := graph.nameToInfo[name]
	info.Sex = sex
	graph.nameToInfo[name] = info
}
func (graph *Graph) AddParentage(name, dam, sire string) {
	info := graph.nameToInfo[name]
	if dam != "" {
		info.Dam = dam
	}
	if sire != "" {
		info.Sire = sire
	}
	graph.nameToInfo[name] = info
}
func (graph *Graph) Info(name string) Info {
	return graph.nameToInfo[name]
}
func (graph *Graph) AddInfo(name string, info Info) {
	graph.nameToInfo[name] = info
}

func (graph *Graph) PruneToShortest(keepLoops bool) *Graph {
	indvs := graph.knowns
	connected := mapset.NewSet()

	for i := 0; i < len(indvs); i++ {
		if src := graph.NodeNamed(indvs[i]); src != nil {
			if shortest, ok := path.BellmanFordFrom(src, graph); ok {
				for j := i + 1; j < len(indvs); j++ {
					if dest := graph.NodeNamed(indvs[j]); dest != nil {
						nodes, _ := shortest.To(dest.ID())
						for _, node := range nodes {
							connected.Add(node)
						}
					}
				}
			}
		}
	}

	nodes := graph.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		if !connected.Contains(n) {
			graph.RemoveNode(n.ID())
		}
		if !keepLoops {
			graph.RemoveEdge(n.ID(), n.ID())
		}
	}

	return graph
}

func (graph *Graph) IsKnown(name string) bool {
	for i := range graph.knowns {
		if name == graph.knowns[i] {
			return true
		}
	}
	return false
}

func (graph *Graph) AddPath(p Path) {
	names := p.Names()
	weights := p.Weights()

	for i := range weights {
		from := names[i]
		to := names[i+1]
		weight := weights[i]
		graph.AddNodeNamed(from)
		graph.AddNodeNamed(to)
		edge := graph.NewWeightedEdgeNamed(from, to, weight)
		graph.SetWeightedEdge(edge)
	}
}

// IDToName converts the id to its corresponding node name
// Returns false if the node does not exist
func (graph *Graph) IDToName(id int64) (string, bool) {
	for name, info := range graph.nameToInfo {
		if info.ID == id {
			return name, true
		}
	}
	return "", false
}

// NameToID converts the name to its corresponding node ID
// Returns false if the node does not exist
func (graph *Graph) NameToID(name string) (int64, bool) {
	info, ok := graph.nameToInfo[name]
	return info.ID, ok
}

func (graph *Graph) RmDisconnected() {
	for name := range graph.nameToInfo {
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
	xinfo := graph.nameToInfo[n1]
	yinfo := graph.nameToInfo[n2]
	return graph.Weight(xinfo.ID, yinfo.ID)
}

func (graph *Graph) From(id int64) gonumGraph.Nodes {
	return graph.wug.From(id)
}

func (graph *Graph) FromNamed(name string) gonumGraph.Nodes {
	if info, ok := graph.nameToInfo[name]; ok {
		return graph.From(info.ID)
	}
	return gonumGraph.Empty
}

func (graph *Graph) RemoveNode(id int64) {
	graph.wug.RemoveNode(id)
}

func (graph *Graph) RemoveNodeNamed(name string) {
	if info, ok := graph.nameToInfo[name]; ok {
		graph.RemoveNode(info.ID)
		delete(graph.nameToInfo, name)
	}
}

func (graph *Graph) AddNode(n gonumGraph.Node) {
	graph.wug.AddNode(n)
}

func (graph *Graph) Nodes() gonumGraph.Nodes {
	return graph.wug.Nodes()
}

func (graph *Graph) AddNodeNamed(name string) {
	if _, ok := graph.nameToInfo[name]; !ok {
		n := graph.NewNode()
		graph.AddNode(n)
		info := graph.nameToInfo[name]
		info.ID = n.ID()
		graph.nameToInfo[name] = info
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

func (graph *Graph) NewWeightedEdge(from, to gonumGraph.Node, weight float64) gonumGraph.WeightedEdge {
	return graph.wug.NewWeightedEdge(from, to, weight)
}

func (graph *Graph) NewWeightedEdgeNamed(n1, n2 string, weight unit.Weight) gonumGraph.WeightedEdge {
	uID, _ := graph.NameToID(n1)
	vID, _ := graph.NameToID(n2)
	return graph.NewWeightedEdge(graph.Node(uID), graph.Node(vID), float64(weight))
}

func (graph *Graph) SetWeightedEdge(e gonumGraph.WeightedEdge) {
	graph.wug.SetWeightedEdge(e)
}

func (graph *Graph) RemoveEdge(fid, tid int64) {
	graph.wug.RemoveEdge(fid, tid)
}

func (graph *Graph) String() string {
	nodes := graph.wug.Nodes()
	edges := graph.wug.Edges()
	return fmt.Sprintf("Graph:\n\tNodes(%d):\n\t%v\n\tEdges(%d):\n\t%t\nMap:\n%v\nmap[527]==%v\n", nodes.Len(), nodes, edges.Len(), edges, graph.nameToInfo, graph.nameToInfo["527"])
}
