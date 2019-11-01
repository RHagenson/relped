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

func NewPedigreeFromGraph(g *graph.Graph, indvs map[string]struct{}) *Pedigree {
	ped := NewPedigree()

	it := g.WeightedEdges()
	for {
		if ok := it.Next(); ok {
			e := it.WeightedEdge()
			i1 := g.NameFromID(e.From().ID())
			i2 := g.NameFromID(e.To().ID())
			_, i1Known := indvs[i1]
			_, i2Known := indvs[i2]
			if i1Known {
				ped.AddKnownIndv(i1)
			} else {
				ped.AddUnknownIndv(i1)
			}
			if i2Known {
				ped.AddKnownIndv(i2)
			} else {
				ped.AddUnknownIndv(i2)
			}
			if i1Known && i2Known {
				ped.AddKnownRel(i1, i2)
			} else { // Bo
				ped.AddUnknownRel(i1, i2)
			}
		} else {
			break
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
