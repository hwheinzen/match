package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func usage() {
	fmt.Println("Compare files as line sets and write matching lines to 'intersection files'. Input files must be sorted in ascending order. Output files are named by the corresponding matching pattern (e.g.: YN - lines only found in first input file; YY - lines found in both input files).")
	fmt.Println("\nusage:", pgmname, "[OPTIONS] FILENAME...\n")
	flag.PrintDefaults()
}

// Args reads the comand arguments and returns them as filenames if all are valid.
func args() (filenames []string, opts optT) {

	var help bool
	flag.BoolVar(&help, "help", false, "prints usage information")
	var nodup bool
	flag.BoolVar(&nodup, "nodup", false, "ignores duplicate lines within a input file")
	var out string
	flag.StringVar(&out, "out", "", "output only for these patterns (comma separated list, e.g. -out=NYY,NYN)")

	flag.Usage = usage
	flag.Parse()

	if help {
		usage()
		os.Exit(0) // not an error
	}

	if nodup {
		opts.nodup = true
	}

	if len(flag.Args()) < 1 {
		fmt.Println("nothing to match")
		usage()
		os.Exit(1)
	}

	for _, v := range flag.Args() {
		assertFile(v)
	}

	if out != "" {
		opts.outs = strings.Split(out, ",")
		for _, v := range opts.outs {
			if len(v) != len(flag.Args()) { // pattern length == number of input files
				log.Fatalln("impossible pattern for -out:", v)
			}
		}
	}

	return flag.Args(), opts
}

func assertFile(filename string) {
	fi, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	if fi.IsDir() {
		log.Fatalln(filename, "is not a file")
	}
	if fi.Size() == 0 {
		log.Fatalln(filename, "is empty")
	}
}
