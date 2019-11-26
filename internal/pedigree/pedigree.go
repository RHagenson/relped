package pedigree

import (
	"fmt"
	"strings"

	"github.com/awalterschulze/gographviz"
	mapset "github.com/deckarep/golang-set"
	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/io/demographics"
	log "github.com/sirupsen/logrus"
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
		"style": "bold",
	}
	unknownRelAttrs = map[string]string{
		"style": "dashed",
	}
	graphAttrs = map[string]string{
		"rankdir": "TB",
		"splines": "ortho",
		"ratio":   "auto",
		"newrank": "true",
	}
)

type Pedigree struct {
	g     *gographviz.Escape
	ranks map[demographics.Age][]string
}

func NewUndirectedPedigree() *Pedigree {
	g := gographviz.NewEscape()
	g.SetDir(false)
	g.SetName("pedigree")
	for attr, val := range graphAttrs {
		g.AddAttr("pedigree", attr, val)
	}
	return &Pedigree{
		g:     g,
		ranks: make(map[demographics.Age][]string),
	}
}

func NewDirectedPedigree() *Pedigree {
	g := gographviz.NewEscape()
	g.SetDir(true)
	g.SetName("pedigree")
	for attr, val := range graphAttrs {
		g.AddAttr("pedigree", attr, val)
	}
	return &Pedigree{
		g:     g,
		ranks: make(map[demographics.Age][]string),
	}
}

func NewPedigreeFromGraph(g graph.Graph, indvs []string) (*Pedigree, []string) {
	var ped *Pedigree
	switch g {
	case g.(*graph.DirectedGraph):
		log.Infof("Built directed pedigree\n")
		ped = NewDirectedPedigree()
	case g.(*graph.UndirectedGraph):
		log.Infof("Built undirected pedigree\n")
		ped = NewUndirectedPedigree()
	}

	mapped := mapset.NewSet()
	var unmapped []string

	iter := g.Edges()
	for iter.Next() {
		e := iter.Edge()

		from, _ := g.IDToName(e.From().ID())
		to, _ := g.IDToName(e.To().ID())
		fromKnown := g.IsKnown(from)
		toKnown := g.IsKnown(to)
		if fromKnown {
			mapped.Add(from)
			fromInfo := g.Info(from)
			if fromInfo.Sex != demographics.Unknown {
				ped.AddKnownIndv(from, fromInfo.Sex)
			} else {
				ped.AddKnownIndv(from, demographics.Unknown)
			}
			if fromInfo.Age != 0 {
				ped.AddToRank(fromInfo.Age, from)
			}
		} else {
			ped.AddUnknownIndv(from)
		}

		if toKnown {
			mapped.Add(to)
			toInfo := g.Info(to)
			if toInfo.Sex != demographics.Unknown {
				ped.AddKnownIndv(to, toInfo.Sex)
			} else {
				ped.AddKnownIndv(to, demographics.Unknown)
			}
			if toInfo.Age != 0 {
				ped.AddToRank(toInfo.Age, to)
			}
		} else {
			ped.AddUnknownIndv(to)
		}

		if fromKnown && toKnown {
			fromInfo := g.Info(from)
			toInfo := g.Info(to)
			if fromInfo.Age >= toInfo.Age {
				ped.AddKnownRel(from, to)
			} else {
				ped.AddKnownRel(to, from)
			}
		} else {
			ped.AddUnknownRel(from, to)
		}
	}

	for _, indv := range indvs {
		if !mapped.Contains(indv) {
			if unmapped == nil {
				unmapped = make([]string, 0, len(indvs)-mapped.Cardinality())
			}
			unmapped = append(unmapped, indv)
		}
	}

	return ped, unmapped
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
	attrs := knownRelAttrs
	return p.g.AddEdge(src, dst, p.g.Directed, attrs)
}

func (p *Pedigree) AddUnknownRel(src, dst string) error {
	return p.g.AddEdge(src, dst, p.g.Directed, unknownRelAttrs)
}

func (p *Pedigree) String() string {
	out := p.g.String()
	ranks := new(strings.Builder)
	for age, indvs := range p.ranks {
		if len(indvs) > 1 {
			ranks.WriteString("\t{rank=same; ")
			ranks.WriteString(strings.Join(indvs, ", "))
			ranks.WriteString(fmt.Sprintf(" }; // Age: %d \n", age))
		}
	}
	out = out[:len(out)-2] + ranks.String() + "}\n"
	return out
}

func (p *Pedigree) AddToRank(a demographics.Age, id string) {
	for _, indv := range p.ranks[a] {
		if indv == id {
			return
		}
	}
	p.ranks[a] = append(p.ranks[a], id)
}
