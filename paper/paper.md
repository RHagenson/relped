---
title: 'relped: Build Relatedness Pedigrees'
tags:
  - genetic
  - visualization
  - graphviz
  - relatedness
  - pedigree
authors:
 - name: Ryan A. Hagenson
   orcid: 0000-0001-9750-1925
   affiliation: "1"
 - name: Caitlin J. Curry
   orcid: 0000-0002-3853-7191
   affiliation: "1"
affiliations:
 - name: Omaha's Henry Doorly Zoo and Aquarium
   index: 1
date: 13 December 2019
bibliography: paper.bib
---

# Summary

Given only the relatedness of a set of individuals as calculate by tools such a ML-Relate [@ML-Relate], in the past the compiling biologist had to manually draw the resulting pedigree via manual deductive reasoning on possible connections. `relped` serves to automate the process including common safe inferences such as differentiating between parent-offspring and full-sibling (both $\approx 0.5$ related) relationships if enough information on parentage is known. By using a combination of relatedness values (with optional normalization) and codified entries, `relped` allows incremental refinement of the resulting pedigree by replacing values with codes as familial determinations are definitively called. As well, there are options to include parentage and/or demographics information to better match the typical appearance of a standard pedigree. Connections made between individuals are consistent between runs of `relped`, therefore the pedigree can be redrawn by repeated `relped` runs until the pedigree is visually satisfactory. Given the findings of Staples et al. ([@Staples2016]), relationships beyond the 9th-level (e.g., 4th cousins) are not mapped. Any individuals related beyond this 9th-level are left unmapped, but will appear in the unmapped individuals file if the `--unmapped` option is used.

# Statement of Need

The primary use of `relped` is for visualization of genetic relationships when in-depth parentage information is not readily available. This tool creates a pedigree network to attain a better understanding of the connections between genetic pairwise comparisons within a sampled population.

A pedigree chart is a visual depiction of relationships across generations. A pedigree chart relies heavily on complete family histories and parentage information, which may be unknown or unattainable in some populations. Relationships between individuals are then reliant on calculated genetic relatedness (i.e., [@ML-Relate], [@Queller1989], [@Peakall2012]).

While relatedness can be calculated from a pedigree ([@Lynch2018]), there are not tools to build a pedigree from relatedness alone. This is primarily due to the amount of information required to produce the hierarchical tree structure within a pedigree chart. The valuable information from a pedigree chart is how individuals within the sampled population connect. In contrast to a pedigree chart, `relped` creates a network to illustrate genetic connections to visualize relationships in a sampled population. The network can be generated with or without known parentage or demographic information. A dataset with minimal information (e.g., rXY only of a subset of the population) will produce a network with all possible genetic connections based on the information provided. As more information is provided, connections are better resolved. The resulting network of a dataset with relationship calls, parentage, and demographic information for all individuals in a population should result in a network that most closely resembles a pedigree chart.

# References
