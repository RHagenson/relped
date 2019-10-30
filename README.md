# `relped`

[![DOI](https://zenodo.org/badge/217557856.svg)](https://zenodo.org/badge/latestdoi/217557856)

`relped` builds a pedigree from relatedness.

## Input

`relped` can use either a three-column CSV format:

Example:

```
Indv1, Indv2, Relatedness
 1234,  5678,        0.50
...
```

or the ten-columns format from [ML-Relate](http://www.montana.edu/kalinowski/software/ml-relate/index.html):

```
Ind1, Ind2,  R, LnL.R.,     U,   HS, FS,   PO, Relationships, Relatedness
 612,  608, FS, -82.43, 15.47, 5.71,  -, 0.86,            FS,      0.6426
...
```

Note that in either format, the position of columns matters, but the header of columns does not.

## Output

`relped` produces a Graphviz-formatted file with attributes generally deemed useful for building pedigree images including `rankdir="TB", splines="ortho"`.


# Contributing

If you find a bug, have a feature request, or otherwise would like to contact the author concerning issues with using `relped` please open an issue on the GitHub repo (<https://github.com/RHagenson/relped/issues>)

# License

This work is licensed under the the MIT License, see full terms of use in [LICENSE](./LICENSE) file.
