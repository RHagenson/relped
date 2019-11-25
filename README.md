# `relped`

[![DOI](https://zenodo.org/badge/217557856.svg)](https://zenodo.org/badge/latestdoi/217557856)

`relped` builds a pedigree from relatedness.

## Installation

Before using `relped` there are a few programs you need to install:

+ [Git](https://git-scm.com/downloads): allows you to download `relped`
+ [Go](https://golang.org/dl/): allows you to compile `relped`
+ [Graphviz](https://graphviz.org/download/): allows you to render `relped`'s output

After installing the programs above, getting `relped` installed should be as easy as:

```bash
go get -u github.com/rhagenson/relped
```

## Input

### Relatedness

`relped` has one required input:

Example:

```csv
ID1,ID2,Rel
123,456,0.50
...
```

Note that your columns **must** be named `ID1`,`ID2`, and `Rel`. It is okay to have either duplicate entries of the same pair of IDs (but perhaps different `Rel`) or have entries where the IDs have been switched -- for duplicate entries, the last entry will be used.

### Pedigree

Example:

```csv
ID,Sire,Dam
123,456,789
...
```

Note that your columns **must** be named `ID`,`Sire`, and `Dam`. If your file contains duplicate ID entries, only the last entry will be used.

### Demographics

Example:

```csv
ID,Sex,BirthYear
123,Male,1990
...
```

Note that your columns **must** be named `ID`,`Sex`, and `BirthYear`. If your file contains duplicate ID entries, only the last entry will be used. `Sex` entries of either full word or first letter are recognized (e.g. `M` or `Male`) -- matching is case insensitive. `Sex` is used to change the formatting attributes in the final pedigree to distinguish males, females, and unknown sex. `BirthYear` is converted to age for the current year under the assumption that all birthdays have passed for the year (age is used to clarify the pedigree output).

## Output

`relped` produces a Graphviz-formatted file (directed or undirected, depending on input) with attributes deemed useful for building pedigrees. Unlike in a regular pedigree, all nodes at the same height in the plot may not be the same age, however all connections will be exactly the same between runs of `relped`.

### Producing multiple plots

Due to the unavoidable randomness of how the pedigree is plotted by Graphviz, the below template can be reused to build multiple plots in a row.

```bash
for run in {1..10}
do
  relped build \
    --relatedness <relatedness> \
    --demographics <demographics> \
    --parentage <parentage> \
    --output $run-<output> \
  && dot -Tsvg -O $run-<output>
done
```

What the above does is loop through the number 1-10 (assigning the current number to `run`), then calls `relped` with your inputs (`<relatedness>`, `<demographics>`, `<parentage>`) and writes multiple output files that are prepended with the run number (`$run-<output>`) -- and calls Graphviz to produce a rendered image for each plot of format stated (note there is no space in `-Tsvg`).

## Contributing

If you find a bug, have a feature request, or otherwise would like to contact the author concerning use of `relped`, please open an [issue](https://github.com/rhagenson/relped/issues).

## License

This work is licensed under the the MIT License, see full terms of use in [LICENSE](./LICENSE) file.
