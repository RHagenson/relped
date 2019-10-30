package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/awalterschulze/gographviz"
	"github.com/rs/xid"
	"github.com/spf13/pflag"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// Required flags
var (
	fIn       = pflag.String("input", "", "Input file (optional)")
	fOut      = pflag.String("output", "", "Output file (required)")
	fMLRelate = pflag.String("ml-relate", "", "Input ML-Relate file (optional)")
)

// General use flags
var (
	opNormalize = pflag.Bool("normalize", false, "Normalize relatedness to [0,1]-bounded")
	opHelp      = pflag.Bool("help", false, "Print help and exit")
	opRmUnrel   = pflag.Bool("rm-unrelated", true, "Remove unrelated individuals from pedigree")
	opMaxDist   = pflag.Uint("max-distance", 9, "Max relational distance to incorporate.")
)

// setup runs the CLI initialization prior to program logic
func setup() {
	pflag.Parse()
	if *opHelp {
		pflag.Usage()
		os.Exit(1)
	}

	// Failure states
	switch {
	case *fOut == "":
		pflag.Usage()
		Errorf("Must provide both an output name.\n")
		os.Exit(1)
	case *fIn == "" && *fMLRelate == "":
		pflag.Usage()
		Errorf("One of --input or --ml-relate is required.\n")
		os.Exit(1)
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Read in CSV input
	switch {
	case *fIn != "":
		in, err := os.Open(*fIn)
		defer in.Close()
		if err != nil {
			Errorf("Could not read input file: %s\n", err)
		}
		inCsv := csv.NewReader(in)
		inCsv.FieldsPerRecord = 3 // Simple three column format: Indv1, Indv2, Relatedness
		records, err := inCsv.ReadAll()
		if err != nil {
			Errorf("Problem parsing line: %s\n", err)
		}

		// Remove header
		records = records[1:]

		// Extract relatedness values
		vals := make([]float64, len(records))
		for rowI, rowV := range records {
			if val, err := strconv.ParseFloat(rowV[2], 64); err == nil {
				vals[rowI] = val
			} else {
				log.Fatalf("Could not read entry as float: %s\n", err)
			}
		}

		// Optionally normalize values
		if *opNormalize { // Normalize
			vals = normalize(vals)
		} else {
			for i, v := range vals { // Replace negatives as unrelated (i.e., 0)
				if v < 0 {
					vals[i] = 0
				}
			}
		}

		// Build graph
		g := NewGraph()
		// Add paths from node to node based on relational distance
		for i := range records {
			if dist, rel := relToLevel(vals[i]); rel { // Related at some distance
				if dist <= *opMaxDist {
					indv1 := records[i][0]
					indv2 := records[i][1]
					if indv1 != indv2 {
						g.AddUnknownPath(indv1, indv2, dist, vals[i])
					}
				}
			}
		}
		// Remove disconnected individuals
		if *opRmUnrel {
			g.RmDisconnected()
		}
		// TODO: Prune edges based on shortest/cheapest path between
		// two known individuals
		// Not implemented

		// Write the outout
		ped := NewPedigree()

		it := g.WeightedEdges()
		for {
			if ok := it.Next(); ok {
				e := it.WeightedEdge()
				node1 := g.NameFromID(e.From().ID())
				node2 := g.NameFromID(e.To().ID())
				ped.AddNode(node1)
				ped.AddNode(node2)
				ped.AddEdge(node1, node2)
			} else {
				break
			}
		}
		if out, err := os.Create(*fOut); err == nil {
			out.WriteString(ped.String())
		}
	case *fMLRelate != "":
		fmt.Fprint(os.Stderr, "ML-Relate input is still experimental.")

	}
	return
}

type Pedigree struct {
	g *gographviz.Graph
}

func NewPedigree() *Pedigree {
	g := gographviz.NewGraph()
	g.SetDir(false)
	g.SetName("pedigree")
	graphAttrs := map[string]string{
		"rankdir": "TB",
		"splines": "ortho",
	}
	for attr, val := range graphAttrs {
		g.AddAttr("pedigree", attr, val)
	}
	return &Pedigree{
		g: g,
	}
}

func (p *Pedigree) AddNode(node string) error {
	nodeAttrs := map[string]string{
		"fontname": "Sans",
		"shape":    "record",
	}
	return p.g.AddNode(p.g.Name, node, nodeAttrs)
}

func (p *Pedigree) AddEdge(src, dst string) error {
	return p.g.AddEdge(src, dst, p.g.Directed, nil)
}

func (p *Pedigree) String() string {
	return p.g.String()
}

// Graph has named nodes/vertexes
type Graph struct {
	g *simple.WeightedUndirectedGraph
	m map[string]graph.Node
}

func NewGraph() *Graph {
	return &Graph{
		g: simple.NewWeightedUndirectedGraph(0, 0),
		m: make(map[string]graph.Node),
	}
}

func (self *Graph) Nodes() graph.Nodes {
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

func (self *Graph) From(name string) graph.Nodes {
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

func (self *Graph) Edge(n1, n2 string) graph.Edge {
	uid := self.m[n1].ID()
	vid := self.m[n2].ID()
	return self.g.Edge(uid, vid)
}

func (self *Graph) Node(name string) graph.Node {
	return self.g.Node(self.m[name].ID())
}

func (self *Graph) Edges() graph.Edges {
	return self.g.Edges()
}

func (self *Graph) WeightedEdges() graph.WeightedEdges {
	return self.g.WeightedEdges()
}

func (self *Graph) NewWeightedEdge(n1, n2 string, weight float64) graph.WeightedEdge {
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
		path[i] = xid.New().String()
	}
	weights := make([]float64, len(path)-1)
	for i := range weights {
		weights[i] = incWeight
	}
	self.AddPath(path, weights)
}

// normalize adjusts the distribution of values to be bounded in [0,1]
func normalize(vals []float64) []float64 {
	// Adding {0,1} causes normalization to [0,1]
	// only if there exist values < 0 or > 1
	// I.e., if the values are already in the range (0,1), do nothing.
	min := floats.Min(append(vals, 0, 1))
	max := floats.Max(append(vals, 0, 1))
	for i, val := range vals {
		vals[i] = (val - min) / (max - min)
	}
	return vals
}

// Errorf standardizes notifying user of failure and failing
func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(2)
}

// relToLevel computes the relational distance given the relatedness score
//
// Examples:
//     relToLevel(0.5)   --> (1, true)
//     relToLevel(0.25)  --> (2, true)
//     relToLevel(0.125) --> (3, true)
//     relToLevel(<=0)   --> (0, false) // Only "unrelated" case
func relToLevel(x float64) (uint, bool) {
	if x <= 0 {
		return 0, false
	}
	return uint(math.Round(math.Log(1/x) / math.Log(2))), true
}
