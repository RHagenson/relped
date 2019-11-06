package main

import (
	"flag"
	"os"

	"github.com/rhagenson/relped/internal/csvin"
	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/pedigree"
	"github.com/rhagenson/relped/internal/unit/relational"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var maxDist = relational.Ninth

// Required flags
var (
	fThreeColumn = pflag.String("three-column", "", "Input standard three-column file (optional)")
	fMLRelate    = pflag.String("ml-relate", "", "Input ML-Relate file (optional)")
	fOut         = pflag.String("output", "", "Output file (required)")
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

	// Information states
	switch {
	case *fMLRelate != "" && *opMaxDistance == uint(maxDist):
		// ML-Relate does not handle relationships beyond distance of 3 (i.e.: PO, FS, HS)
		maxDist = relational.Third
		*opMaxDistance = uint(maxDist)
		log.Infof("Setting --max-distance=%d\n", maxDist)
	}

	// Warning states
	switch {
	case *fMLRelate != "" && *opNormalize:
		log.Warnf("Normalizing relatedness scores with ML-Relate input has no effect.\n")
	case uint(maxDist) < *opMaxDistance:
		log.Warnf("Estimating relational distance beyond %d is ill-advised.", maxDist)
	}

	// Failure states
	switch {
	case *fOut == "":
		pflag.Usage()
		log.Fatalf("Must provide both an output name.\n")
	case *fThreeColumn == "" && *fMLRelate == "":
		pflag.Usage()
		log.Fatalf("One of --input or --ml-relate is required.\n")
	case *fMLRelate != "" && 3 < *opMaxDistance:
		log.Fatalf("ML-Relate does not handle distance > 3, set --max-distance <= 3. Set at: %d\n", *opMaxDistance)
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Read in CSV input
	switch {
	case *fThreeColumn != "":
		// Open input file
		in, err := os.Open(*fThreeColumn)
		defer in.Close()
		if err != nil {
			log.Fatalf("Could not read input file: %s\n", err)
		}
		input := csvin.NewThreeColumnCsv(in, *opNormalize)

		// Build graph
		g := graph.NewGraphFromCsvInput(input, maxDist)

		// Prune edges to only the shortest between two knowns
		indvs := input.Indvs()
		g = g.PruneToShortest(indvs)

		// Write the outout
		ped := pedigree.NewPedigreeFromGraph(g, indvs)
		out, err := os.Create(*fOut)
		defer out.Close()
		if err != nil {
			log.Fatalf("Could not create output file: %s\n", err)
		}
		out.WriteString(ped.String())
	case *fMLRelate != "":
		in, err := os.Open(*fMLRelate)
		defer in.Close()
		if err != nil {
			log.Errorf("Could not read input file: %s\n", err)
		}

		input := csvin.NewMLRelateCsv(in, *opNormalize)
		indvs := input.Indvs()

		// Build graph
		g := graph.NewGraphFromCsvInput(input, maxDist)

		// Prune edges to only the shortest between two knowns
		g = g.PruneToShortest(indvs)

		// Write the outout
		ped := pedigree.NewPedigreeFromGraph(g, indvs)
		out, err := os.Create(*fOut)
		defer out.Close()
		if err != nil {
			log.Fatalf("Could not create output file: %s\n", err)
		}
		out.WriteString(ped.String())
	}
	return
}
