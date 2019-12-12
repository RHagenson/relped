#!/bin/bash
set -ev

result=0
trap 'result=1' ERR

# Go tests
go test -v -race ./...

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

exit "$result"
