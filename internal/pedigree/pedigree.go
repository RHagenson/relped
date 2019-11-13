package pedigree

import (
	"github.com/awalterschulze/gographviz"
	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/io/demographics"
)

// "Constant" maps for attributes
var (
	knownIndvAttrs = map[string]string{
		"fontname":  "Sans",
		"shape":     "record",
		"style":     "filled",
		"fillcolor": "yellow",
	}
	unknownIndvAttrs = map[string]string{
		"fontname": "Sans",
		"shape":    "diamond",
		"style":    "dashed",
	}
	knownRelAttrs = map[string]string{
		"style":    "solid",
		"penwidth": "2.5",
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

func NewPedigreeFromGraph(g *graph.Graph, indvs []string, dems demographics.CsvInput) *Pedigree {
	ped := NewPedigree()

	iter := g.Edges()
	for iter.Next() {
		e := iter.Edge()

		from, _ := g.IDToName(e.From().ID())
		to, _ := g.IDToName(e.To().ID())
		fromKnown := g.IsKnown(from)
		toKnown := g.IsKnown(to)
		if fromKnown {
			if dems != nil {
				if fromSex, ok := dems.Sex(from); ok {
					ped.AddKnownIndv(from, fromSex)
				} else {
					ped.AddKnownIndv(from, demographics.Unknown)
				}
			} else {
				ped.AddKnownIndv(from, demographics.Unknown)
			}
		} else {
			ped.AddUnknownIndv(from)
		}

		if toKnown {
			if dems != nil {
				if toSex, ok := dems.Sex(to); ok {
					ped.AddKnownIndv(to, toSex)
				} else {
					ped.AddKnownIndv(to, demographics.Unknown)
				}
			} else {
				ped.AddKnownIndv(to, demographics.Unknown)
			}
		} else {
			ped.AddUnknownIndv(to)
		}

		if fromKnown && toKnown {
			ped.AddKnownRel(from, to)
		} else {
			ped.AddUnknownRel(from, to)
		}
	}
	return ped
}

func (p *Pedigree) AddKnownIndv(node string, sex demographics.Sex) error {
	attrs := knownIndvAttrs
	switch sex {
	case demographics.Female:
		attrs["shape"] = "ellipse"
	case demographics.Male:
		attrs["shape"] = "box"
	case demographics.Unknown:
		attrs["shape"] = "record"
	default:
		attrs["shape"] = "record"
	}

	return p.g.AddNode(p.g.Name, node, attrs)
}

func (p *Pedigree) AddUnknownIndv(node string) error {
	attrs := unknownIndvAttrs
	return p.g.AddNode(p.g.Name, node, attrs)
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
