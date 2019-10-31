package util

import "github.com/gonum/floats"

// Normalize adjusts the distribution of values to be bounded in [0,1]
func Normalize(vals []float64) []float64 {
	// Adding {0,1} causes normalization to [0,1]
	// only if there exist values < 0 or > 1
	// I.e., if the values are already in the range (0,1), do nothing.
	min := floats.Min(append(vals, 0, 1))
	max := floats.Max(append(vals, 0, 1))
	for i, val := range vals {
		vals[i] = (val - min) / (max - min)
	}
	return vals
}
