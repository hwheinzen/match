package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
)

const pgmname = "match"

type optT struct {
	nodup bool
	outs  []string
}

type inT struct {
	name  string
	file  *os.File
	read  *bufio.Reader
	count int
	rec   string
	eof   bool
	eod   bool
}

type outT struct {
	name  string
	file  *os.File
	writ  *bufio.Writer
	count int
}

type allT struct {
	opts optT
	ins  []inT  // input files
	outs []outT // output files
}

var pats []string          // all possible patterns; init in main()
var	filter map[string]bool // desired output files; init in main()

// Flip is a helper.
func flip(ss []string) {
	for i, s := range ss {
		if i+1 > len(ss)/2 {
			break
		}
		ss[i] = ss[len(ss)-1-i]
		ss[len(ss)-1-i] = s
	}
}

// Command match handles arbitrary input files, compares them line by line,
// and creates new files whose names reflect the match/nomatch patterns
// and contain matched lines respectively. Input files must be sorted.
func main() {
	repFile := mustCreate("Report.txt")
	defer repFile.Close()
	repWrit := bufio.NewWriter(repFile)
	defer repWrit.Flush()

	filenames, opts := args() // returns at least one filename

	var cmd string
	for _, v := range os.Args {
		cmd = cmd + v + " "
	}
	mustWrite(repWrit, "Command:\n\t$ "+cmd) // report

	flip(filenames)           // so that position in arguments matches position in matching pattern

	pats = allPats(len(filenames)) // init all possible patterns

	var all allT

	// process options
	if opts.nodup {
		all.opts.nodup = opts.nodup
	}
	if len(opts.outs) > 0 {
		all.opts.outs  = opts.outs
	}
	filter = make(map[string]bool, len(pats))
	for _, v := range pats {
		filter[v] = false
	}
	if len(all.opts.outs) > 0 {
		for _, v := range all.opts.outs { // desired output
			filter[v] = true
		}
	} else {
		for _, v := range pats { // all output
			filter[v] = true
		}
	}

	// input files
	all.ins = make([]inT, 0, len(filenames))
	var in inT
	for _, fn := range filenames {
		in.name = fn
		in.file = mustOpen(in.name)
		defer in.file.Close()
		in.read = bufio.NewReader(in.file)
		all.ins = append(all.ins, in)
	}

	// output files
	all.outs = make([]outT, 0, len(pats)-1) // no file for Ns only pattern
	var out outT
	for i, pat := range pats {
		if i == 0 {  // no file for Ns only pattern
			continue
		}
		out.name = pat
		if filter[pat] { // desired output only
			out.file = mustCreate(out.name)
			defer out.file.Close()
			out.writ = bufio.NewWriter(out.file)
			defer out.writ.Flush()
		}
		all.outs = append(all.outs, out)
	}

	err := matchAll(&all) // ACTION
	if err != nil {
		log.Fatalln(err)
	}

	mustWrite(repWrit, "\nInput files:") // report
	for i := len(all.ins)-1; i >= 0; i-- {
		v := all.ins[i]
		mustWrite(repWrit, "\t" + v.name + ": " + strconv.Itoa(v.count))
	}
	mustWrite(repWrit, "\nOutput files:") // report
	for _, v := range all.outs {
		mustWrite(repWrit, "\t" + v.name + ": " + strconv.Itoa(v.count))
	}
	

	return
}

// AllPats returns all possible Y/N patterns with length len.
// Indices of slice pats correspond to the contained patterns, 
// i.e. you get the Index if you replace all N with 0 and all Y with 1
// and read the result as as binary number.
// 
// Example: for len 2 allPats return pats = {"NN" "NY" "YN" "YY"}.
func allPats(len int) (pats []string) {
	if len == 0 { // no pattern
		return nil
	}

	pats = make([]string, 0, (1 << len))

	for i := 0; i < (1 << len); i++ { // 1<<len == 2**len
		pat, val := "", i
		for pos := len; pos > 0; pos-- {
			if val >= (1 << (pos - 1)) { // 1<<(pos-1) == 2**(pos-1)
				pat += "Y"
				val -= (1 << (pos - 1))
			} else {
				pat += "N"
			}
		}
		pats = append(pats, pat)
	}
	return pats
}

