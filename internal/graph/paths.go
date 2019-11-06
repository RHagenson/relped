package graph

import (
	"fmt"

	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	"github.com/rs/xid"
)

type Path interface {
	Names() []string
	Weights() []unit.Weight
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

func NewRelationalWeightPath(from, to string, dist relational.Degree, weight unit.Weight) (*RelationalWeightPath, error) {
	if dist == relational.Unrelated {
		return nil, fmt.Errorf("%q and %q are unrelated, no path possible", from, to)
	}
	names := make([]string, dist+1)
	// Add knowns
	names[0] = from
	names[len(names)-1] = to
	for i := range names {
		if i == 0 || i == len(names)-1 {
			continue
		} else {
			name := xid.New().String()
			names[i] = name[len(name)-lenUnknownNames:]
		}
	}
	return &RelationalWeightPath{&FractionalWeightPath{names, weight}}, nil
}
