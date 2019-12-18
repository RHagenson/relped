#!/bin/bash
set -ev

relatedness=$1

result=0
trap 'result=1' ERR

# Go tests
go test -v -cover -race ./...

# Minimum running case
relped build \
    --relatedness=$relatedness \
    --output=/dev/null

# With parentage
relped build \
    --relatedness=$relatedness \
    --output=/dev/null \
    --parentage=example-data/parentage.csv

# With demographics
relped build \
    --relatedness=$relatedness \
    --output=/dev/null \
    --demographics=example-data/demographics.csv 

# With parentage and demographics
relped build \
    --relatedness=$relatedness \
    --output=/dev/null \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv

# --rm-arrows creates undirected graph rather than directed digraph
relped build \
    --relatedness=$relatedness \
    --output=/tmp/relped-out.txt \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv \
    --rm-arrows \
&& grep -q "graph " /tmp/relped-out.txt

# Directed equivalent without --rm-arrows
relped build \
    --relatedness=$relatedness \
    --output=/tmp/relped-out.txt \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv \
&& (grep -q "digraph " /tmp/relped-out.txt || rm /tmp/relped-out.txt)

# Input file extension does not matter, using <(...) causes no extension as it is a pipe
relped build \
    --relatedness=<(cat $relatedness) \
    --output=/dev/null

# Check for graceful exit on absent --output
relped build \
    --relatedness=$relatedness 2>&1 \
| grep -q 'Error: required flag(s) "output" not set'

# Check for graceful exit on absent --relatedness
relped build \
    --output=/dev/null 2>&1 \
| grep -q 'Error: required flag(s) "relatedness" not set'

# Check for fatal exit on optional input have ID not in required input
relped build \
    --relatedness=<( head $relatedness ) \
    --output=/dev/null \
    --parentage=example-data/parentage.csv 2>&1 \
| grep -q 'Cancelled further processing due to previous errors'

# Check for fatal exit on optional input have ID not in required input
relped build \
    --relatedness=<( head $relatedness ) \
    --output=/dev/null \
    --demographics=example-data/demographics.csv 2>&1 \
| grep -q 'Cancelled further processing due to previous errors'

# --unmapped produces a list of unmapped IDs
relped build \
    --relatedness=<( head -n 1 $relatedness && tail -n +1 $relatedness | shuf -n 20 ) \
    --output=/dev/null \
    --unmapped=/tmp/relped-unmapped.txt \
&& [[ -f /tmp/relped-unmapped.txt && -s /tmp/relped-unmapped.txt ]]  # Checks that file exists (-f) and has a size (-s)

exit "$result"
