package pedigree

import "github.com/awalterschulze/gographviz"

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
