package cmd

import (
	"os"

	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/parentage"
	"github.com/rhagenson/relped/internal/io/relatedness"
	"github.com/rhagenson/relped/internal/pedigree"
	"github.com/rhagenson/relped/internal/unit/relational"
	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var maxDist = relational.Ninth

// Required flags
var (
	fRelatedness  string
	fDemographics string
	fParentage    string
	fOut          string
)

// General use flags
var (
	opNormalize   bool
	opMaxDistance uint
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build relatedness pedigree",
	Long: `Use pairwise relatedness scores in addition to optional
demographics and parentage information to build an effective pedigree, 
generating the necessary number of unknown individuals.`,
	Run: func(cmd *cobra.Command, args []string) {
		build()
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Required flags
	buildCmd.Flags().StringVar(&fRelatedness, "relatedness", "", "Three-column relatedness file (required)")
	buildCmd.MarkFlagRequired("relatedness")
	buildCmd.Flags().StringVar(&fOut, "output", "", "Output DOT file (required)")
	buildCmd.MarkFlagRequired("output")

	// Optional inputs
	buildCmd.Flags().StringVar(&fDemographics, "demographics", "", "Three-column demographics file")
	buildCmd.Flags().StringVar(&fParentage, "parentage", "", "Three-column parentage file")

	// Behavioral changes
	buildCmd.Flags().BoolVar(&opNormalize, "normalize", false, "Normalize relatedness to [0,1]-bounded")
	buildCmd.Flags().UintVar(&opMaxDistance, "max-distance", uint(relational.Ninth), "Max relational distance to incorporate")
}

// setup runs the CLI initialization prior to program logic
func setup() {
	// Set maxDist
	maxDist = relational.Degree(opMaxDistance)

	// Information states
	// None

	// Warning states
	switch {
	case uint(maxDist) < opMaxDistance:
		log.Warnf("Estimating relational distance beyond %d is ill-advised.", maxDist)
	}

	// Failure states
	switch {
	case fOut == "":
		pflag.Usage()
		log.Fatalf("Must provide --output.\n")
	case fRelatedness == "":
		pflag.Usage()
		log.Fatalf("Must provide --relatedness.\n")
	}
}

func build() {
	// Parse CLI arguments
	setup()

	var (
		input relatedness.CsvInput
		dems  demographics.CsvInput
		pars  parentage.CsvInput
	)

	// Read in CSV input
	in, err := os.Open(fRelatedness)
	defer in.Close()
	if err != nil {
		log.Fatalf("Could not read input file: %s\n", err)
	}
	input = relatedness.NewThreeColumnCsv(in, opNormalize)

	// Open demographics file
	if fDemographics != "" {
		inDem, err := os.Open(fDemographics)
		defer inDem.Close()
		if err != nil {
			log.Fatalf("Could not read demographics file: %s\n", err)
		}
		dems = demographics.NewThreeColumnCsv(inDem)
	}

	// Open parentage file
	if fParentage != "" {
		inPar, err := os.Open(fParentage)
		defer inPar.Close()
		if err != nil {
			log.Fatalf("Could not read parentage file: %s\n", err)
		}
		pars = parentage.NewThreeColumnCsv(inPar)
	}

	// Check demographics and parentage for consistency
	if msg := util.DemsAndParsAgree(dems, pars); msg != "" {
		log.Fatalf("The demographics and parentage files disagree:\n%s", msg)
	}

	// Build graph
	g := graph.NewGraphFromCsvInput(input, maxDist, pars)

	// Prune edges to only the shortest between two knowns
	indvs := input.Indvs()
	g = g.PruneToShortest(indvs)

	// Write the outout
	ped := pedigree.NewPedigreeFromGraph(g, indvs, dems)
	out, err := os.Create(fOut)
	defer out.Close()
	if err != nil {
		log.Fatalf("Could not create output file: %s\n", err)
	}
	out.WriteString(ped.String())
	return
}