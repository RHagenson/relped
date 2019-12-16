#!/bin/bash
set -ev

relped build \
    --relatedness=example-data/relatedness-nums-and-codes.csv \
    --output=imgs/relped.dot \
    --parentage=example-data/parentage.csv \
    --demographics=example-data/demographics.csv \
&& dot -Tpng -O imgs/relped.dot \
&& rm imgs/relped.dot