// MatchAll reads all input in a read loop and writes to the output files.
func matchAll(all *allT) (err error) {

	for pos, _ := range all.ins { // initial readings
		err = get(pos, all)
		if err != nil {
			return err
		}
	}
	curPatInx, inInxs := compare(all)

	for curPatInx != 0 {
		for patInx := range pats {
			if curPatInx == patInx { // MATCH

				if filter[pats[patInx]] { // desired output only
					// [patInx-1], because pattern with Ns only makes no output file
					mustWrite(all.outs[patInx-1].writ, all.ins[inInxs[0]].rec)
					all.outs[patInx-1].count += 1
				}

				for _, pos := range inInxs { // read next
					err = get(pos, all)
					if err != nil {
						return err
					}
				}
				curPatInx, inInxs = compare(all)
				break
			}
		}
	}

	return nil
}

// Compare compares all input lines currently in focus,
// returns the pattern index for the minimum line and all indices
// that lead to the input files containing this minimum line.
func compare(all *allT) (curPatInx int, inInxs []int) {
	if allEOD(all) {
		return 0, nil
	}

	var recs = make([]string, 0, len(all.ins))
	for pos := range all.ins {
		recs = append(recs, all.ins[pos].rec)
	}
	minRec := min(recs...)

	for pos := range all.ins {
		if all.ins[pos].rec == minRec {
			curPatInx |= (1 << pos)      // für Mustervergleich
			inInxs = append(inInxs, pos) // am Muster Beteiligte
		}
	}

	return curPatInx, inInxs
}

func allEOD(all *allT) bool {
	for pos := range all.ins {
		if !all.ins[pos].eod {
			return false
		}
	}
	return true
}

func min(ss ...string) (s string) {
	s = ss[0]
	for _, val := range ss {
		if val < s {
			s = val
		}
	}
	return s
}

// Get reads the next line from all.ins[pos].read and feeds the
// variables all.ins[pos].eof/eod/rec.
// With option -nodup it ignores duplicate lines, and empty ones too.
// If an error occurs get stops working and returns an error value.
func get(pos int, all *allT) error {
	if all.ins[pos].eof { // earlier: EOF + data
		all.ins[pos].eod = true
		all.ins[pos].rec = string(byte(0xFF))
		return nil
	}

readagain:
	rec, err := all.ins[pos].read.ReadString('\n')

	if err != nil && err != io.EOF { // serious read error
		log.Fatalln(all.ins[pos].name+": read error after line:", all.ins[pos].rec, "\n\t", err)
	}

	if err == io.EOF && len(rec) == 0 { // EOF + no data
		all.ins[pos].eof = true
		all.ins[pos].eod = true
		all.ins[pos].rec = string(byte(0xFF))
		return nil
	}

	if err == io.EOF && len(rec) > 0 { // EOF + data
		all.ins[pos].eof = true
	}

	if rec[len(rec)-1] == '\n' {
		rec = rec[:len(rec)-1] // data without delimiter
	}

	all.ins[pos].count += 1

	// -nodup ignores duplicate lines (and empty ones)
	if all.opts.nodup && rec == all.ins[pos].rec {
		goto readagain
	}

	// lines must be ordered ascendingly (compare with previous line)
	if rec < all.ins[pos].rec {
		return errors.New(all.ins[pos].name + ": wrong sequence near:\n" + all.ins[pos].rec)
	}

	all.ins[pos].rec = rec

	return nil
}

func mustOpen(name string) (f *os.File) {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func mustCreate(name string) (f *os.File) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// MustWrite writes rec and ends the line.
func mustWrite(f *bufio.Writer, rec string) {
	_, err := f.WriteString(rec)
	if err != nil {
		log.Fatal(err)
	}
	err = f.WriteByte('\n')
	if err != nil {
		log.Fatal(err)
	}
}
