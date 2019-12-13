package pedigree_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/pedigree"
)

// "Constant" maps for attributes
// Copies of expected attributes (double bookkeeping)
var (
	knownIndvAttrs = map[string]string{
		"fontname":  "Sans",
		"shape":     "record",
		"style":     "filled",
		"fillcolor": "yellow",
	}
	unknownIndvAttrs = map[string]string{
		"fontname": "Sans",
		"shape":    "diamond",
		"style":    "dashed",
		"label":    "\"\"",
	}
	knownRelAttrs = map[string]string{
		"style": "bold",
	}
	unknownRelAttrs = map[string]string{
		"style": "dashed",
	}
	graphAttrs = map[string]string{
		"rankdir": "TB",
		"splines": "ortho",
		"ratio":   "auto",
		"newrank": "true",
	}
)

func TestPedigree(t *testing.T) {
	t.Run("known individual attributes", func(t *testing.T) {
		p := pedigree.NewPedigree()
		p.AddKnownIndv("I1", demographics.Unknown)
		if pattern, err := regexp.Compile("I1.*"); err == nil {
			line := pattern.FindString(p.String())
			if line == "" {
				t.Errorf("regex failed to match")
			} else {
				for attr, val := range knownIndvAttrs {
					str := attr + "=" + val
					t.Run(str, func(t *testing.T) {
						if !strings.Contains(line, str) {
							t.Errorf("expected %s in line: %s", str, line)
						}
					})
				}
			}
		} else {
			t.Errorf("regex to find known individual attributes failed to compile")
		}
	})

	t.Run("unknown individual attributes", func(t *testing.T) {
		p := pedigree.NewPedigree()
		p.AddUnknownIndv("U1")
		if pattern, err := regexp.Compile("U1.*"); err == nil {
			line := pattern.FindString(p.String())
			if line == "" {
				t.Errorf("regex failed to match")
			} else {
				for attr, val := range unknownIndvAttrs {
					str := attr + "=" + val
					t.Run(str, func(t *testing.T) {
						if !strings.Contains(line, str) {
							t.Errorf("expected %s in line: %s", str, line)
						}
					})
				}
			}
		} else {
			t.Errorf("regex to find unknown individual attributes failed to compile")
		}
	})

	t.Run("unknown relationship attributes", func(t *testing.T) {
		p := pedigree.NewPedigree()
		p.AddUnknownIndv("U1")
		p.AddUnknownIndv("U2")
		p.AddUnknownRel("U1", "U2")
		if pattern, err := regexp.Compile("U1->U2.*"); err == nil {
			line := pattern.FindString(p.String())
			if line == "" {
				t.Errorf("regex failed to match")
			} else {
				for attr, val := range unknownRelAttrs {
					str := attr + "=" + val
					t.Run(str, func(t *testing.T) {
						if !strings.Contains(line, str) {
							t.Errorf("expected %s in line: %s", str, line)
						}
					})
				}
			}
		} else {
			t.Errorf("regex to find unknown relationship attributes failed to compile")
		}
	})

	t.Run("known relationship attributes", func(t *testing.T) {
		p := pedigree.NewPedigree()
		p.AddUnknownIndv("U1")
		p.AddUnknownIndv("U2")
		p.AddKnownRel("U1", "U2")
		if pattern, err := regexp.Compile("U1->U2.*"); err == nil {
			line := pattern.FindString(p.String())
			if line == "" {
				t.Errorf("regex failed to match")
			} else {
				for attr, val := range knownRelAttrs {
					str := attr + "=" + val
					t.Run(str, func(t *testing.T) {
						if !strings.Contains(line, str) {
							t.Errorf("expected %s in line: %s", str, line)
						}
					})
				}
			}
		} else {
			t.Errorf("regex to find known relationship attributes failed to compile")
		}
	})

	t.Run("sex changes shape", func(t *testing.T) {
		p := pedigree.NewPedigree()
		p.AddKnownIndv("Male", demographics.Male)
		p.AddKnownIndv("Female", demographics.Female)
		p.AddKnownIndv("Unknown", demographics.Unknown)

		for _, sex := range []string{"Male", "Female", "Unknown"} {
			if pattern, err := regexp.Compile(sex + ".*"); err == nil {
				line := pattern.FindString(p.String())
				if line == "" {
					t.Errorf("regex failed to match")
				} else {
					switch sex {
					case "Male":
						t.Run("shape=box for Male", func(t *testing.T) {
							if !strings.Contains(line, "shape=box") {
								t.Errorf("expected %s in line: %s", "shape=box", line)
							}
						})
					case "Female":
						t.Run("shape=ellipse for Female", func(t *testing.T) {
							if !strings.Contains(line, "shape=ellipse") {
								t.Errorf("expected %s in line: %s", "shape=ellipse", line)
							}
						})
					case "Unknown":
						t.Run("shape=record for Unknown", func(t *testing.T) {
							if !strings.Contains(line, "shape=record") {
								t.Errorf("expected %s in line: %s", "shape=record", line)
							}
						})
					}
				}
			} else {
				t.Errorf("regex to find known relationship attributes failed to compile")
			}
		}
	})

	t.Run("ranks are added properly", func(t *testing.T) {
		p := pedigree.NewPedigree()
		p.AddUnknownIndv("U1")
		p.AddUnknownIndv("U2")
		p.AddToRank(demographics.Age(10), "U1")
		p.AddToRank(demographics.Age(10), "U2")
		if pattern, err := regexp.Compile("{rank=same.*"); err == nil {
			line := pattern.FindString(p.String())
			if line == "" {
				t.Errorf("regex failed to match")
			} else {
				str := "{rank=same; U1, U2 }; // Age: 10"
				if !strings.Contains(line, str) {
					t.Errorf("expected %s in line: %s", str, line)
				}
			}
		} else {
			t.Errorf("regex to find added ranks failed to compile")
		}
	})
}
