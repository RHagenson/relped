package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// Required flags
var (
	fIn  = pflag.String("input", "", "Input file (required)")
	fOut = pflag.String("output", "", "Output file (required)")
)

// General use flags
var (
	opNormalize = pflag.Bool("normalize", false, "Normalize relatedness to [0,1]-bounded")
	opHelp      = pflag.Bool("help", false, "Print help and exit")
	opRmUnrel   = pflag.Bool("rm-unrelated", true, "Remove unrelated individuals from pedigree")
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
	case *fIn == "" || *fOut == "":
		pflag.Usage()
		Errorf("Must provide both an input and output name.\n")
		os.Exit(1)
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Read in CSV input
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
		for i, v := range vals { // Replace negatives with
			if v < 0 {
				vals[i] = 0
			}
		}
	}

	// Build graph
	g := new(Graph)
	// Add known vertexes/nodes
	for i := range records {
		indv1 := records[i][1]
		indv2 := records[i][2]
		g.AddNode(indv1)
		g.AddNode(indv2)
	}
	// Add edges based on relational distance
	for i := range records {
		dist := relToLevel(vals[i])
		n1 := records[i][0]
		n2 := records[i][1]
		weight := vals[i] / float64(dist)
		switch dist {
		case 0:
			continue
		case 1:
			g.AddEdge(n1, n2, weight, nil)
			g.AddEdge(n2, n1, weight, nil)
		case 2:
			g.AddVertex(unknownID, nil)
			// "Forward"
			g.AddEdge(n1, unknownID, weight, nil)
			g.AddEdge(unknownID, n2, weight, nil)
			// "Reverse"
			g.AddEdge(n2, unknownID, weight, nil)
			g.AddEdge(unknownID, n1, weight, nil)
			unknownID++
		}
	}
}

// Graph has named nodes/vertexes
type Graph struct {
	g simple.WeightedUndirectedGraph
	m map[string]graph.Node
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

func (self *Graph) Edges() graph.Edges {
	return self.g.Edges()
}

func (self *Graph) NewWeightedEdge(n1, n2 string, weight float64) graph.WeightedEdge {
	uid := self.m[n1]
	vid := self.m[n2]
	return self.g.NewWeightedEdge(uid, vid, weight)
}

func (self *Graph) AddPath(names []string, weights []float64) {
	if len(weights) != len(names)-1 {
		log.Fatalf("Weights along path should be one less than names along path.")
	}
	for i := 1; i <= len(names); i++ {
		self.NewWeightedEdge(names[i-1], names[i], weights[i-1])
	}
}

// normalize adjusts the distribution of values to be bounded in [0,1]
func normalize(vals []float64) []float64 {
	// Adding {0,1} causes normalization to [0,1]
	// only if there exist values < 0 or > 1
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
//     relToLevel(0.5)   --> 1
//     relToLevel(0.25)  --> 2
//     relToLevel(0.125) --> 3
//     relToLevel(<=0)   --> MaxUint64
func relToLevel(x float64) uint {
	if x <= 0 {
		return math.MaxUint64
	}
	return uint(math.Round(math.Log(1/x) / math.Log(2)))
}
