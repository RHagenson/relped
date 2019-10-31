package pedigree

import (
	"github.com/awalterschulze/gographviz"
	"github.com/rhagenson/relped/internal/graph"
)

type Pedigree struct {
	g *gographviz.Graph
}

func NewPedigree() *Pedigree {
	g := gographviz.NewGraph()
	g.SetDir(false)
	g.SetName("pedigree")
	graphAttrs := map[string]string{
		"rankdir":  "TB",
		"splines":  "ortho",
		"ratio":    "auto",
		"mincross": "2.0",
	}
	for attr, val := range graphAttrs {
		g.AddAttr("pedigree", attr, val)
	}
	return &Pedigree{
		g: g,
	}
}

func NewPedigreeFromGraph(g *graph.Graph) *Pedigree {
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
	return ped
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
