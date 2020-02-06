# match
Command that compares arbitrary many files line by line


## Overview
The command `match` compares arbitrary many input files line by line and creates output files for the possible matching patterns. 

Input files must be sorted in ascendng order, the key for matching is the line without the newline character. 

Matching patterns form the output file names, e.g. for matching two files the output file named `YN` contains the lines only found in the first input file, `NY` contains the lines found in the second input file, and `YY` contains the lines found in both files. 


## Download
Provided you have Go installed, run:

`$ go get github.com/hwheinzen/match`


## Help
`$ match -help`


## Bugs
`match` seems to be pretty stable.
