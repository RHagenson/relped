package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/parentage"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/pedigree"
	"github.com/rhagenson/relped/internal/unit/relational"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var maxDist = relational.Ninth

// Required flags
var (
	fIn           = pflag.String("input", "", "Input standard three-column file  (required)")
	fDemographics = pflag.String("demographics", "", "Input demographics file (optional)")
	fParentage    = pflag.String("parentage", "", "Input parentage file (optional)")
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
		pars  parentage.CsvInput
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

		// Open demographics file
		if *fDemographics != "" {
			inDem, err := os.Open(*fDemographics)
			defer inDem.Close()
			if err != nil {
				log.Fatalf("Could not read demographics file: %s\n", err)
			}
			dems = demographics.NewThreeColumnCsv(inDem)
		}

		// Open parentage file
		if *fParentage != "" {
			inPar, err := os.Open(*fParentage)
			defer inPar.Close()
			if err != nil {
				log.Fatalf("Could not read parentage file: %s\n", err)
			}
			pars = parentage.NewThreeColumnCsv(inPar)
		}

		// Check demographics and parentage for consistency
		if msg := DemsAndParsAgree(dems, pars); msg != "" {
			log.Fatalf("The demographics and parentage files disagree:\n%s", msg)
		}

		// Build graph
		g := graph.NewGraphFromCsvInput(input, maxDist, pars)

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

func DemsAndParsAgree(dems demographics.CsvInput, pars parentage.CsvInput) string {
	var s strings.Builder
	for _, indv := range pars.Indvs() {
		if sire, ok := pars.Sire(indv); ok {
			if sex, ok := dems.Sex(sire); ok {
				if sex != demographics.Male {
					s.WriteString(fmt.Sprintf("Sire %s for ID %s should be male, but is stated as %s in demographics\n", sire, indv, sex))
				}
			}
		}
		if dam, ok := pars.Dam(indv); ok {
			if sex, ok := dems.Sex(dam); ok {
				if sex != demographics.Female {
					s.WriteString(fmt.Sprintf("Dam %s for ID %s should be female, but is stated as %s in demographics\n", dam, indv, sex))
				}
			}
		}
	}
	return s.String()
}
