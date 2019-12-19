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

Given the relatedness of a set of individuals as calculated by tools such a ML-Relate [@ML-Relate], in the past the compiling biologist had to manually draw the resulting pedigree chart via manual reasoning on inferred connections --`relped` automates this process. By using a combination of relatedness values (with optional normalization) and codified entries, `relped` allows incremental refinement of the resulting pedigree network by replacing values with codes as familial determinations are definitively called. As well, there are options to include parentage and/or demographics information to better match the typical appearance of a standard pedigree chart. With parentage information provided, `relped` can differentiate between parent-offspring and full-sibling (both $\approx 0.5$ related) relationships. Pedigree network connections made between individuals are consistent between runs of `relped`; therefore, the pedigree network can be redrawn by repeatedly running `relped` until the pedigree network is visually satisfactory. Given the findings of Staples et al. ([@Staples2016]), by default relationships beyond the 9th-level (e.g., 4th cousins) are not mapped. Any individuals related beyond this cutoff are left unmapped, but a list of these IDs can be saved.

# Statement of Need

While relatedness can be calculated from a pedigree ([@Lynch2018]), there are not tools to build a pedigree chart from relatedness alone. This is primarily due to the amount of information required to produce the hierarchical tree structure within a pedigree chart. The valuable information from a pedigree chart is how individuals within the sampled population connect.

The primary use of `relped` is for visualization of genetic relationships when in-depth parentage information is not readily available. This tool creates a pedigree network to attain a better understanding of the connections formed from genetic pairwise comparisons within a sampled population.

A pedigree chart is a visual depiction of relationships across generations. A pedigree chart relies heavily on complete family histories and parentage information, which may be unknown or unattainable in some populations. Relationships between individuals are then reliant on calculated genetic relatedness ($r_{xy}$; i.e., [@ML-Relate], [@Queller1989], [@Peakall2012]).

In contrast to a pedigree chart, `relped` creates a pedigree network to illustrate genetic connections to visualize relationships in a sampled population. The network can be generated with or without known parentage or demographic information. A dataset with minimal information (e.g., $r_{xy}$ only of a subset of the population) will produce a network with genetic connections based on the information provided. As more information is provided, connections are better resolved. The resulting network built from relationship calls, parentage, and demographic information for all individuals in a population should result in a network that most closely resembles a pedigree chart.

# References
