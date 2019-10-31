package main

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/rhagenson/relped/internal/graph"
	"github.com/rhagenson/relped/internal/pedigree"
	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// The maximum estimatable relational distance accroding to [@doi:10.1016/j.ajhg.2016.05.020]
const maxdist = 9

// Required flags
var (
	fIn       = pflag.String("input", "", "Input file (optional)")
	fOut      = pflag.String("output", "", "Output file (required)")
	fMLRelate = pflag.String("ml-relate", "", "Input ML-Relate file (optional, implies --max-distance=3)")
)

// General use flags
var (
	opNormalize     = pflag.Bool("normalize", false, "Normalize relatedness to [0,1]-bounded")
	opHelp          = pflag.Bool("help", false, "Print help and exit")
	opKeepUnrelated = pflag.Bool("keep-unrelated", false, "Keep disconnect/unrelated individuals in pedigree")
	opMaxDist       = pflag.Uint("max-distance", maxdist, "Max relational distance to incorporate.")
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
	case *fMLRelate != "" && *opMaxDist == maxdist:
		const maxMLDist = 3 // ML-Relate does not handle relationships beyond distance of 3 (i.e.: PO, FS, HS)
		*opMaxDist = maxMLDist
		log.Infof("Setting --max-distance=%d\n", maxMLDist)
	}

	// Warning states
	switch {
	case *fMLRelate != "" && *opNormalize:
		log.Warnf("Normalizing relatedness scores with ML-Relate input has no effect.\n")
	case maxdist < *opMaxDist:
		log.Warnf("Estimating relational distance beyond %d is ill-advised.", maxdist)
	}

	// Failure states
	switch {
	case *fOut == "":
		pflag.Usage()
		log.Fatalf("Must provide both an output name.\n")
	case *fIn == "" && *fMLRelate == "":
		pflag.Usage()
		log.Fatalf("One of --input or --ml-relate is required.\n")
	case *fMLRelate != "" && 3 < *opMaxDist:
		log.Fatalf("ML-Relate does not handle distance > 3, set --max-distance <= 3. Set at: %d\n", *opMaxDist)
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Read in CSV input
	switch {
	case *fIn != "":
		in, err := os.Open(*fIn)
		defer in.Close()
		if err != nil {
			log.Errorf("Could not read input file: %s\n", err)
		}
		inCsv := csv.NewReader(in)
		inCsv.FieldsPerRecord = 3 // Simple three column format: Indv1, Indv2, Relatedness
		records, err := inCsv.ReadAll()
		if err != nil {
			log.Errorf("Problem parsing line: %s\n", err)
		}

		// Remove header
		records = util.RmHeader(records)

		// Extract relatedness values
		vals := make([]float64, len(records))
		for rowI, rowV := range records {
			if val, err := strconv.ParseFloat(rowV[2], 64); err == nil {
				vals[rowI] = val
			} else {
				log.Errorf("Could not read entry as float: %s\n", err)
			}
		}

		// Optionally normalize values
		if *opNormalize {
			vals = util.Normalize(vals)
		} else {
			vals = util.RmZeros(vals)
		}

		// Build graph
		g := graph.NewGraph()
		// Add paths from node to node based on relational distance
		for i := range records {
			if dist, rel := util.RelToLevel(vals[i]); rel { // Related at some distance
				if dist <= *opMaxDist {
					indv1 := records[i][0]
					indv2 := records[i][1]
					if indv1 != indv2 {
						g.AddUnknownPath(indv1, indv2, dist, vals[i])
					}
				}
			}
		}
		// Remove disconnected individuals
		if !*opKeepUnrelated {
			g.RmDisconnected()
		}
		// Prune edges to only the shortest between two knowns
		g = g.PruneToShortest()

		// Write the outout
		ped := pedigree.NewPedigree()

		it := g.WeightedEdges()
		for {
			if ok := it.Next(); ok {
				e := it.WeightedEdge()
				node1 := g.NameFromID(e.From().ID())
				node2 := g.NameFromID(e.To().ID())
				ped.AddNode(node1)
				ped.AddNode(node2)
				ped.AddEdge(node1, node2)
			} else {
				break
			}
		}
		if out, err := os.Create(*fOut); err == nil {
			out.WriteString(ped.String())
			out.Close()
		}
	case *fMLRelate != "":
		in, err := os.Open(*fMLRelate)
		defer in.Close()
		if err != nil {
			log.Errorf("Could not read input file: %s\n", err)
		}
		inCsv := csv.NewReader(in)
		// Columns:
		// Ind1, Ind2, R, LnL.R., U, HS, FS, PO, Relationships, Relatedness
		inCsv.FieldsPerRecord = 10
		records, err := inCsv.ReadAll()
		if err != nil {
			log.Errorf("Problem parsing line: %s\n", err)
		}

		// Remove header
		records = util.RmHeader(records)

		// Extract relatedness distance and values
		dists := make([]uint, len(records))
		vals := make([]float64, len(records))
		for rowI, rowV := range records {
			if dist, err := util.MLRelateToDist(rowV[2]); err == nil {
				dists[rowI] = dist
			} else {
				log.Errorf("Did not recognize codified entry: %s\n", err)
			}
			if val, err := strconv.ParseFloat(rowV[9], 64); err == nil {
				vals[rowI] = val
			} else {
				log.Errorf("Could not read entry as float: %s\n", err)
			}
		}

		// Optionally normalize values
		if *opNormalize {
			vals = util.Normalize(vals)
		} else {
			vals = util.RmZeros(vals)
		}
		// Build graph
		g := graph.NewGraph()
		// Add paths from node to node based on relational distance
		for i := range records {
			dist := dists[i]
			if dist <= *opMaxDist {
				indv1 := records[i][0]
				indv2 := records[i][1]
				if indv1 != indv2 {
					g.AddUnknownPath(indv1, indv2, dist, vals[i])
				}
			}
		}
		// Remove disconnected individuals
		if !*opKeepUnrelated {
			g.RmDisconnected()
		}
		// Prune edges to only the shortest between two knowns
		g = g.PruneToShortest()

		// Write the outout
		ped := pedigree.NewPedigree()
		it := g.WeightedEdges()
		for {
			if ok := it.Next(); ok {
				e := it.WeightedEdge()
				node1 := g.NameFromID(e.From().ID())
				node2 := g.NameFromID(e.To().ID())
				ped.AddNode(node1)
				ped.AddNode(node2)
				ped.AddEdge(node1, node2)
			} else {
				break
			}
		}
		if out, err := os.Create(*fOut); err == nil {
			out.WriteString(ped.String())
			out.Close()
		}
	}
	return
}

