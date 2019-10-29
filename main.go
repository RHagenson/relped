package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// Required flags
var (
	fIn  = pflag.String("input", "", "Input file (required)")
	fOut = pflag.String("output", "", "Output file (required)")
)

// General use flags
var (
	opCorrect = pflag.Bool("correct", false, "Correct Relatedness to be [0,1]-bounded")
	opHelp    = pflag.Bool("help", false, "Print help and exit")
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
}

// Errorf standardizes notifying user of failure and failing
func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(2)
}
