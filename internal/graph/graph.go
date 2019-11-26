package graph

import (
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/parentage"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	gonumGraph "gonum.org/v1/gonum/graph"
)

const lenUnknownNames = 6

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
	PruneToShortest(indvs []string) Graph
	Info(string) Info
	AddInfo(name string, info Info)
}

type Info struct {
	ID        int64
	Sex       demographics.Sex
	Age       demographics.Age
	Dam, Sire string
}

func NewGraphFromCsvInput(in relatedness.CsvInput, maxDist relational.Degree, pars parentage.CsvInput, dems demographics.CsvInput) Graph {
	if pars != nil && dems != nil {
		return newDirectedGraph(in, maxDist, pars, dems)
	}
	return newUndirectedGraph(in, maxDist)
}