type CsvInput interface {
	Indvs() []string
	Relatedness(i1, i2 string) float64
	RelDistance(i1, i2 string) uint
}

type ThreeColumnCsv struct {
	rels     map[string]map[string]float64
	indvs    []string
	min, max float64
}

var _ CsvInput = new(ThreeColumnCsv)

func NewThreeColumnCsv(f *os.File, normalize bool) *ThreeColumnCsv {
	in, err := os.Open(*fIn)
	defer in.Close()
	if err != nil {
		log.Errorf("Could not read input file: %s\n", err)
	}
	inCsv := csv.NewReader(in)
	inCsv.FieldsPerRecord = 3 // Simple three column format: Indv1, Indv2, Relatedness
	records, err := inCsv.ReadAll()
	if err != nil {
		log.Errorf("Problem parsing line: %s\n", err)
	}
	records = util.RmHeader(records)

	c := &ThreeColumnCsv{
		rels:  make(map[string]map[string]float64, len(records)),
		indvs: make([]string, 0, len(records)),
		min:   0,
		max:   1,
	}

	indvMap := make(map[string]struct{}, len(records))

	for i := range records {
		i1 := records[i][0]
		i2 := records[i][1]
		rel := 0.0
		if val, err := strconv.ParseFloat(records[i][2], 64); err == nil {
			rel = val
		}
		c.rels[i1][i2] = rel
		if rel < c.min {
			c.min = rel
		}
		if c.max < rel {
			c.max = rel
		}
		indvMap[i1] = struct{}{}
		indvMap[i2] = struct{}{}
	}
	c.indvs = make([]string, 0, len(indvMap))
	for indv := range indvMap {
		c.indvs = append(c.indvs, indv)
	}

	if normalize {
		for i1, m := range c.rels {
			for i2, rel := range m {
				c.rels[i1][i2] = (rel - c.min) / (c.max - c.min)
			}
		}
	}

	return c
}

func (c *ThreeColumnCsv) Indvs() []string {
	return c.indvs
}

func (c *ThreeColumnCsv) Relatedness(i1, i2 string) float64 {
	return c.rels[i1][i2]
}
func (c *ThreeColumnCsv) RelDistance(i1, i2 string) uint {
	return util.RelToLevel(c.Relatedness(i1, i2))
}
