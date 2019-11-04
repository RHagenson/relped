package pedigree

import (
	"github.com/awalterschulze/gographviz"
	"github.com/rhagenson/relped/internal/graph"
)

// "Constant" maps for attributes
var (
	knownIndvAttrs = map[string]string{
		"fontname": "Sans",
		"shape":    "record",
		"style":    "solid",
	}
	unknownIndvAttrs = map[string]string{
		"fontname": "Sans",
		"shape":    "record",
		"style":    "dashed",
	}
	knownRelAttrs = map[string]string{
		"style": "solid",
	}
	unknownRelAttrs = map[string]string{
		"style": "dashed",
	}
	graphAttrs = map[string]string{
		"rankdir":  "TB",
		"splines":  "ortho",
		"ratio":    "auto",
		"mincross": "2.0",
	}
)

type Pedigree struct {
	g *gographviz.Escape
}

func NewPedigree() *Pedigree {
	g := gographviz.NewEscape()
	g.SetDir(false)
	g.SetName("pedigree")
	for attr, val := range graphAttrs {
		g.AddAttr("pedigree", attr, val)
	}
	return &Pedigree{
		g: g,
	}
}

func NewPedigreeFromGraph(g *graph.Graph, indvs []string) *Pedigree {
	ped := NewPedigree()

	iter := g.Edges()
	for iter.Next() {
		e := iter.Edge()

		n1, n1OK := g.IDToName(e.From().ID())
		if n1OK {
			ped.AddKnownIndv(n1)
		} else {
			ped.AddUnknownIndv(n1)
		}

		n2, n2OK := g.IDToName(e.To().ID())
		if n2OK {
			ped.AddKnownIndv(n2)
		} else {
			ped.AddUnknownIndv(n2)
		}

		if n1OK && n2OK {
			ped.AddKnownRel(n1, n2)
		} else {
			ped.AddUnknownRel(n1, n2)
		}
	}
	return ped
}

func (p *Pedigree) AddKnownIndv(node string) error {
	return p.g.AddNode(p.g.Name, node, knownIndvAttrs)
}

func (p *Pedigree) AddUnknownIndv(node string) error {
	return p.g.AddNode(p.g.Name, node, unknownIndvAttrs)
}

func (p *Pedigree) AddKnownRel(src, dst string) error {
	return p.g.AddEdge(src, dst, p.g.Directed, knownRelAttrs)
}

func (p *Pedigree) AddUnknownRel(src, dst string) error {
	return p.g.AddEdge(src, dst, p.g.Directed, unknownRelAttrs)
}

func (p *Pedigree) String() string {
	return p.g.String()
}
