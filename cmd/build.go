package cmd

import (
	"os"
	"strings"

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
	fRelatedness string
	fOut         string
)

// Optional flags
var (
	fDemographics string
	fParentage    string
	fUnmapped     string
)

// General use flags
var (
	opNormalize     bool
	opMaxDistance   uint
	opRmArrows      bool
	opKeepSelfLoops bool
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
	buildCmd.Flags().StringVar(&fUnmapped, "unmapped", "", "File of unmapped individuals from relatedness")

	// Behavioral changes
	buildCmd.Flags().BoolVar(&opNormalize, "normalize", false, "Normalize relatedness to [0,1]-bounded")
	buildCmd.Flags().UintVar(&opMaxDistance, "max-distance", uint(relational.Ninth), "Max relational distance to incorporate")
	buildCmd.Flags().BoolVar(&opRmArrows, "rm-arrows", false, "Remove arrows heads from pedigree, instead use simple lines")
	buildCmd.Flags().BoolVar(&opKeepSelfLoops, "keep-loops", false, "Keep any loops drawn between an individual and itself")
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

	// Open connections to the required files
	in, err := os.Open(fRelatedness)
	defer in.Close()
	if err != nil {
		log.Fatalf("Could not read input file: %s\n", err)
	}
	out, err := os.Create(fOut)
	defer out.Close()
	if err != nil {
		log.Fatalf("Could not create output file: %s\n", err)
	}

	// Read in CSV input
	input = relatedness.NewThreeColumnCsv(in, opNormalize)
	indvs := input.Indvs()

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

	// Issue #30: If there is an ID in optional files, but not in required files then error
	var errored = false
	for _, child := range pars.Indvs() {
		if !indvs.Contains(child) {
			log.Errorf("No corresponding relatedness data for parentage entry: %s\n", child)
			errored = true
		}
		if sire, ok := pars.Sire(child); ok {
			if !indvs.Contains(sire) {
				log.Errorf("Sire %s of parentage ID %s not found in relatedness file\n", sire, child)
				errored = true
			}
		}
		if dam, ok := pars.Dam(child); ok {
			if !indvs.Contains(dam) {
				log.Errorf("Dam %s of parentage ID %s not found in relatedness file\n", dam, child)
				errored = true
			}
		}
	}
	for _, id := range dems.Indvs() {
		if !indvs.Contains(id) {
			log.Errorf("No corresponding relatedness data for demographics entry of %s\n", id)
			errored = true
		}
	}
	if errored {
		log.Fatalf("Cancelled further processing due to previous errors\n")
	}

	// Build graph
	g := graph.NewGraphFromCsvInput(input, maxDist, pars, dems)

	// Prune edges to only the shortest between two knowns
	g.PruneToShortest(opKeepSelfLoops)

	// Write the outout
	strIndvs := make([]string, 0, indvs.Cardinality())
	for _, indv := range indvs.ToSlice() {
		strIndvs = append(strIndvs, indv.(string))
	}
	ped, unmapped := pedigree.NewPedigreeFromGraph(g, strIndvs, opRmArrows)
	if fUnmapped != "" {
		if unmapped != nil {
			un, err := os.Create(fUnmapped)
			defer un.Close()
			if err != nil {
				log.Fatalf("Could not create output file: %s\n", err)
			}
			un.WriteString(strings.Join(unmapped, "\n"))
		} else {
			log.Infof("No unmapped individuals\n")
		}
	}
	out.WriteString(ped.String())
	return
}
