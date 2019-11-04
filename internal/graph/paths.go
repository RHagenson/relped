package graph

import (
	"log"

	"github.com/rhagenson/relped/internal/unit"
	"github.com/rs/xid"
)

type Path interface {
	Names() []string
	Weights() []unit.Weight
}

var _ Path = new(BasicPath)

type BasicPath struct {
	names   []string
	weights []unit.Weight
}

func (p *BasicPath) Names() []string {
	return p.names
}

func (p *BasicPath) Weights() []unit.Weight {
	return p.weights
}

func NewBasicPath(names []string, weights []unit.Weight) *BasicPath {
	if len(weights) != len(names)-1 {
		log.Fatalf("Weights along path should be one less than names along path.")
	}
	return &BasicPath{names, weights}
}

func (self *Graph) AddPath(p Path) {
	names := p.Names()
	weights := p.Weights()

	for i := range weights {
		self.AddNodeNamed(names[i])
		self.AddNodeNamed(names[i+1])
		self.NewWeightedEdgeNamed(names[i], names[i+1], weights[i])
	}
}

type EqualWeightPath struct {
	names  []string
	weight unit.Weight
}

func (p EqualWeightPath) Names() []string {
	return p.names
}

func (p EqualWeightPath) Weights() []unit.Weight {
	weights := make([]unit.Weight, len(p.names)-1)
	for i := range weights {
		weights[i] = p.weight
	}
	return weights
}

func NewEqualWeightPath(names []string, weight unit.Weight) *EqualWeightPath {
	return &EqualWeightPath{names, weight}
}

type FractionalWeightPath struct {
	names  []string
	weight unit.Weight
}

func (p FractionalWeightPath) Names() []string {
	return p.names
}

func (p FractionalWeightPath) Weights() []unit.Weight {
	weights := make([]unit.Weight, len(p.names)-1)
	fracWeight := float64(p.weight) / float64(len(weights))
	for i := range weights {
		weights[i] = unit.Weight(fracWeight)
	}
	return weights
}

func NewFractionalWeightPath(names []string, weight unit.Weight) *FractionalWeightPath {
	return &FractionalWeightPath{names, weight}
}

type RelationalWeightPath struct {
	p *FractionalWeightPath
}

func (p RelationalWeightPath) Names() []string {
	return p.p.names
}

func (p RelationalWeightPath) Weights() []unit.Weight {
	return p.p.Weights()
}

func NewRelationalWeightPath(n1, n2 string, dist unit.GraphDistance, weight unit.Weight) *RelationalWeightPath {
	names := make([]string, dist+2)
	// Add knowns
	names[0] = n1
	names[len(names)-1] = n2

	for i := range names {
		if i == 0 || i == len(names)-1 {
			continue
		} else {
			name := xid.New().String()
			names[i] = name[len(name)-lenUnknownNames:]
		}
	}
	return &RelationalWeightPath{&FractionalWeightPath{names, weight}}
}
