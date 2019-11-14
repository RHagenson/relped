package main

import (
	"flag"
	"os"

	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/pedigree"
	"github.com/rhagenson/relped/internal/unit/relational"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var maxDist = relational.Ninth

// Required flags
var (
	fIn           = pflag.String("input", "", "Input standard three-column file")
	fDemographics = pflag.String("demographics", "", "Input demographics file (optional)")
	fOut          = pflag.String("output", "", "Output file (required)")
)

// General use flags
var (
	opNormalize   = pflag.Bool("normalize", false, "Normalize relatedness to [0,1]-bounded")
	opHelp        = pflag.Bool("help", false, "Print help and exit")
	opMaxDistance = pflag.Uint("max-distance", uint(relational.Ninth), "Max relational distance to incorporate.")
	cpuprofile    = flag.String("cpuprofile", "", "write cpu profile to file")
)

// setup runs the CLI initialization prior to program logic
func setup() {
	pflag.Parse()
	if *opHelp {
		pflag.Usage()
		os.Exit(1)
	}

	// Set maxDist
	maxDist = relational.Degree((*opMaxDistance))

	// Information states
	// None

	// Warning states
	switch {
	case uint(maxDist) < *opMaxDistance:
		log.Warnf("Estimating relational distance beyond %d is ill-advised.", maxDist)
	}

	// Failure states
	switch {
	case *fOut == "":
		pflag.Usage()
		log.Fatalf("Must provide --output.\n")
	case *fIn == "":
		pflag.Usage()
		log.Fatalf("Must provide --input.\n")
	}
}

func main() {
	// Parse CLI arguments
	setup()

	var (
		input relatedness.CsvInput
		dems  demographics.CsvInput
	)

	// Read in CSV input
	if *fIn != "" {
		// Open input file
		in, err := os.Open(*fIn)
		defer in.Close()
		if err != nil {
			log.Fatalf("Could not read input file: %s\n", err)
		}
		input = relatedness.NewThreeColumnCsv(in, *opNormalize)

		inDem, err := os.Open(*fDemographics)
		defer inDem.Close()
		if err != nil {
			log.Fatalf("Could not read demographics file: %s\n", err)
		}
		dems = demographics.NewThreeColumnCsv(inDem)

		// Build graph
		g := graph.NewGraphFromCsvInput(input, maxDist)

		// Prune edges to only the shortest between two knowns
		indvs := input.Indvs()
		g = g.PruneToShortest(indvs)

		// Write the outout
		ped := pedigree.NewPedigreeFromGraph(g, indvs, dems)
		out, err := os.Create(*fOut)
		defer out.Close()
		if err != nil {
			log.Fatalf("Could not create output file: %s\n", err)
		}
		out.WriteString(ped.String())
	}

	return
}
