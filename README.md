# match
Command that compares arbitrary many files line by line


## Overview
The command `match` compares arbitraryly many input files line by line and creates output files for the possible matching patterns. ("Arbitraryly many" is in fact limited - you'll hit the operating system limit for open files pretty soon.)

Input files must be sorted in ascending order, matching key is the line.

Matching patterns form the output file names, e.g. for matching two files the output file named `YN` contains the lines only found in the first input file, `NY` contains the lines found in the second input file, and `YY` contains the lines found in both files. 


## Download
Provided you have Go installed, run:

`$ go install github.com/hwheinzen/match@latest`

(Has been `$ go get github.com/hwheinzen/match` before.)


## Help
`$ match -help`


## Bugs
`match` seems to be stable.
