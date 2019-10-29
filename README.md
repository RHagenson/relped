# `relped`

`relped` builds a pedigree from relatedness.

## Input

Early version of `relped` use a three-column CSV format.

Example:

```
Indv1, Indv2, Relatedness
 1234,  5678,        0.50
...
```

Note that in this format, the position of columns matters, but the header of columns does not. Entries are textually considered as: "`Indv1` is related to `Indv2` by a ratio of `Relatedness`."

Later versions will include the ability to read ML-Relate's output format.

Example:

```
Ind1, Ind2, R, LnL.R.,     U,   HS, FS,   PO, Relationships, Relatedness
 612,  608,FS, -82.43, 15.47, 5.71,  -, 0.86,            FS, 0.6426
...
```

Note that in this format, the position of columns matters, but the header of solumns does not. Entries are textually considered as: "`Ind2` is a `R` of `Ind1`, but could also be any of `Relationships` with a measured relatedness of `Relatedness`."

## Output

`relped` produces a Graphviz `.dot` file for graphing. Later versions might include output to common phylogenetic tree formats.

The following formating rules are automatically applied:

```
...
graph [charset="UTF-8", rankdir="TB", splines="ortho"]
node [fontname="Sans", shape="record"]
...
```

# Contributing

If you find a bug, have a feature request, or otherwise would like to contact the author concerning issues with using `relped` please open an issue on the GitHub repo (<https://github.com/RHagenson/relped/issues>)

# License

This work is licensed under the the MIT License, see full terms of use in [LICENSE](./LICENSE) file.
