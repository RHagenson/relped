package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"gonum.org/v1/gonum/floats"
)

// Required flags
var (
	fIn  = pflag.String("input", "", "Input file (required)")
	fOut = pflag.String("output", "", "Output file (required)")
)

// General use flags
var (
	opNormalize = pflag.Bool("normalize", false, "Normalize relatedness to [0,1]-bounded")
	opHelp      = pflag.Bool("help", false, "Print help and exit")
)

func setup() {
	pflag.Parse()
	if *opHelp {
		pflag.Usage()
		os.Exit(1)
	}

	// Failure states
	switch {
	case *fIn == "" || *fOut == "":
		pflag.Usage()
		Errorf("Must provide both an input and output name.\n")
		os.Exit(1)
	}
}

func main() {
	// Parse CLI arguments
	setup()

	in, err := os.Open(*fIn)
	defer in.Close()
	if err != nil {
		Errorf("Could not read input file: %s", err)
	}
	inCsv := csv.NewReader(in)
	inCsv.FieldsPerRecord = 3 // Simple three column format: Indv1, Indv2, Relatedness
	records, err := inCsv.ReadAll()
	if err != nil {
		Errorf("Problem parsing line: %s", err)
	}
	fmt.Print(records)
	if *opNormalize {
		vals := normalize(records)
	}
}

// normalize adjusts the distribution of values to be bounded in [0,1]
func normalize(rs [][]string) []float64 {
	vals := make([]float64, len(rs))
	for rowI, rowV := range rs {
		if val, err := strconv.ParseFloat(rowV[2], 64); err == nil {
			vals[rowI] = val
		} else {
			log.Printf("Could not read entry as float: %s", err)
		}
	}

	// Normalize to [0,1] only if there exist
	// values < 0 or > 1
	min := floats.Min(append(vals, 0, 1))
	max := floats.Max(append(vals, 0, 1))
	for i, val := range vals {
		vals[i] = (val - min) / (max - min)
	}
	return vals

}

// Errorf standardizes notifying user of failure and failing
func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(2)
}

// MinMax finds the minimum and maximum values in one pass
func MinMax(array []int) (int, int) {
	var max int = array[0]
	var min int = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}
