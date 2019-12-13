#!/bin/bash
set -ev

result=0
trap 'result=1' ERR

# Go tests
go test -v -cover -race ./...

# Minimum running case
relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=/dev/null

# With parentage
relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=/dev/null \
    --parentage=example-data/parentage.csv

# With demographics
relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=/dev/null \
    --demographics=example-data/demographics.csv 

# With parentage and demographics
relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=/dev/null \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv

# --rm-arrows creates undirected graph rather than directed digraph
relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=/tmp/relped-out.txt \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv \
    --rm-arrows \
&& grep -q "graph " /tmp/relped-out.txt

# Directed equivalent without --rm-arrows
relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=/tmp/relped-out.txt \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv \
&& (grep -q "digraph " /tmp/relped-out.txt || rm /tmp/relped-out.txt)

# Input file extension does not matter, using <(...) causes no extension as it is a pipe
relped build \
    --relatedness=<(cat example-data/relatedness-nums-and-codes.csv) \
    --output=/dev/null

exit "$result"
